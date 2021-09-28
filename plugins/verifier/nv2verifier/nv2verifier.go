package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/plugin/skel"
	l "github.com/deislabs/hora/plugins/verifier/nv2verifier/local"
	"github.com/docker/libtrust"
	"github.com/notaryproject/notary/v2"

	x509nv2 "github.com/notaryproject/notary/v2/signature/x509"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

type PluginConfig struct {
	Name              string   `json:"name"`
	VerificationCerts []string `json:"verificationCerts"`
}

type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

func main() {
	skel.PluginMain("nv2verifier", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginConfig, error) {
	conf := PluginInputConfig{}

	//fmt.Print("test\n")
	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %v", err)
	}

	if len(conf.Config.VerificationCerts) == 0 {
		return nil, errors.New("verification certs are missins")
	}

	return &conf.Config, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	vservice, err := getSigningService("", input.VerificationCerts...)
	if err != nil {
		return nil, err
	}

	// TODO get the subject descriptor
	desc := oci.Descriptor{
		Digest: subjectReference.Digest,
	}

	ctx := context.Background()
	referenceManifest, err := referrerStore.GetReferenceManifest(ctx, subjectReference, referenceDescriptor)

	if err != nil {
		return nil, err
	}

	for _, blobDesc := range referenceManifest.Blobs {
		refBlob, err := referrerStore.GetBlobContent(ctx, subjectReference, blobDesc.Digest)
		if err != nil {
			return nil, err
		}

		_, err = vservice.Verify(context.Background(), desc, refBlob)
		if err != nil {
			return &verifier.VerifierResult{
				Subject:   subjectReference.String(),
				Name:      input.Name,
				IsSuccess: false,
				Results:   []string{fmt.Sprintf("signature verification failed: %v", err)},
			}, nil
		}
	}

	return &verifier.VerifierResult{
		Name:      input.Name,
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
