package main

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/plugin/skel"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/sigstore/pkg/signature"
)

type PluginConfig struct {
	Name   string `json:"name"`
	KeyRef string `json:"key"`
	// config specific to the plugin
}

type StoreConfig struct {
	UseHttp      bool   `json:"useHttp,omitempty"`
	AuthProvider string `json:"auth-provider,omitempty"`
}

type PluginInputConfig struct {
	Config      PluginConfig `json:"config"`
	StoreConfig StoreConfig  `json:"storeConfig"`
}

func main() {
	skel.PluginMain("cosign", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginInputConfig, error) {
	conf := PluginInputConfig{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %v", err)
	}

	return &conf, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	payload, err := signatures(context.Background(), subjectReference.Original, input.Config.KeyRef, input)
	if err != nil {
		return &verifier.VerifierResult{
			Name:      input.Config.Name,
			IsSuccess: false,
			Results:   []string{fmt.Sprintf("cosign verification failed with error %v", err)},
		}, nil
	} else if len(payload) > 0 {
		return &verifier.VerifierResult{
			Name:      input.Config.Name,
			IsSuccess: true,
			Results:   []string{"cosign verification success. valid signatures found"},
		}, nil
	}

	return &verifier.VerifierResult{
		Name:      input.Config.Name,
		IsSuccess: false,
		Results:   []string{"cosign verification failed. no valid signatures found"},
	}, nil
}

func signatures(ctx context.Context, img string, keyRef string, config *PluginInputConfig) ([]cosign.SignedPayload, error) {
	var options []name.Option
	if config.StoreConfig.UseHttp {
		options = append(options, name.Insecure)
	}

	ref, err := name.ParseReference(img, options...)
	if err != nil {
		return nil, err
	}

	ecdsaVerifier, err := loadPublicKey(ctx, keyRef)
	if err != nil {
		return nil, err
	}

	if config.StoreConfig.AuthProvider != "" {
		return nil, fmt.Errorf("auth provider %s is not supported", config.StoreConfig.AuthProvider)
	}

	registryClientOptions := []remote.Option{
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	return cosign.Verify(ctx, ref, &cosign.CheckOpts{
		RootCerts:          nil, // TODO: TUF related metadata fulcio.Roots,
		SigVerifier:        ecdsaVerifier,
		ClaimVerifier:      cosign.SimpleClaimVerifier,
		RegistryClientOpts: registryClientOptions,
	})
}

func loadPublicKey(ctx context.Context, keyRef string) (verifier signature.Verifier, err error) {
	raw, err := ioutil.ReadFile(filepath.Clean(keyRef))
	if err != nil {
		return nil, err
	}

	// PEM encoded file.
	ed, err := cosign.PemToECDSAKey(raw)
	if err != nil {
		return nil, errors.Wrap(err, "pem to ecdsa")
	}
	return signature.LoadECDSAVerifier(ed, crypto.SHA256)
}
