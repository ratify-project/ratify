package main

import (
	"encoding/json"
	"fmt"

	"github.com/notaryproject/hora/pkg/common"
	"github.com/notaryproject/hora/pkg/ocispecs"
	"github.com/notaryproject/hora/pkg/referrerstore"
	"github.com/notaryproject/hora/pkg/referrerstore/plugin/skel"
	"github.com/notaryproject/hora/plugins/referrerstore/ociregistry/registry"
	"github.com/opencontainers/go-digest"
)

type PluginConf struct {
	Name    string `json:"name"`
	UseHttp bool   `json:"useHttp,omitempty"`
	// TODO Credential provider
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func main() {
	skel.PluginMain("ociregistry", "1.0.0", ListReferrers, GetBlobContent, GetReferenceManifest, []string{"1.0.0"})
}

func parseConfig(stdin []byte) (PluginConf, error) {
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return PluginConf{}, fmt.Errorf("failed to parse ociregistry configuration: %v", err)
	}

	return conf, nil
}

func ListReferrers(args *skel.CmdArgs, subjectReference common.Reference, artifactTypes []string, nextToken string) (referrerstore.ListReferrersResult, error) {
	client, err := createRegistryClient(args)
	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	referrers, err := client.GetReferrers(subjectReference, artifactTypes, nextToken)

	if err != nil {
		return referrerstore.ListReferrersResult{}, err
	}

	return referrerstore.ListReferrersResult{Referrers: referrers}, nil
}

func GetBlobContent(args *skel.CmdArgs, subjectReference common.Reference, digest digest.Digest) ([]byte, error) {
	client, err := createRegistryClient(args)
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
	client, err := createRegistryClient(args)
	if err != nil {
		return ocispecs.ReferenceManifest{}, err
	}

	subjectReference.Digest = refdigest
	return client.GetReferenceManifest(subjectReference)
}

func createRegistryClient(args *skel.CmdArgs) (*registry.Client, error) {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return nil, err
	}

	return registry.NewClient(
		registry.NewAuthtransport(
			nil,
			conf.Username,
			conf.Password,
		),
		conf.UseHttp,
	), nil
}
