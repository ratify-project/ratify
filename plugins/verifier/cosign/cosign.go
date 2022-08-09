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
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/cosign/pkg/oci"
	ociremote "github.com/sigstore/cosign/pkg/oci/remote"
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
	skel.PluginMain("cosign", "1.1.0", VerifyReference, []string{"1.0.0"})
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
	var options []name.Option
	if config.StoreConfig.UseHttp {
		options = append(options, name.Insecure)
	}

	ref, err := name.ParseReference(img, options...)
	if err != nil {
		return nil, false, err
	}

	ecdsaVerifier, err := loadPublicKey(ctx, keyRef)
	if err != nil {
		return nil, false, err
	}

	if config.StoreConfig.AuthProvider != "" {
		return nil, false, fmt.Errorf("auth provider %s is not supported", config.StoreConfig.AuthProvider)
	}

	registryClientOptions := []remote.Option{
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	registryClientOptionsWrapper := []ociremote.Option{
		ociremote.WithRemoteOptions(registryClientOptions...),
	}

	return cosign.VerifyImageSignatures(ctx, ref, &cosign.CheckOpts{
		RootCerts:          nil, // TODO: TUF related metadata fulcio.Roots,
		SigVerifier:        ecdsaVerifier,
		ClaimVerifier:      cosign.SimpleClaimVerifier,
		RegistryClientOpts: registryClientOptionsWrapper,
	})
}

func loadPublicKey(ctx context.Context, keyRef string) (verifier signature.Verifier, err error) {
	keyPath := filepath.Clean(utils.ReplaceHomeShortcut(keyRef))
	raw, err := ioutil.ReadFile(keyPath)
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
