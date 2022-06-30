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
	"strings"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/utils"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/deislabs/ratify/pkg/verifier/config"
	"github.com/deislabs/ratify/pkg/verifier/factory"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/signature/jws"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	verifierName = "notaryv2"
)

// NotaryV2VerifierConfig describes the configuration of notation verifier
type NotaryV2VerifierConfig struct {
	Name              string   `json:"name"`
	ArtifactTypes     string   `json:"artifactTypes"`
	VerificationCerts []string `json:"verificationCerts"`
}

type notaryV2Verifier struct {
	artifactTypes    []string
	notationVerifier *jws.Verifier
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

	//fmt.Print("test\n")
	if err := json.Unmarshal(verifierConfigBytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse config for the input: %v", err)
	}

	if len(conf.VerificationCerts) == 0 {
		return nil, errors.New("verification certs are missing")
	}

	artifactTypes := strings.Split(fmt.Sprintf("%s", conf.ArtifactTypes), ",")

	verfiyService, err := getVerifierService(conf.VerificationCerts...)
	if err != nil {
		return nil, err
	}

	return &notaryV2Verifier{artifactTypes: artifactTypes, notationVerifier: verfiyService}, nil
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

	referenceManifest, err := store.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)

	if err != nil {
		return verifier.VerifierResult{IsSuccess: false}, err
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := store.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return verifier.VerifierResult{IsSuccess: false}, err
		}

		var opts notation.VerifyOptions
		vdesc, err := v.notationVerifier.Verify(context.Background(), refBlob, opts)
		if err != nil {
			return verifier.VerifierResult{IsSuccess: false}, err
		}

		// TODO get the subject descriptor and verify all the properties other than digest.
		if desc.Digest != vdesc.Digest {
			return verifier.VerifierResult{
				Subject:   subjectReference.String(),
				Name:      verifierName,
				IsSuccess: false,
				Results:   []string{fmt.Sprintf("verification failure: digest mismatch: %v: %v", desc.Digest, vdesc.Digest)}}, nil
		}
	}

	return verifier.VerifierResult{
		Name:      verifierName,
		IsSuccess: true,
		Results:   []string{"signature verification success"},
	}, nil
}

func getVerifierService(certPaths ...string) (*jws.Verifier, error) {
	roots := x509.NewCertPool()
	for _, path := range certPaths {

		bundledCerts, err := utils.GetCertificatesFromPath(path)

		if err != nil {
			return nil, err
		}

		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}
	verifier := jws.NewVerifier()
	verifier.VerifyOptions.Roots = roots
	return verifier, nil
}
