package oras

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/referrerstore/config"
	"github.com/deislabs/hora/pkg/referrerstore/factory"
	"github.com/deislabs/hora/plugins/referrerstore/ociregistry/registry"
	"github.com/opencontainers/go-digest"
)

const (
	storeName = "oras"
)

type OrasStoreConf struct {
	Name          string `json:"name"`
	UseHttp       bool   `json:"useHttp,omitempty"`
	CosignEnabled bool   `json:"cosign-enabled,omitempty"`
	AuthProvider  string `json:"auth-provider,omitempty"`
}

type orasStoreFactory struct{}

type orasStore struct {
	config    *OrasStoreConf
	rawConfig config.StoreConfig
}

// Detect the loopback IP (127.0.0.1)
var reLoopback = regexp.MustCompile(regexp.QuoteMeta("127.0.0.1"))

// Detect the loopback IPV6 (::1)
var reipv6Loopback = regexp.MustCompile(regexp.QuoteMeta("::1"))

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

	if conf.AuthProvider != "" {
		return nil, fmt.Errorf("auth provider %s is not supported", conf.AuthProvider)
	}

	return &orasStore{config: &conf, rawConfig: config.StoreConfig{Version: version, Store: storeConfig}}, nil
}

func (store *orasStore) Name() string {
	return storeName
}

func (store *orasStore) GetConfig() *config.StoreConfig {
	return &store.rawConfig
}

func (store *orasStore) ListReferrers(ctx context.Context, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	client, err := store.createRegistryClient(subjectReference.Path)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	referrers, err := client.GetReferrers(subjectReference, artifactTypes, nextToken)

	if err != nil && err != registry.ReferrersNotSupported {
		return referrerstore.ListReferrersResult{}, err
	}

	if store.config.CosignEnabled {
		cosignReferences, err := getCosignReferences(client, subjectReference)
		if err != nil {
			return referrerstore.ListReferrersResult{}, err
		}
		referrers = append(referrers, *cosignReferences...)
	}

	return referrerstore.ListReferrersResult{Referrers: referrers}, nil
}

func (store *orasStore) GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	client, err := store.createRegistryClient(subjectReference.Path)
	if err != nil {
		return nil, err
	}

	blob, _, err := client.GetReferenceBlob(subjectReference, digest)

	if err != nil {
		return nil, err
	}

	return blob, nil

}

func (store *orasStore) GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error) {
	client, err := store.createRegistryClient(subjectReference.Path)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	subjectReference.Digest = referenceDesc.Digest
	return client.GetReferenceManifest(subjectReference)
}

func (store *orasStore) createRegistryClient(path string) (*registry.Client, error) {
	registryStr, _ := registry.GetRegistryRepoString(path)
	authConfig, err := registry.DefaultAuthProvider.Provide(registryStr)
	if err != nil {
		return nil, err
	}

	return registry.NewClient(
		registry.NewAuthtransport(
			nil,
			authConfig.Username,
			authConfig.Password,
		),
		isInsecureRegistry(registryStr, store.config),
	), nil
}

func isInsecureRegistry(registry string, config *OrasStoreConf) bool {
	if config.UseHttp {
		return true
	}
	if strings.HasPrefix(registry, "localhost:") {
		return true
	}

	if reLoopback.MatchString(registry) {
		return true
	}
	if reipv6Loopback.MatchString(registry) {
		return true
	}

	return false
}
