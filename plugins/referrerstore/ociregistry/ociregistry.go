package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/referrerstore/plugin/skel"
	"github.com/deislabs/hora/plugins/referrerstore/ociregistry/registry"
	"github.com/opencontainers/go-digest"
)

// Detect the loopback IP (127.0.0.1)
var reLoopback = regexp.MustCompile(regexp.QuoteMeta("127.0.0.1"))

// Detect the loopback IPV6 (::1)
var reipv6Loopback = regexp.MustCompile(regexp.QuoteMeta("::1"))

type PluginConf struct {
	Name          string `json:"name"`
	UseHttp       bool   `json:"useHttp,omitempty"`
	CosignEnabled bool   `json:"cosign-enabled,omitempty"`
	AuthProvider  string `json:"auth-provider,omitempty"`
}

func main() {
	skel.PluginMain("ociregistry", "1.0.0", ListReferrers, GetBlobContent, GetReferenceManifest, []string{"1.0.0"})
}

func parseConfig(stdin []byte) (*PluginConf, error) {
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse ociregistry configuration: %v", err)
	}

	return &conf, nil
}

func ListReferrers(args *skel.CmdArgs, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	client, config, err := createRegistryClient(args, subjectReference.Path)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	referrers, err := client.GetReferrers(subjectReference, artifactTypes, nextToken)

	if err != nil && err != registry.ReferrersNotSupported {
		return referrerstore.ListReferrersResult{}, err
	}

	if config.CosignEnabled {
		cosignReferences, err := getCosignReferences(client, subjectReference)
		if err != nil {
			return referrerstore.ListReferrersResult{}, err
		}
		referrers = append(referrers, *cosignReferences...)
	}

	return referrerstore.ListReferrersResult{Referrers: referrers}, nil
}

func GetBlobContent(args *skel.CmdArgs, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	client, _, err := createRegistryClient(args, subjectReference.Path)
	if err != nil {
		return nil, err
	}

	blob, _, err := client.GetReferenceBlob(subjectReference, digest)

	if err != nil {
		return nil, err
	}

	return blob, nil

}

func GetReferenceManifest(args *skel.CmdArgs, subjectReference common.Reference, refdigest digest.Digest) (ocispecs.ReferenceManifest, error) {
	client, _, err := createRegistryClient(args, subjectReference.Path)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	subjectReference.Digest = refdigest
	return client.GetReferenceManifest(subjectReference)
}

func createRegistryClient(args *skel.CmdArgs, path string) (*registry.Client, *PluginConf, error) {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return nil, nil, err
	}

	if conf.AuthProvider != "" {
		return nil, nil, fmt.Errorf("auth provider %s is not supported", conf.AuthProvider)
	}

	registryStr, _ := registry.GetRegistryRepoString(path)
	authConfig, err := registry.DefaultAuthProvider.Provide(registryStr)
	if err != nil {
		return nil, nil, err
	}

	return registry.NewClient(
		registry.NewAuthtransport(
			nil,
			authConfig.Username,
			authConfig.Password,
		),
		isInsecureRegistry(registryStr, conf),
	), conf, nil
}

func isInsecureRegistry(registry string, config *PluginConf) bool {
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
