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
	"fmt"
	"io"
	"net/http"
	paths "path/filepath"
	"strings"
	"time"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
	ocitarget "oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	ratifyconfig "github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/homedir"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider/azure"
	"github.com/opencontainers/go-digest"
	artifactspec "github.com/oras-project/artifacts-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
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
	CosignEnabled  bool                            `json:"cosign-enabled,omitempty"`
	AuthProvider   authprovider.AuthProviderConfig `json:"auth-provider,omitempty"`
	LocalCachePath string                          `json:"localCachePath,omitempty"`
}

type orasStoreFactory struct{}

type authCacheEntry struct {
	client    *remote.Repository
	expiresOn time.Time
}

type orasStore struct {
	config       *OrasStoreConf
	rawConfig    config.StoreConfig
	localCache   *ocitarget.Store
	authProvider authprovider.AuthProvider
	authCache    map[string]authCacheEntry
}

func init() {
	factory.Register(storeName, &orasStoreFactory{})
}

func (s *orasStoreFactory) Create(version string, storeConfig config.StorePluginConfig) (referrerstore.ReferrerStore, error) {
	conf := OrasStoreConf{}

	storeConfigBytes, err := json.Marshal(storeConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(storeConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse oras store configuration: %v", err)
	}

	authenticationProvider, err := authprovider.CreateAuthProviderFromConfig(conf.AuthProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider from configuration: %v", err)
	}

	// Set up the local cache where content will land when we pull
	if conf.LocalCachePath == "" {
		conf.LocalCachePath = paths.Join(homedir.Get(), ratifyconfig.ConfigFileDir, defaultLocalCachePath)
	}

	localRegistry, err := ocitarget.New(conf.LocalCachePath)
	if err != nil {
		return nil, fmt.Errorf("could not create local oras cache at path %s: %s", conf.LocalCachePath, err)
	}

	return &orasStore{config: &conf,
		rawConfig:    config.StoreConfig{Version: version, Store: storeConfig},
		localCache:   localRegistry,
		authProvider: authenticationProvider,
		authCache:    make(map[string]authCacheEntry)}, nil
}

func (store *orasStore) Name() string {
	return storeName
}

func (store *orasStore) GetConfig() *config.StoreConfig {
	return &store.rawConfig
}

func (store *orasStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc *ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	repository, err := store.createRepository(ctx, subjectReference)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	// resolve subject descriptor if not provided
	var resolvedSubjectDesc *ocispecs.SubjectDescriptor
	if subjectDesc != nil {
		resolvedSubjectDesc = subjectDesc
	} else {
		resolvedSubjectDesc, err = store.GetSubjectDescriptor(ctx, subjectReference)
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				delete(store.authCache, subjectReference.Original)
			}
			return referrerstore.ListReferrersResult{}, err
		}
	}

	// find all referrers referencing subject descriptor
	var referrerDescriptors []artifactspec.Descriptor
	if err := repository.Referrers(ctx, resolvedSubjectDesc.Descriptor, func(referrers []artifactspec.Descriptor) error {
		referrerDescriptors = referrers
		return nil
	}); err != nil {
		if strings.Contains(err.Error(), "401") {
			delete(store.authCache, subjectReference.Original)
		}
		return referrerstore.ListReferrersResult{}, err
	}

	// convert artifact descriptors to oci descriptor with artifact type
	var referrers []ocispecs.ReferenceDescriptor
	for _, referrer := range referrerDescriptors {
		referrers = append(referrers, ArtifactDescriptorToReferenceDescriptor(referrer))
	}

	if store.config.CosignEnabled {
		cosignReferences, err := getCosignReferences(subjectReference)
		if err != nil {
			return referrerstore.ListReferrersResult{}, err
		}
		referrers = append(referrers, *cosignReferences...)
	}

	return referrerstore.ListReferrersResult{Referrers: referrers}, nil
}

func (store *orasStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest, blobDesc oci.Descriptor) ([]byte, error) {
	var err error
	repository, err := store.createRepository(ctx, subjectReference)
	if err != nil {
		return nil, err
	}

	// resolve blob descriptor if not provided
	var resolvedBlobDesc oci.Descriptor
	if blobDesc.Digest != "" {
		resolvedBlobDesc = blobDesc
	} else {
		ref := fmt.Sprintf("%s@%s", subjectReference.Path, digest)
		resolvedBlobDesc, err = repository.Blobs().Resolve(ctx, ref)
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				delete(store.authCache, subjectReference.Original)
			}
			return nil, err
		}
	}

	// check if blob exists in local ORAS cache
	isCached, err := store.localCache.Exists(ctx, resolvedBlobDesc)
	if err != nil {
		return nil, err
	}

	if !isCached {
		// fetch blob content from remote repository
		rc, err := repository.Fetch(ctx, resolvedBlobDesc)
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				delete(store.authCache, subjectReference.Original)
			}
			return nil, err
		}

		// push fetched content to local ORAS cache
		err = store.localCache.Push(ctx, resolvedBlobDesc, rc)
		if err != nil {
			return nil, err
		}
	}

	return store.getRawContentFromCache(ctx, resolvedBlobDesc)
}

func (store *orasStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	repository, err := store.createRepository(ctx, subjectReference)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}
	var manifestBytes []byte
	// check if manifest exists in local ORAS cache
	isCached, err := store.localCache.Exists(ctx, referenceDesc.Descriptor)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	if !isCached {
		// fetch manifest content from repository
		manifestReader, err := repository.Fetch(ctx, referenceDesc.Descriptor)
		if err != nil {
			if strings.Contains(err.Error(), "401") {
				delete(store.authCache, subjectReference.Original)
			}
			return ocispecs.ReferenceManifest{}, err
		}

		manifestBytes, err = io.ReadAll(manifestReader)
		if err != nil {
			return ocispecs.ReferenceManifest{}, err
		}

		// push fetched manifest to local ORAS cache
		err = store.localCache.Push(ctx, referenceDesc.Descriptor, bytes.NewReader(manifestBytes))
		if err != nil {
			return ocispecs.ReferenceManifest{}, err
		}
	} else {
		manifestBytes, err = store.getRawContentFromCache(ctx, referenceDesc.Descriptor)
		if err != nil {
			return ocispecs.ReferenceManifest{}, err
		}
	}

	// marshal manifest bytes into reference manifest descriptor
	referenceManifest := ocispecs.ReferenceManifest{}
	if err := json.Unmarshal(manifestBytes, &referenceManifest); err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	return referenceManifest, nil
}

func (store *orasStore) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	repository, err := store.createRepository(ctx, subjectReference)
	if err != nil {
		return nil, err
	}

	desc, err := repository.Resolve(ctx, subjectReference.Original)
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			delete(store.authCache, subjectReference.Original)
		}
		return nil, err
	}

	return &ocispecs.SubjectDescriptor{Descriptor: desc}, nil
}

func (store *orasStore) createRepository(ctx context.Context, targetRef common.Reference) (*remote.Repository, error) {
	if store.authProvider == nil || !store.authProvider.Enabled(ctx) {
		return nil, fmt.Errorf("auth provider not properly enabled")
	}

	if cacheEntry, ok := store.authCache[targetRef.Original]; ok {
		// if the auth cache entry expiration has not expired or it was never set
		if cacheEntry.expiresOn.IsZero() || cacheEntry.expiresOn.After(time.Now()) {
			return cacheEntry.client, nil
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
		return nil, err
	}

	// set the provider to return the resolved credentials
	credentialProvider := func(ctx context.Context, registry string) (auth.Credential, error) {
		if authConfig.Username != "" || authConfig.Password != "" {
			return auth.Credential{
				Username: authConfig.Username,
				Password: authConfig.Password,
			}, nil
		}
		return auth.EmptyCredential, nil
	}

	// set the repository client credentials
	repoClient := &auth.Client{
		Header: http.Header{
			"User-Agent": {ratifyUserAgent},
		},
		Cache:      auth.DefaultCache,
		Credential: credentialProvider,
	}

	// enable insecure if specified in config
	if isInsecureRegistry(targetRef.Original, store.config) {
		repoClient.Client = http.DefaultClient
		repoClient.Client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	repository.Client = repoClient
	// enable plain HTTP if specified in config
	repository.PlainHTTP = store.config.UseHttp

	store.authCache[targetRef.Original] = authCacheEntry{
		client:    repository,
		expiresOn: authConfig.ExpiresOn,
	}

	return repository, nil
}

func (store *orasStore) getRawContentFromCache(ctx context.Context, descriptor oci.Descriptor) ([]byte, error) {
	reader, err := store.localCache.Fetch(ctx, descriptor)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, descriptor.Size)
	_, err = reader.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
