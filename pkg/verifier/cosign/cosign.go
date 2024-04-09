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

package cosign

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deislabs/ratify/internal/logger"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/utils"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/config"
	"github.com/deislabs/ratify/pkg/verifier/factory"
	"github.com/deislabs/ratify/pkg/verifier/types"

	re "github.com/deislabs/ratify/errors"
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
	Name             string   `json:"name"`
	Type             string   `json:"type,omitempty"`
	ArtifactTypes    string   `json:"artifactTypes"`
	KeyRef           string   `json:"key,omitempty"`
	RekorURL         string   `json:"rekorURL,omitempty"`
	NestedReferences []string `json:"nestedArtifactTypes,omitempty"`
}

type Extension struct {
	SignatureExtension []cosignExtension `json:"signatures,omitempty"`
}

type cosignExtension struct {
	SignatureDigest digest.Digest `json:"signatureDigest"`
	IsSuccess       bool          `json:"isSuccess"`
	BundleVerified  bool          `json:"bundleVerified"`
	Err             string        `json:"error,omitempty"`
}

type cosignVerifier struct {
	name             string
	verifierType     string
	artifactTypes    []string
	nestedReferences []string
	config           *PluginConfig
}

type cosignVerifierFactory struct{}

var logOpt = logger.Option{
	ComponentType: logger.Verifier,
}

const verifierType string = "cosign"

// init() registers the cosign verifier with the factory
func init() {
	factory.Register(verifierType, &cosignVerifierFactory{})
}

// Create creates a new cosign verifier
func (f *cosignVerifierFactory) Create(_ string, verifierConfig config.VerifierConfig, _ string, namespace string) (verifier.ReferenceVerifier, error) {
	logger.GetLogger(context.Background(), logOpt).Debugf("creating cosign verifier with config %v, namespace '%v'", verifierConfig, namespace)
	verifierName, hasName := verifierConfig[types.Name].(string)
	if !hasName {
		return nil, re.ErrorCodeConfigInvalid.WithComponentType(re.Verifier).WithPluginName(verifierName).WithDetail("missing name in verifier config")
	}

	config, err := parseVerifierConfig(verifierConfig)
	if err != nil {
		return nil, re.ErrorCodeConfigInvalid.WithComponentType(re.Verifier).WithPluginName(verifierName)
	}

	return &cosignVerifier{
		name:             verifierName,
		verifierType:     config.Type,
		artifactTypes:    strings.Split(config.ArtifactTypes, ","),
		nestedReferences: config.NestedReferences,
		config:           config,
	}, nil
}

// Name returns the name of the cosign verifier
func (v *cosignVerifier) Name() string {
	return v.name
}

// Type returns 'cosign' as the type of the verifier
func (v *cosignVerifier) Type() string {
	return verifierType
}

// CanVerify returns true if the referenceDescriptor's artifact type is in the list of artifact types supported by the verifier
func (v *cosignVerifier) CanVerify(_ context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool {
	for _, at := range v.artifactTypes {
		if at == "*" || at == referenceDescriptor.ArtifactType {
			return true
		}
	}
	return false
}

// Verify verifies the subject reference using the cosign verifier
func (v *cosignVerifier) Verify(ctx context.Context, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (verifier.VerifierResult, error) {
	cosignOpts := &cosign.CheckOpts{
		ClaimVerifier: cosign.SimpleClaimVerifier,
	}

	var ecdsaVerifier signature.Verifier
	var roots *x509.CertPool
	var err error
	if v.config.KeyRef != "" {
		ecdsaVerifier, err = loadPublicKey(v.config.KeyRef)
		if err != nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to load public key: %w", err)), nil
		}
		cosignOpts.SigVerifier = ecdsaVerifier
	} else {
		roots, err = fulcio.GetRoots()
		if err != nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to get fulcio roots: %w", err)), nil
		}
		cosignOpts.RootCerts = roots
		if cosignOpts.RootCerts == nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to initialize root certificates")), nil
		}
	}

	if v.config.RekorURL != "" {
		cosignOpts.RekorClient, err = rekor.NewClient(v.config.RekorURL)
		if err != nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to create Rekor client from URL %s: %w", v.config.RekorURL, err)), nil
		}
		cosignOpts.CTLogPubKeys, err = cosign.GetCTLogPubs(ctx)
		if err != nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to set Certificate Transparency Log public keys: %w", err)), nil
		}
		// Fetches the Rekor public keys from the Rekor server
		cosignOpts.RekorPubKeys, err = cosign.GetRekorPubs(ctx)
		if err != nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to set Rekor public keys: %w", err)), nil
		}
	} else {
		// if no rekor url is provided, turn off transparency log verification and ignore SCTs
		cosignOpts.IgnoreTlog = true
		cosignOpts.IgnoreSCT = true
	}

	referenceManifest, err := referrerStore.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)
	if err != nil {
		return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to get reference manifest: %w", err)), nil
	}

	// manifest must be an OCI Image
	if referenceManifest.MediaType != imgspec.MediaTypeImageManifest {
		return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("reference manifest is not an image")), nil
	}

	subjectDesc, err := referrerStore.GetSubjectDescriptor(ctx, subjectReference)
	if err != nil {
		return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to create subject hash: %w", err)), nil
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
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to get blob content: %w", err)), nil
		}
		staticOpts, err := staticLayerOpts(blob)
		if err != nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to parse static signature opts: %w", err)), nil
		}
		sig, err := static.NewSignature(blobBytes, blob.Annotations[static.SignatureAnnotationKey], staticOpts...)
		if err != nil {
			return errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("failed to generate static signature: %w", err)), nil
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
			extension.Err = err.Error()
		} else {
			signatures = append(signatures, sig)
		}
		sigExtensions = append(sigExtensions, extension)
	}

	if len(signatures) > 0 {
		return verifier.VerifierResult{
			Name:       v.name,
			Type:       v.verifierType,
			IsSuccess:  true,
			Message:    "cosign verification success. valid signatures found",
			Extensions: Extension{SignatureExtension: sigExtensions},
		}, nil
	}

	errorResult := errorToVerifyResult(v.name, v.verifierType, fmt.Errorf("no valid signatures found"))
	errorResult.Extensions = Extension{SignatureExtension: sigExtensions}
	return errorResult, nil
}

// GetNestedReferences returns the nested reference artifact types configured
func (v *cosignVerifier) GetNestedReferences() []string {
	return v.nestedReferences
}

// ParseVerifierConfig parses the verifier config and returns a PluginConfig
func parseVerifierConfig(verifierConfig config.VerifierConfig) (*PluginConfig, error) {
	verifierName, hasName := verifierConfig[types.Name].(string)
	if !hasName {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.Verifier, "", re.EmptyLink, nil, "missing name in verifier config", re.HideStackTrace)
	}
	conf := PluginConfig{}

	verifierConfigBytes, err := json.Marshal(verifierConfig)
	if err != nil {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.Verifier, verifierName, re.EmptyLink, err, nil, re.HideStackTrace)
	}

	if err := json.Unmarshal(verifierConfigBytes, &conf); err != nil {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.Verifier, verifierName, re.EmptyLink, err, fmt.Sprintf("failed to unmarshal to cosign verifier config from: %+v.", verifierConfig), re.HideStackTrace)
	}

	// if Type is not provided, use the Name as the Type (backwards compatibility)
	if conf.Type == "" {
		conf.Type = conf.Name
	}

	if conf.ArtifactTypes == "" {
		return nil, re.ErrorCodeConfigInvalid.NewError(re.Verifier, verifierName, re.EmptyLink, nil, "missing artifactTypes in verifier config", re.HideStackTrace)
	}

	return &conf, nil
}

// LoadPublicKey loads the public key from the keyRef
func loadPublicKey(keyRef string) (verifier signature.Verifier, err error) {
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

// StaticLayerOpts builds the cosign options for static layer signatures
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

// ErrorToVerifyResult returns a verifier result with the error message and isSuccess set to false
func errorToVerifyResult(name string, verifierType string, err error) verifier.VerifierResult {
	return verifier.VerifierResult{
		IsSuccess: false,
		Name:      name,
		Type:      verifierType,
		Message:   errors.Wrap(err, "cosign verification failed").Error(),
	}
}
