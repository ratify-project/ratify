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
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/pkg/cosign"
)

const (
	CosignArtifactType = "org.sigstore.cosign.v1"
	regRepoDelimiter   = "/"
	dDefaultRegistry   = "index.docker.io"
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

	registryStr := getRegistryString(path)
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

func getCosignReferences(client *registry.Client, subjectReference common.Reference) (*[]ocispecs.ReferenceDescriptor, error) {
	var references []ocispecs.ReferenceDescriptor
	ref, err := name.ParseReference(subjectReference.Original)
	if err != nil {
		return nil, err
	}
	hash := v1.Hash{
		Algorithm: subjectReference.Digest.Algorithm().String(),
		Hex:       subjectReference.Digest.Hex(),
	}
	signatureTag := cosign.AttachedImageTag(ref.Context(), hash, cosign.SignatureTagSuffix)
	tagRef := common.Reference{
		Path: subjectReference.Path,
		Tag:  signatureTag.TagStr(),
	}
	desc, err := client.GetManifestMetadata(tagRef)

	if err != nil && err != registry.ManifestNotFound {
		return nil, err
	}

	if err == nil {
		references = append(references, ocispecs.ReferenceDescriptor{
			ArtifactType: CosignArtifactType,
			Descriptor: oci.Descriptor{
				MediaType: desc.MediaType,
				Digest:    desc.Digest,
				Size:      desc.Size,
			},
		})
	}

	return &references, nil
}

func getRegistryString(path string) string {
	var registry string
	parts := strings.SplitN(path, regRepoDelimiter, 2)
	if len(parts) == 2 && (strings.ContainsRune(parts[0], '.') || strings.ContainsRune(parts[0], ':')) {
		// The first part of the repository is treated as the registry domain
		// iff it contains a '.' or ':' character, otherwise it is all repository
		// and the domain defaults to Docker Hub.
		registry = parts[0]
	}

	if registry == "" {
		return dDefaultRegistry
	}

	return registry
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
