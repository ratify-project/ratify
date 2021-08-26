package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/deislabs/hora/pkg/common"
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

type PluginInput struct {
	Config PluginConfig `json:"config"`
	Blob   []byte       `json:"blob"`
}

func main() {
	skel.PluginMain("nv2verifier", "1.0.0", VerifyReference, []string{"1.0.0"})
}

func parseInput(stdin []byte) (*PluginInput, error) {
	conf := PluginInput{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %v", err)
	}

	if len(conf.Config.VerificationCerts) == 0 {
		return nil, errors.New("verification certs are missins")
	}

	return &conf, nil
}

func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference) (*verifier.VerifierResult, error) {
	input, err := parseInput(args.StdinData)
	if err != nil {
		return nil, err
	}

	vservice, err := getSigningService("", input.Config.VerificationCerts...)
	if err != nil {
		return nil, err
	}

	// TODO get the subject descriptor
	desc := oci.Descriptor{
		Digest: subjectReference.Digest,
	}

	_, err = vservice.Verify(context.Background(), desc, input.Blob)
	if err != nil {
		return &verifier.VerifierResult{
			Subject:   subjectReference.String(),
			Name:      input.Config.Name,
			IsSuccess: false,
			Results:   []string{fmt.Sprintf("%v", err)},
		}, nil
	}

	return &verifier.VerifierResult{
		Name:      input.Config.Name,
		IsSuccess: true,
		Results:   []string{"Notary verification success"},
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
