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

package main

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/utils"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/plugin/skel"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/sigstore/cosign/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/cosign/pkg/oci"
	ociremote "github.com/sigstore/cosign/pkg/oci/remote"
	"github.com/sigstore/sigstore/pkg/signature"
)

type PluginConfig struct {
	Name     string `json:"name"`
	KeyRef   string `json:"key"`
	RekorURL string `json:"rekorURL"`
	// config specific to the plugin
}

type StoreConfig struct {
	UseHttp      bool   `json:"useHttp,omitempty"`
	AuthProvider string `json:"authProvider,omitempty"`
}

type StoreWrapperConfig struct {
	StoreConfig StoreConfig `json:"store"`
}

type PluginInputConfig struct {
	Config             PluginConfig       `json:"config"`
	StoreWrapperConfig StoreWrapperConfig `json:"storeConfig"`
}

func main() {
	skel.PluginMain("cosign", "1.1.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginInputConfig, error) {
	conf := PluginInputConfig{}
	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %w", err)
	}

	return &conf, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	payload, _, err := signatures(context.Background(), subjectReference.Original, input.Config.KeyRef, input)
	if err != nil {
		return &verifier.VerifierResult{
			Name:      input.Config.Name,
			IsSuccess: false,
			Message:   fmt.Sprintf("cosign verification failed with error %v", err),
		}, nil
	} else if len(payload) > 0 {
		return &verifier.VerifierResult{
			Name:      input.Config.Name,
			IsSuccess: true,
			Message:   "cosign verification success. valid signatures found",
		}, nil
	}

	return &verifier.VerifierResult{
		Name:      input.Config.Name,
		IsSuccess: false,
		Message:   "cosign verification failed. no valid signatures found",
	}, nil
}

func signatures(ctx context.Context, img string, keyRef string, config *PluginInputConfig) (checkedSignatures []oci.Signature, bundleVerified bool, err error) {
	registryClientOptions := []remote.Option{
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}
	var options []name.Option
	if config.StoreWrapperConfig.StoreConfig.UseHttp {
		options = append(options, name.Insecure)
		// #nosec G402
		registryClientOptions = append(registryClientOptions, remote.WithTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}))
	}

	ref, err := name.ParseReference(img, options...)
	if err != nil {
		return nil, false, err
	}

	registryClientOptionsWrapper := []ociremote.Option{
		ociremote.WithRemoteOptions(registryClientOptions...),
	}
	cosignOpts := &cosign.CheckOpts{
		ClaimVerifier:      cosign.SimpleClaimVerifier,
		RegistryClientOpts: registryClientOptionsWrapper,
	}

	var ecdsaVerifier signature.Verifier
	var roots *x509.CertPool
	if keyRef != "" {
		ecdsaVerifier, err = loadPublicKey(ctx, keyRef)
		if err != nil {
			return nil, false, err
		}
		cosignOpts.SigVerifier = ecdsaVerifier
	} else {
		roots, err = fulcio.GetRoots()
		if err != nil {
			return nil, false, err
		}
		cosignOpts.RootCerts = roots
		if cosignOpts.RootCerts == nil {
			return nil, false, fmt.Errorf("failed to initialize root certificates")
		}
	}

	if config.Config.RekorURL != "" {
		cosignOpts.RekorClient, err = rekor.NewClient(config.Config.RekorURL)
		if err != nil {
			return nil, false, fmt.Errorf("failed to create Rekor client from URL %s: %w", config.Config.RekorURL, err)
		}
	}

	if config.StoreWrapperConfig.StoreConfig.AuthProvider != "" {
		return nil, false, fmt.Errorf("auth provider %s is not supported", config.StoreWrapperConfig.StoreConfig.AuthProvider)
	}

	return cosign.VerifyImageSignatures(ctx, ref, cosignOpts)
}

func loadPublicKey(ctx context.Context, keyRef string) (verifier signature.Verifier, err error) {
	keyPath := filepath.Clean(utils.ReplaceHomeShortcut(keyRef))
	raw, err := os.ReadFile(keyPath)
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
