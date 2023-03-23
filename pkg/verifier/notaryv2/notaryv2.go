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

package notaryv2

import (
	"context"
	"encoding/json"
	"fmt"
	paths "path/filepath"
	"strings"

	ratifyconfig "github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/homedir"

	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/config"
	"github.com/deislabs/ratify/pkg/verifier/factory"

	_ "github.com/notaryproject/notation-core-go/signature/cose"
	_ "github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	notaryVerifier "github.com/notaryproject/notation-go/verifier"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	verifierName    = "notaryv2"
	defaultCertPath = "ratify-certs/notary/truststore"
)

// NotaryV2VerifierConfig describes the configuration of notation verifier
type NotaryV2VerifierConfig struct {
	Name          string `json:"name"`
	ArtifactTypes string `json:"artifactTypes"`

	// VerificationCerts is array of directories containing certificates.
	VerificationCerts []string `json:"verificationCerts"`
	// VerificationCerts is map defining which keyvault certificates belong to which trust store
	VerificationCertStores map[string][]string `json:"verificationCertStores"`
	// TrustPolicyDoc represents a trustpolicy.json document. Reference: https://pkg.go.dev/github.com/notaryproject/notation-go@v0.12.0-beta.1.0.20221125022016-ab113ebd2a6c/verifier/trustpolicy#Document
	TrustPolicyDoc trustpolicy.Document `json:"trustPolicyDoc"`
}

type notaryV2Verifier struct {
	artifactTypes    []string
	notationVerifier *notation.Verifier
}

type notaryv2VerifierFactory struct{}

func init() {
	factory.Register(verifierName, &notaryv2VerifierFactory{})
}

func (f *notaryv2VerifierFactory) Create(version string, verifierConfig config.VerifierConfig) (verifier.ReferenceVerifier, error) {
	conf, err := parseVerifierConfig(verifierConfig)
	if err != nil {
		return nil, err
	}

	verfiyService, err := getVerifierService(conf)
	if err != nil {
		return nil, err
	}

	artifactTypes := strings.Split(conf.ArtifactTypes, ",")
	return &notaryV2Verifier{
		artifactTypes:    artifactTypes,
		notationVerifier: &verfiyService,
	}, nil
}

func (v *notaryV2Verifier) Name() string {
	return verifierName
}

func (v *notaryV2Verifier) CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool {
	for _, at := range v.artifactTypes {
		if at == "*" || at == referenceDescriptor.ArtifactType {
			return true
		}
	}
	return false
}

func (v *notaryV2Verifier) Verify(ctx context.Context,
	subjectReference common.Reference,
	referenceDescriptor ocispecs.ReferenceDescriptor,
	store referrerstore.ReferrerStore) (verifier.VerifierResult, error) {
	extensions := make(map[string]string)

	subjectDesc, err := store.GetSubjectDescriptor(ctx, subjectReference)
	if err != nil {
		return verifier.VerifierResult{IsSuccess: false}, fmt.Errorf("failed to resolve subject: %+v, err: %w", subjectReference, err)
	}

	referenceManifest, err := store.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)
	if err != nil {
		return verifier.VerifierResult{IsSuccess: false}, fmt.Errorf("failed to get reference manifest for reference: %s, err: %w", subjectReference.Original, err)
	}

	if len(referenceManifest.Blobs) == 0 {
		return verifier.VerifierResult{IsSuccess: false}, fmt.Errorf("no signature content found for referrer: %s@%s", subjectReference.Path, referenceDescriptor.Digest.String())
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := store.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return verifier.VerifierResult{IsSuccess: false}, fmt.Errorf("failed to get blob content of digest: %s, err: %w", blobDesc.Digest, err)
		}

		// TODO: notary verify API only accepts digested reference now.
		// Pass in tagged reference instead once notation-go supports it.
		subjectRef := fmt.Sprintf("%s@%s", subjectReference.Path, subjectReference.Digest.String())
		outcome, err := v.verifySignature(ctx, subjectRef, blobDesc.MediaType, subjectDesc.Descriptor, refBlob)
		if err != nil {
			return verifier.VerifierResult{IsSuccess: false, Extensions: extensions}, fmt.Errorf("failed to verify signature, err: %w", err)
		}

		// Note: notary verifier already validates certificate chain is not empty.
		cert := outcome.EnvelopeContent.SignerInfo.CertificateChain[0]
		extensions["Issuer"] = cert.Issuer.String()
		extensions["SN"] = cert.Subject.String()
	}

	return verifier.VerifierResult{
		Name:       verifierName,
		IsSuccess:  true,
		Message:    "signature verification success",
		Extensions: extensions,
	}, nil
}

func getVerifierService(conf *NotaryV2VerifierConfig) (notation.Verifier, error) {
	store := &trustStore{
		certPaths:  conf.VerificationCerts,
		certStores: conf.VerificationCertStores,
	}

	return notaryVerifier.New(&conf.TrustPolicyDoc, store, nil)
}

func (v *notaryV2Verifier) verifySignature(ctx context.Context, subjectRef, mediaType string, subjectDesc oci.Descriptor, refBlob []byte) (*notation.VerificationOutcome, error) {
	opts := notation.VerifyOptions{
		SignatureMediaType: mediaType,
		ArtifactReference:  subjectRef,
	}

	return (*v.notationVerifier).Verify(ctx, subjectDesc, refBlob, opts)
}

func parseVerifierConfig(verifierConfig config.VerifierConfig) (*NotaryV2VerifierConfig, error) {
	conf := &NotaryV2VerifierConfig{}

	verifierConfigBytes, err := json.Marshal(verifierConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(verifierConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to notaryV2VerifierConfig from： %+v, err: %w", verifierConfig, err)
	}

	defaultCertsDir := paths.Join(homedir.Get(), ratifyconfig.ConfigFileDir, defaultCertPath)
	conf.VerificationCerts = append(conf.VerificationCerts, defaultCertsDir)
	return conf, nil
}

// signatures should not have nested references
func (v *notaryV2Verifier) GetNestedReferences() []string {
	return []string{}
}
