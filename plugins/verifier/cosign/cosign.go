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
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	_ "github.com/deislabs/ratify/pkg/referrerstore/oras"
	"github.com/deislabs/ratify/pkg/utils"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/plugin/skel"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/opencontainers/go-digest"
	imgspec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/cosign/bundle"
	"github.com/sigstore/cosign/v2/pkg/oci"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
	"github.com/sigstore/sigstore/pkg/signature"
)

type PluginConfig struct {
	Name     string `json:"name"`
	KeyRef   string `json:"key"`
	RekorURL string `json:"rekorURL"`
	// config specific to the plugin
}

type StoreConfig struct {
	UseHttp bool `json:"useHttp,omitempty"`
}

type StoreWrapperConfig struct {
	StoreConfig StoreConfig `json:"store"`
}

type PluginInputConfig struct {
	Config             PluginConfig       `json:"config"`
	StoreWrapperConfig StoreWrapperConfig `json:"storeConfig"`
}

type Extension struct {
	SignatureExtension []cosignExtension `json:"signatures,omitempty"`
}

type cosignExtension struct {
	SignatureDigest digest.Digest `json:"signatureDigest"`
	IsSuccess       bool          `json:"isSuccess"`
	BundleVerified  bool          `json:"bundleVerified"`
	Err             error         `json:"error,omitempty"`
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
	ctx := context.Background()
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}
	keyRef := input.Config.KeyRef
	rekorURL := input.Config.RekorURL
	cosignOpts := &cosign.CheckOpts{
		ClaimVerifier: cosign.SimpleClaimVerifier,
	}

	var ecdsaVerifier signature.Verifier
	var roots *x509.CertPool
	if keyRef != "" {
		ecdsaVerifier, err = loadPublicKey(ctx, keyRef)
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to load public key: %w", err)), nil
		}
		cosignOpts.SigVerifier = ecdsaVerifier
	} else {
		roots, err = fulcio.GetRoots()
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to get fulcio roots: %w", err)), nil
		}
		cosignOpts.RootCerts = roots
		if cosignOpts.RootCerts == nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to initialize root certificates")), nil
		}
	}

	if rekorURL != "" {
		cosignOpts.RekorClient, err = rekor.NewClient(rekorURL)
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to create Rekor client from URL %s: %w", rekorURL, err)), nil
		}
		cosignOpts.CTLogPubKeys, err = cosign.GetCTLogPubs(ctx)
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to set Certificate Transparency Log public keys: %w", err)), nil
		}
		// Fetches the Rekor public keys from the Rekor server
		cosignOpts.RekorPubKeys, err = cosign.GetRekorPubs(ctx)
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to set Rekor public keys: %w", err)), nil
		}
	} else {
		// if no rekor url is provided, turn off transparency log verification and ignore SCTs
		cosignOpts.IgnoreTlog = true
		cosignOpts.IgnoreSCT = true
	}

	referenceManifest, err := referrerStore.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)
	if err != nil {
		return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to get reference manifest: %w", err)), nil
	}

	// manifest must be an OCI Image
	if referenceManifest.MediaType != imgspec.MediaTypeImageManifest {
		return errorToVerifyResult(input.Config.Name, fmt.Errorf("reference manifest is not an image")), nil
	}

	subjectDesc, err := referrerStore.GetSubjectDescriptor(ctx, subjectReference)
	if err != nil {
		return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to create subject hash: %w", err)), nil
	}
	subjectDescHash := v1.Hash{
		Algorithm: subjectDesc.Digest.Algorithm().String(),
		Hex:       subjectDesc.Digest.Hex(),
	}

	sigExtensions := make([]cosignExtension, 0)
	signatures := []oci.Signature{}
	for _, blob := range referenceManifest.Blobs {
		blobBytes, err := referrerStore.GetBlobContent(ctx, subjectReference, blob.Digest)
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to get blob content: %w", err)), nil
		}
		staticOpts, err := staticLayerOpts(blob)
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to parse static signature opts: %w", err)), nil
		}
		sig, err := static.NewSignature(blobBytes, blob.Annotations[static.SignatureAnnotationKey], staticOpts...)
		if err != nil {
			return errorToVerifyResult(input.Config.Name, fmt.Errorf("failed to generate static signature: %w", err)), nil
		}
		// The verification will return an error if the signature is not valid.
		bundleVerified, err := cosign.VerifyImageSignature(ctx, sig, subjectDescHash, cosignOpts)
		extension := cosignExtension{
			SignatureDigest: blob.Digest,
			IsSuccess:       true,
			BundleVerified:  bundleVerified,
		}
		if err != nil {
			extension.IsSuccess = false
			extension.Err = err
		} else {
			signatures = append(signatures, sig)
		}
		sigExtensions = append(sigExtensions, extension)
	}

	if len(signatures) > 0 {
		return &verifier.VerifierResult{
			Name:       input.Config.Name,
			IsSuccess:  true,
			Message:    "cosign verification success. valid signatures found",
			Extensions: Extension{SignatureExtension: sigExtensions},
		}, nil
	}

	errorResult := errorToVerifyResult(input.Config.Name, fmt.Errorf("no valid signatures found"))
	errorResult.Extensions = Extension{SignatureExtension: sigExtensions}
	return errorResult, nil
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

func staticLayerOpts(desc imgspec.Descriptor) ([]static.Option, error) {
	options := []static.Option{}
	options = append(options, static.WithAnnotations(desc.Annotations))
	cert := desc.Annotations[static.CertificateAnnotationKey]
	chain := desc.Annotations[static.ChainAnnotationKey]
	if cert != "" && chain != "" {
		options = append(options, static.WithCertChain([]byte(cert), []byte(chain)))
	}
	var rekorBundle bundle.RekorBundle
	if val, ok := desc.Annotations[static.BundleAnnotationKey]; ok {
		if err := json.Unmarshal([]byte(val), &rekorBundle); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal bundle from blob payload")
		}
		options = append(options, static.WithBundle(&rekorBundle))
	}

	return options, nil
}

func errorToVerifyResult(name string, err error) *verifier.VerifierResult {
	return &verifier.VerifierResult{
		IsSuccess: false,
		Name:      name,
		Message:   errors.Wrap(err, "cosign verification failed").Error(),
	}
}
