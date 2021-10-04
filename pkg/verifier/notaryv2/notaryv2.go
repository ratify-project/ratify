package notaryv2

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/executor"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/config"
	"github.com/deislabs/hora/pkg/verifier/factory"
	l "github.com/deislabs/hora/plugins/verifier/nv2verifier/local"
	"github.com/docker/libtrust"
	"github.com/notaryproject/notary/v2"

	x509nv2 "github.com/notaryproject/notary/v2/signature/x509"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	verifierName = "notaryv2"
)

type NotaryV2VerifierConfig struct {
	Name              string   `json:"name"`
	ArtifactTypes     string   `json:"artifactTypes"`
	VerificationCerts []string `json:"verificationCerts"`
}

type notaryV2Verifier struct {
	signingService notary.SigningService
	artifactTypes  []string
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

	vservice, err := getSigningService("", conf.VerificationCerts...)
	if err != nil {
		return nil, err
	}

	return &notaryV2Verifier{signingService: vservice, artifactTypes: artifactTypes}, nil
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
		return verifier.VerifierResult{}, err
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := store.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return verifier.VerifierResult{}, err
		}

		_, err = v.signingService.Verify(context.Background(), desc, refBlob)
		if err != nil {
			return verifier.VerifierResult{
				Subject:   subjectReference.String(),
				Name:      verifierName,
				IsSuccess: false,
				Results:   []string{fmt.Sprintf("signature verification failed: %v", err)},
			}, nil
		}
	}

	return verifier.VerifierResult{
		Name:      verifierName,
		IsSuccess: true,
		Results:   []string{"signature verification success"},
	}, nil
}

func getSigningService(keyPath string, certPaths ...string) (notary.SigningService, error) {
	var (
		key         libtrust.PrivateKey
		commonCerts []*x509.Certificate
		rootCerts   *x509.CertPool
		err         error
	)
	if keyPath != "" {
		key, err = x509nv2.ReadPrivateKeyFile(keyPath)
		if err != nil {
			return nil, err
		}
	}
	if len(certPaths) != 0 {
		rootCerts = x509.NewCertPool()
		for _, certPath := range certPaths {
			certs, err := x509nv2.ReadCertificateFile(certPath)
			if err != nil {
				return nil, err
			}
			commonCerts = append(commonCerts, certs...)
			for _, cert := range certs {
				rootCerts.AddCert(cert)
			}
		}
	}
	return l.NewSigningService(key, commonCerts, commonCerts, rootCerts)
}
