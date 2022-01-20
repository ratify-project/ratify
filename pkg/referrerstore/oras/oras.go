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
	"encoding/json"
	"fmt"
	"strings"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
	"oras.land/oras-go/pkg/target"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
	"github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"github.com/opencontainers/go-digest"
	artifactspec "github.com/oras-project/artifacts-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
)

const (
	storeName             = "oras"
	defaultLocalCachePath = "~/.ratify/local_oras_cache"
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

type orasStore struct {
	config       *OrasStoreConf
	rawConfig    config.StoreConfig
	localCache   *content.OCI
	authProvider authprovider.AuthProvider
	authCache    map[common.Reference]*content.Registry
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
		conf.LocalCachePath = defaultLocalCachePath
	}
	localRegistry, err := content.NewOCI(conf.LocalCachePath)
	if err != nil {
		return nil, fmt.Errorf("could not create local oras cache at path #{conf.LocalCachePath}: #{err}")
	}

	return &orasStore{config: &conf,
		rawConfig:    config.StoreConfig{Version: version, Store: storeConfig},
		localCache:   localRegistry,
		authProvider: authenticationProvider,
		authCache:    make(map[common.Reference]*content.Registry)}, nil
}

func (store *orasStore) Name() string {
	return storeName
}

func (store *orasStore) GetConfig() *config.StoreConfig {
	return &store.rawConfig
}

func (store *orasStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	// TODO: handle nextToken
	registryClient, err := store.createRegistryClient(subjectReference)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	ref := fmt.Sprintf("%s@%s", subjectReference.Path, subjectReference.Digest)
	var referrerDescriptors []artifactspec.Descriptor
	if artifactTypes == nil {
		artifactTypes = []string{""}
	}
	for _, artifactType := range artifactTypes {
		_, res, err := oras.Discover(ctx, registryClient.Resolver, ref, artifactType)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") && store.config.CosignEnabled {
				logrus.Info("Registry doesn't support oras artifacts, but we can check for cosign artifacts")
			} else {
				return referrerstore.ListReferrersResult{}, err
			}
		}
		referrerDescriptors = append(referrerDescriptors, res...)
	}

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

func (store *orasStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	registryClient, err := store.createRegistryClient(subjectReference)
	if err != nil {
		return nil, err
	}

	ref := fmt.Sprintf("%s@%s", subjectReference.Path, digest)
	desc, err := oras.Copy(ctx, registryClient, ref, store.localCache, "")
	if err != nil {
		return nil, err
	}

	return store.getRawContentFromCache(ctx, desc)
}

func (store *orasStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	client, err := store.createRegistryClient(subjectReference)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	var result ocispecs.ReferenceManifest
	artifactManifestFound := false
	_, err = oras.Graph(ctx, subjectReference.Original, referenceDesc.ArtifactType, client.Resolver,
		func(parent artifactspec.Descriptor, parentManifest artifactspec.Manifest, objects []target.Object) error {
			if parent.Digest == referenceDesc.Digest {
				result = ArtifactManifestToReferenceManifest(parentManifest)
				artifactManifestFound = true
			}
			return nil
		})

	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	if !artifactManifestFound {
		return ocispecs.ReferenceManifest{}, fmt.Errorf("cannot find artifact manifest with digest %s", referenceDesc.Digest)
	}

	return result, nil
}

func (store *orasStore) GetSubjectDescriptor(ctx context.Context, subjectReference common.Reference) (*ocispecs.SubjectDescriptor, error) {
	registryClient, err := store.createRegistryClient(subjectReference)
	if err != nil {
		return nil, err
	}
	_, desc, err := registryClient.Resolve(ctx, subjectReference.Original)
	if err != nil {
		return nil, err
	}
	return &ocispecs.SubjectDescriptor{Descriptor: desc}, nil
}

func (store *orasStore) createRegistryClient(targetRef common.Reference) (*content.Registry, error) {
	if store.authProvider == nil || !store.authProvider.Enabled() {
		return nil, fmt.Errorf("auth provider not properly enabled")
	}

	if registryClient, ok := store.authCache[targetRef]; ok {
		return registryClient, nil
	}

	authConfig, err := store.authProvider.Provide(targetRef.Original)
	if err != nil {
		return nil, err
	}

	registryOpts := content.RegistryOptions{
		Username:  authConfig.Username,
		Password:  authConfig.Password,
		Insecure:  isInsecureRegistry(targetRef.Original, store.config),
		PlainHTTP: store.config.UseHttp,
	}

	registryClient, err := content.NewRegistryWithDiscover(targetRef.Original, registryOpts)
	if err != nil {
		return nil, err
	}

	store.authCache[targetRef] = registryClient

	return content.NewRegistryWithDiscover(targetRef.Original, registryOpts)
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
