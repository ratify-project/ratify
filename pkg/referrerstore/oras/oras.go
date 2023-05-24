/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package oras

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	paths "path/filepath"
	"sync"
	"time"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	ocitarget "oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/errcode"
	"oras.land/oras-go/v2/registry/remote/retry"

	ratifyconfig "github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/common/oras/authprovider"
	_ "github.com/deislabs/ratify/pkg/common/oras/authprovider/aws"
	_ "github.com/deislabs/ratify/pkg/common/oras/authprovider/azure"
	commonutils "github.com/deislabs/ratify/pkg/common/utils"
	"github.com/deislabs/ratify/pkg/homedir"
	"github.com/deislabs/ratify/pkg/metrics"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
)

const (
	HttpMaxIdleConns                      = 100
	HttpMaxConnsPerHost                   = 100
	HttpMaxIdleConnsPerHost               = 100
	HttpRetryMax                          = 5
	HttpRetryDurationMin    time.Duration = 200 * time.Millisecond
	HttpRetryDurationMax    time.Duration = 1750 * time.Millisecond
)

const (
	storeName             = "oras"
	defaultLocalCachePath = "local_oras_cache"
	dockerConfigFileName  = "config.json"
	ratifyUserAgent       = "ratify"
)

// OrasStoreConf describes the configuration of ORAS store
type OrasStoreConf struct {
	Name           string                          `json:"name"`
	UseHttp        bool                            `json:"useHttp,omitempty"`
	CosignEnabled  bool                            `json:"cosignEnabled,omitempty"`
	AuthProvider   authprovider.AuthProviderConfig `json:"authProvider,omitempty"`
	LocalCachePath string                          `json:"localCachePath,omitempty"`
}

type orasStoreFactory struct{}

type authCacheEntry struct {
	client    registry.Repository
	expiresOn time.Time
}

type orasStore struct {
	config                 *OrasStoreConf
	rawConfig              config.StoreConfig
	localCache             content.Storage
	authProvider           authprovider.AuthProvider
	authCache              sync.Map
	subjectDescriptorCache sync.Map
	httpClient             *http.Client
	httpClientInsecure     *http.Client
	createRepository       func(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error)
}

func init() {
	factory.Register(storeName, &orasStoreFactory{})
}

func (s *orasStoreFactory) Create(version string, storeConfig config.StorePluginConfig) (referrerstore.ReferrerStore, error) {
	storeBase, err := createBaseStore(version, storeConfig)
	if err != nil {
		return nil, err
	}

	cacheConf, err := toCacheConfig(storeBase.GetConfig().Store)
	if err != nil {
		return nil, err
	}
	if !cacheConf.Enabled {
		return storeBase, nil
	}

	return createCachedStore(storeBase, cacheConf)
}

func createBaseStore(version string, storeConfig config.StorePluginConfig) (*orasStore, error) {
	conf := OrasStoreConf{}

	storeConfigBytes, err := json.Marshal(storeConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(storeConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse oras store configuration: %w", err)
	}

	authenticationProvider, err := authprovider.CreateAuthProviderFromConfig(conf.AuthProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider from configuration: %w", err)
	}

	// Set up the local cache where content will land when we pull
	if conf.LocalCachePath == "" {
		conf.LocalCachePath = paths.Join(homedir.Get(), ratifyconfig.ConfigFileDir, defaultLocalCachePath)
	}

	localRegistry, err := ocitarget.New(conf.LocalCachePath)
	if err != nil {
		return nil, fmt.Errorf("could not create local oras cache at path %s: %w", conf.LocalCachePath, err)
	}

	var customPredicate retry.Predicate = func(resp *http.Response, err error) (bool, error) {
		host := ""
		if resp != nil {
			if resp.Request != nil && resp.Request.URL != nil {
				host = resp.Request.URL.Host
			}
			metrics.ReportRegistryRequestCount(resp.Request.Context(), resp.StatusCode, host)
		}
		return retry.DefaultPredicate(resp, err)
	}

	customRetryPolicy := func() retry.Policy {
		return &retry.GenericPolicy{
			Retryable: customPredicate,
			Backoff:   retry.DefaultBackoff,
			MinWait:   HttpRetryDurationMin,
			MaxWait:   HttpRetryDurationMax,
			MaxRetry:  HttpRetryMax,
		}
	}

	// define the http client for TLS enabled
	secureTransport := http.DefaultTransport.(*http.Transport).Clone()
	secureTransport.MaxIdleConns = HttpMaxIdleConns
	secureTransport.MaxConnsPerHost = HttpMaxConnsPerHost
	secureTransport.MaxIdleConnsPerHost = HttpMaxIdleConnsPerHost
	secureRetryTransport := retry.NewTransport(secureTransport)
	secureRetryTransport.Policy = customRetryPolicy

	// define the http client for TLS disabled
	insecureTransport := http.DefaultTransport.(*http.Transport).Clone()
	insecureTransport.MaxIdleConns = HttpMaxIdleConns
	insecureTransport.MaxConnsPerHost = HttpMaxConnsPerHost
	insecureTransport.MaxIdleConnsPerHost = HttpMaxIdleConnsPerHost
	// #nosec G402
	insecureTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	insecureRetryTransport := retry.NewTransport(insecureTransport)
	insecureRetryTransport.Policy = customRetryPolicy

	return &orasStore{config: &conf,
		rawConfig:          config.StoreConfig{Version: version, Store: storeConfig},
		localCache:         localRegistry,
		authProvider:       authenticationProvider,
		httpClient:         &http.Client{Transport: secureRetryTransport},
		httpClientInsecure: &http.Client{Transport: insecureRetryTransport},
		createRepository:   createDefaultRepository}, nil
}

func (store *orasStore) Name() string {
	return storeName
}

func (store *orasStore) GetConfig() *config.StoreConfig {
	return &store.rawConfig
}

func (store *orasStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	repository, expiry, err := store.createRepository(ctx, store, subjectReference)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	// resolve subject descriptor if not provided
	var resolvedSubjectDesc *ocispecs.SubjectDescriptor
	if subjectDesc != nil {
		resolvedSubjectDesc = subjectDesc
	} else {
		if resolvedSubjectDesc, err = store.GetSubjectDescriptor(ctx, subjectReference); err != nil {
			var ec errcode.Error
			if errors.As(err, &ec) && (ec.Code == fmt.Sprint(http.StatusForbidden) || ec.Code == fmt.Sprint(http.StatusUnauthorized)) {
				store.evictAuthCache(subjectReference.Original, err)
			}
			return referrerstore.ListReferrersResult{}, err
		}
	}

	// find all referrers referencing subject descriptor
	artifactTypeFilter := ""
	var referrerDescriptors []oci.Descriptor
	if err := repository.Referrers(ctx, resolvedSubjectDesc.Descriptor, artifactTypeFilter, func(referrers []oci.Descriptor) error {
		referrerDescriptors = append(referrerDescriptors, referrers...)
		return nil
	}); err != nil && !errors.Is(err, errdef.ErrNotFound) {
		var ec errcode.Error
		if errors.As(err, &ec) && (ec.Code == fmt.Sprint(http.StatusForbidden) || ec.Code == fmt.Sprint(http.StatusUnauthorized)) {
			store.evictAuthCache(subjectReference.Original, err)
		}
		return referrerstore.ListReferrersResult{}, err
	}
	// add the repository client to the auth cache if all repository operations successful
	store.addAuthCache(subjectReference.Original, repository, expiry)

	// convert artifact descriptors to oci descriptor with artifact type
	referrers := []ocispecs.ReferenceDescriptor{}
	for _, referrer := range referrerDescriptors {
		referrers = append(referrers, OciDescriptorToReferenceDescriptor(referrer))
	}

	return referrerstore.ListReferrersResult{Referrers: referrers}, nil
}

func (store *orasStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	var err error
	repository, expiry, err := store.createRepository(ctx, store, subjectReference)
	if err != nil {
		return nil, err
	}

	// create a dummy Descriptor to check the local store cache
	blobDescriptor := oci.Descriptor{
		Digest: digest,
		Size:   0, // dummy size value
	}

	// check if blob exists in local ORAS cache
	isCached, err := store.localCache.Exists(ctx, blobDescriptor)
	if err != nil {
		return nil, err
	}
	metrics.ReportBlobCacheCount(ctx, isCached)

	if !isCached {
		// generate the reference path with digest
		ref := fmt.Sprintf("%s@%s", subjectReference.Path, digest)

		// fetch blob content from remote repository
		blobDesc, rc, err := repository.Blobs().FetchReference(ctx, ref)
		if err != nil {
			var ec errcode.Error
			if errors.As(err, &ec) && (ec.Code == fmt.Sprint(http.StatusForbidden) || ec.Code == fmt.Sprint(http.StatusUnauthorized)) {
				store.evictAuthCache(subjectReference.Original, err)
			}
			return nil, err
		}

		// push fetched content to local ORAS cache
		orasExistsExpectedError := fmt.Errorf("%s: %s: %w", blobDesc.Digest, blobDesc.MediaType, errdef.ErrAlreadyExists)
		err = store.localCache.Push(ctx, blobDesc, rc)
		if err != nil && err.Error() != orasExistsExpectedError.Error() {
			return nil, err
		}
	}

	// add the repository client to the auth cache if all repository operations successful
	store.addAuthCache(subjectReference.Original, repository, expiry)

	return store.getRawContentFromCache(ctx, blobDescriptor)
}

func (store *orasStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	repository, expiry, err := store.createRepository(ctx, store, subjectReference)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}
	var manifestBytes []byte
	// check if manifest exists in local ORAS cache
	isCached, err := store.localCache.Exists(ctx, referenceDesc.Descriptor)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}
	metrics.ReportBlobCacheCount(ctx, isCached)

	if !isCached {
		// fetch manifest content from repository
		manifestReader, err := repository.Fetch(ctx, referenceDesc.Descriptor)
		if err != nil {
			var ec errcode.Error
			if errors.As(err, &ec) && (ec.Code == fmt.Sprint(http.StatusForbidden) || ec.Code == fmt.Sprint(http.StatusUnauthorized)) {
				store.evictAuthCache(subjectReference.Original, err)
			}
			return ocispecs.ReferenceManifest{}, err
		}

		manifestBytes, err = io.ReadAll(manifestReader)
		if err != nil {
			return ocispecs.ReferenceManifest{}, err
		}

		// push fetched manifest to local ORAS cache
		orasExistsExpectedError := fmt.Errorf("%s: %s: %w", referenceDesc.Descriptor.Digest, referenceDesc.Descriptor.MediaType, errdef.ErrAlreadyExists)
		err = store.localCache.Push(ctx, referenceDesc.Descriptor, bytes.NewReader(manifestBytes))
		if err != nil && err.Error() != orasExistsExpectedError.Error() {
			return ocispecs.ReferenceManifest{}, err
		}

		// add the repository client to the auth cache if all repository operations successful
		store.addAuthCache(subjectReference.Original, repository, expiry)
	} else {
		manifestBytes, err = store.getRawContentFromCache(ctx, referenceDesc.Descriptor)
		if err != nil {
			return ocispecs.ReferenceManifest{}, err
		}
	}

	referenceManifest := ocispecs.ReferenceManifest{}

	// marshal manifest bytes into reference manifest descriptor
	if referenceDesc.Descriptor.MediaType == oci.MediaTypeImageManifest {
		var imageManifest oci.Manifest
		if err := json.Unmarshal(manifestBytes, &imageManifest); err != nil {
			return ocispecs.ReferenceManifest{}, err
		}
		referenceManifest = commonutils.OciManifestToReferenceManifest(imageManifest)
	} else if referenceDesc.Descriptor.MediaType == oci.MediaTypeArtifactManifest {
		if err := json.Unmarshal(manifestBytes, &referenceManifest); err != nil {
			return ocispecs.ReferenceManifest{}, err
		}
	} else {
		return ocispecs.ReferenceManifest{}, fmt.Errorf("unsupported manifest media type: %s", referenceDesc.Descriptor.MediaType)
	}

	return referenceManifest, nil
}

func (store *orasStore) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	var desc oci.Descriptor
	if cachedDesc, ok := store.subjectDescriptorCache.Load(subjectReference.Digest); ok && subjectReference.Digest != "" {
		desc = cachedDesc.(oci.Descriptor)
		return &ocispecs.SubjectDescriptor{Descriptor: desc}, nil
	}

	logrus.Debugf("no digest provided for reference %s. attempting to resolve...", subjectReference.Original)
	repository, expiry, err := store.createRepository(ctx, store, subjectReference)
	if err != nil {
		return nil, err
	}

	desc, err = repository.Resolve(ctx, subjectReference.Original)
	if err != nil {
		var ec errcode.Error
		if errors.As(err, &ec) && (ec.Code == fmt.Sprint(http.StatusForbidden) || ec.Code == fmt.Sprint(http.StatusUnauthorized)) {
			store.evictAuthCache(subjectReference.Original, err)
		}
		return nil, err
	}
	// add the subject descriptor to cache
	store.subjectDescriptorCache.Store(desc.Digest, desc)
	// add the repository client to the auth cache if all repository operations successful
	store.addAuthCache(subjectReference.Original, repository, expiry)

	return &ocispecs.SubjectDescriptor{Descriptor: desc}, nil
}

func createDefaultRepository(ctx context.Context, store *orasStore, targetRef common.Reference) (registry.Repository, time.Time, error) {
	if store.authProvider == nil || !store.authProvider.Enabled(ctx) {
		return nil, time.Now(), fmt.Errorf("auth provider not properly enabled")
	}

	if entry, ok := store.authCache.Load(targetRef.Original); ok {
		// if the auth cache entry expiration has not expired or it was never set
		cacheEntry := entry.(authCacheEntry)
		if cacheEntry.expiresOn.IsZero() || cacheEntry.expiresOn.After(time.Now()) {
			return cacheEntry.client, cacheEntry.expiresOn, nil
		}
	}

	authConfig, err := store.authProvider.Provide(ctx, targetRef.Original)
	if err != nil {
		logrus.Warningf("auth provider failed with err, %v", err)
		logrus.Info("attempting to use anonymous credentials")
	}

	// create new ORAS repository target to the image/repository reference
	repository, err := remote.NewRepository(targetRef.Original)
	if err != nil {
		return nil, time.Now(), err
	}

	// set the provider to return the resolved credentials
	credentialProvider := func(ctx context.Context, registry string) (auth.Credential, error) {
		if authConfig.Username != "" || authConfig.Password != "" || authConfig.IdentityToken != "" {
			return auth.Credential{
				Username:     authConfig.Username,
				Password:     authConfig.Password,
				RefreshToken: authConfig.IdentityToken,
			}, nil
		}
		return auth.EmptyCredential, nil
	}

	// set the repository client credentials
	repoClient := &auth.Client{
		Client: store.httpClient,
		Header: http.Header{
			"User-Agent": {ratifyUserAgent},
		},
		Cache:      auth.NewCache(),
		Credential: credentialProvider,
	}

	// enable insecure if specified in config
	if isInsecureRegistry(targetRef.Original, store.config) {
		repoClient.Client = store.httpClientInsecure
	}

	repository.Client = repoClient
	// enable plain HTTP if specified in config
	repository.PlainHTTP = store.config.UseHttp

	return repository, authConfig.ExpiresOn, nil
}

func (store *orasStore) getRawContentFromCache(ctx context.Context, descriptor oci.Descriptor) ([]byte, error) {
	reader, err := store.localCache.Fetch(ctx, descriptor)
	if err != nil {
		return nil, err
	}

	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (store *orasStore) addAuthCache(ref string, repository registry.Repository, expiry time.Time) {
	store.authCache.LoadOrStore(ref, authCacheEntry{
		client:    repository,
		expiresOn: expiry,
	})
}

func (store *orasStore) evictAuthCache(ref string, err error) {
	store.authCache.Delete(ref)
}
