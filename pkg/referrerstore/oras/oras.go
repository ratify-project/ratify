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
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	paths "path/filepath"
	"time"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	ocitarget "oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/internal/graph"
	"oras.land/oras-go/v2/internal/status"
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
	"github.com/sirupsen/logrus"
)

const (
	storeName             = "oras"
	defaultLocalCachePath = "local_oras_cache"
	dockerConfigFileName  = "config.json"
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
	// localRegistry, err := content.NewOCI(conf.LocalCachePath)
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

func (store *orasStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string, subjectDesc ...*ocispecs.SubjectDescriptor) (referrerstore.ListReferrersResult, error) {
	// TODO: handle nextToken
	repository, err := store.createRegistryClient(ctx, subjectReference)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	var resolvedSubjectDesc *ocispecs.SubjectDescriptor
	if len(subjectDesc) > 0 {
		resolvedSubjectDesc = subjectDesc[0]
	} else {
		resolvedSubjectDesc, err = store.GetSubjectDescriptor(ctx, subjectReference)
		if err != nil {
			return referrerstore.ListReferrersResult{}, err
		}
	}

	referrerDescriptors, err := repository.UpEdges(ctx, resolvedSubjectDesc.Descriptor)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	var referrers []ocispecs.ReferenceDescriptor
	for _, referrer := range referrerDescriptors {
		referrers = append(referrers, ocispecs.ReferenceDescriptor{Descriptor: referrer})
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

func (store *orasStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	repository, err := store.createRegistryClient(ctx, subjectReference)
	if err != nil {
		return nil, err
	}

	ref := fmt.Sprintf("%s@%s", subjectReference.Path, digest)
	desc, err := oras.Copy(ctx, repository, ref, store.localCache, "")
	if err != nil {
		return nil, err
	}

	return store.getRawContentFromCache(ctx, desc)
}

func (store *orasStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor, subjectDesc ...*ocispecs.SubjectDescriptor) (ocispecs.ReferenceManifest, error) {
	repository, err := store.createRegistryClient(ctx, subjectReference)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	var err error
	var resolvedSubjectDesc *ocispecs.SubjectDescriptor
	if len(subjectDesc) > 0 {
		resolvedSubjectDesc = subjectDesc[0]
	} else {
		resolvedSubjectDesc, err = store.GetSubjectDescriptor(ctx, subjectReference)
		if err != nil {
			return ocispecs.ReferenceManifest{}, err
		}
	}
	tracker := status.NewTracker()
	var result ocispecs.ReferenceManifest
	artifactManifestFound := false
	prehandler := graph.HandlerFunc(func(ctx context.Context, desc oci.Descriptor) ([]oci.Descriptor, error) {
		// skip the descriptor if other go routine is working on it
		done, committed := tracker.TryCommit(desc)
		if !committed {
			return nil, graph.ErrSkipDesc
		}
		if desc.Digest == referenceDesc.Digest {

			desc, err := repository.Resolve(ctx, "")
			if err != nil {
				return []oci.Descriptor{}, err
			}
			result = ArtifactManifestToReferenceManifest(parentManifest)
			artifactManifestFound = true
		}

		// mark the node as done on success
		close(done)
		return []oci.Descriptor{}, nil
	})
	posthandler := graph.Handlers()

	err = graph.Dispatch(ctx, prehandler, posthandler, nil, resolvedSubjectDesc)
	if err != nil {
		return ocispecs.ReferenceManifest{}, nil
	}
	// var result ocispecs.ReferenceManifest
	// artifactManifestFound := false
	// _, err = oras.Graph(ctx, subjectReference.Original, referenceDesc.ArtifactType, client.Resolver,
	// 	func(parent artifactspec.Descriptor, parentManifest artifactspec.Manifest, objects []target.Object) error {
	// 		if parent.Digest == referenceDesc.Digest {
	// 			result = ArtifactManifestToReferenceManifest(parentManifest)
	// 			artifactManifestFound = true
	// 		}
	// 		return nil
	// 	})

	// if err != nil {
	// 	return ocispecs.ReferenceManifest{}, err
	// }

	if !artifactManifestFound {
		return ocispecs.ReferenceManifest{}, fmt.Errorf("cannot find artifact manifest with digest %s", referenceDesc.Digest)
	}

	return result, nil
}

func (store *orasStore) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	repository, err := store.createRegistryClient(ctx, subjectReference)
	if err != nil {
		return nil, err
	}
	desc, err := repository.Resolve(ctx, subjectReference.Original)
	if err != nil {
		return nil, err
	}
	return &ocispecs.SubjectDescriptor{Descriptor: desc}, nil
}

func (store *orasStore) createRegistryClient(ctx context.Context, targetRef common.Reference) (*remote.Repository, error) {
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

	// registryOpts := content.RegistryOptions{
	// 	Username:  authConfig.Username,
	// 	Password:  authConfig.Password,
	// 	Insecure:  isInsecureRegistry(targetRef.Original, store.config),
	// 	PlainHTTP: store.config.UseHttp,
	// }

	// registryClient, err := content.NewRegistryWithDiscover(targetRef.Original, registryOpts)
	// if err != nil {
	// 	return nil, err
	// }

	repository, err := remote.NewRepository(targetRef.Original)
	if err != nil {
		return nil, err
	}

	credentialProvider := func(ctx context.Context, registry string) (auth.Credential, error) {
		if authConfig.Username != "" || authConfig.Password != "" {
			return auth.Credential{
				Username: authConfig.Username,
				Password: authConfig.Password,
			}, nil
		}
		return auth.EmptyCredential, nil
	}

	// Set the Repository Client Credentials
	repoClient := &auth.Client{
		Header: http.Header{
			"User-Agent": {"ratify"},
		},
		Cache:      auth.DefaultCache,
		Credential: credentialProvider,
	}

	if isInsecureRegistry(targetRef.Original, store.config) {
		repoClient.Client = http.DefaultClient
		repoClient.Client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	repository.Client = repoClient
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
