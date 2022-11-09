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
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	paths "path/filepath"
	"strings"

	ratifyconfig "github.com/deislabs/ratify/config"
	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/homedir"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/utils"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/config"
	"github.com/deislabs/ratify/pkg/verifier/factory"

	_ "github.com/notaryproject/notation-core-go/signature/cose"
	_ "github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/crypto/jwsutil"
	"github.com/notaryproject/notation-go/signature"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	verifierName    = "notaryv2"
	defaultCertPath = "ratify-certs/notary"
)

// NotaryV2VerifierConfig describes the configuration of notation verifier
type NotaryV2VerifierConfig struct {
	Name              string   `json:"name"`
	ArtifactTypes     string   `json:"artifactTypes"`
	VerificationCerts []string `json:"verificationCerts"`
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
	conf := NotaryV2VerifierConfig{}

	verifierConfigBytes, err := json.Marshal(verifierConfig)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(verifierConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse config for the input: %v", err)
	}

	defaultDir := paths.Join(homedir.Get(), ratifyconfig.ConfigFileDir, defaultCertPath)
	conf.VerificationCerts = append(conf.VerificationCerts, defaultDir)

	artifactTypes := strings.Split(conf.ArtifactTypes, ",")

	verfiyService, err := getVerifierService(conf.VerificationCerts...)
	if err != nil {
		return nil, err
	}

	return &notaryV2Verifier{artifactTypes: artifactTypes, notationVerifier: &verfiyService}, nil
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
	store referrerstore.ReferrerStore,
	executor executor.Executor) (verifier.VerifierResult, error) {

	// TODO get the subject descriptor
	desc := oci.Descriptor{
		Digest: subjectReference.Digest,
	}

	extensions := make(map[string]string)

	referenceManifest, err := store.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)

	if err != nil {
		return verifier.VerifierResult{IsSuccess: false}, err
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := store.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return verifier.VerifierResult{IsSuccess: false}, err
		}

		cert, err := getCert(refBlob)
		if err != nil {
			return verifier.VerifierResult{
				Subject:   subjectReference.String(),
				IsSuccess: false,
				Name:      verifierName,
				Message:   "error getting extension data from root cert",
			}, err
		}
		extensions["Issuer"] = cert.Issuer.String()
		extensions["SN"] = cert.Subject.String()

		opts := notation.VerifyOptions{
			SignatureMediaType: blobDesc.MediaType,
		}

		vdesc, err := (*v.notationVerifier).Verify(context.Background(), refBlob, opts)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return verifier.VerifierResult{IsSuccess: false, Extensions: extensions}, err
		}

		// TODO get the subject descriptor and verify all the properties other than digest.
		if desc.Digest != vdesc.Digest {
			return verifier.VerifierResult{
				Subject:    subjectReference.String(),
				Name:       verifierName,
				IsSuccess:  false,
				Message:    fmt.Sprintf("verification failure: digest mismatch: %v: %v", desc.Digest, vdesc.Digest),
				Extensions: extensions}, nil
		}
	}

	return verifier.VerifierResult{
		Name:       verifierName,
		IsSuccess:  true,
		Message:    "signature verification success",
		Extensions: extensions,
	}, nil
}

// This function is borrowed internals from the notation-go jws verifier
// https://github.com/notaryproject/notation-go/blob/main/signature/jws/verifier.go
func getCert(refBlob []byte) (*x509.Certificate, error) {
	var envelope jwsutil.Envelope
	if err := json.Unmarshal(refBlob, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Signatures) != 1 {
		return nil, errors.New("single signature envelope expected")
	}
	sig := envelope.Open()

	var header struct {
		TimeStampToken []byte   `json:"timestamp,omitempty"`
		CertChain      [][]byte `json:"x5c,omitempty"`
	}
	if err := json.Unmarshal(sig.Unprotected, &header); err != nil {
		return nil, err
	}
	if len(header.CertChain) == 0 {
		return nil, errors.New("signer certificates not found")
	}

	cert, err := x509.ParseCertificate(header.CertChain[0])
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func getVerifierService(certPaths ...string) (notation.Verifier, error) {
	certs := make([]*x509.Certificate, 0)
	for _, path := range certPaths {

		bundledCerts, err := utils.GetCertificatesFromPath(path)

		if err != nil {
			return nil, err
		}

		certs = append(certs, bundledCerts...)
	}

	verifier := signature.NewVerifier()
	verifier.TrustedCerts = certs
	return verifier, nil
}
