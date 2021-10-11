package local

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/docker/libtrust"
	"github.com/notaryproject/notary/v2"
	"github.com/notaryproject/notary/v2/signature"
	x509nv2 "github.com/notaryproject/notary/v2/signature/x509"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

type signingService struct {
	*signature.Scheme
}

// NewSigningService create a simple signing service.
func NewSigningService(signingKey libtrust.PrivateKey, signingCerts, verificationCerts []*x509.Certificate, roots *x509.CertPool) (notary.SigningService, error) {
	scheme := signature.NewScheme()

	if signingKey != nil {
		signer, err := x509nv2.NewSigner(signingKey, signingCerts)
		if err != nil {
			return nil, err
		}
		scheme.RegisterSigner("", signer)
	}

	verifier, err := x509nv2.NewVerifier(verificationCerts, roots)
	if err != nil {
		return nil, err
	}
	scheme.RegisterVerifier(verifier)

	return &signingService{
		Scheme: scheme,
	}, nil
}

func (s *signingService) Sign(ctx context.Context, desc oci.Descriptor, references ...string) ([]byte, error) {
	claims := signature.Claims{
		Manifest: signature.Manifest{
			Descriptor: convertDescriptor(desc),
			References: references,
		},
		IssuedAt: time.Now().Unix(),
	}

	sig, err := s.Scheme.Sign("", claims)
	if err != nil {
		return nil, err
	}

	return []byte(sig), nil
}

func (s *signingService) Verify(ctx context.Context, desc oci.Descriptor, sig []byte) ([]string, error) {
	claims, err := s.Scheme.Verify(string(sig))
	if err != nil {
		return nil, fmt.Errorf("verification failure: %v", err)
	}

	// TODO remove this by passing subject desc
	if desc.Digest.String() != claims.Manifest.Descriptor.Digest {
		return nil, fmt.Errorf("verification failure: digest mismatch: %v: %v",
			desc,
			claims.Manifest.Descriptor,
		)
	}

	return claims.Manifest.References, nil
}

func convertDescriptor(desc oci.Descriptor) signature.Descriptor {
	return signature.Descriptor{
		MediaType: desc.MediaType,
		Digest:    desc.Digest.String(),
		Size:      desc.Size,
	}
}
