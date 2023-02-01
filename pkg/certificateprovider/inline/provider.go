package inline

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

const (
	// ValueParameter is the name of the parameter that contains the certificate (chain) as a string in PEM format
	ValueParameter = "value"
)

func GetCertificates(ctx context.Context, attrib map[string]string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	value, ok := attrib[ValueParameter]
	if !ok {
		return nil, fmt.Errorf("value parameter is not set")
	}

	block, rest := pem.Decode([]byte(value))
	if block == nil && len(rest) > 0 {
		return nil, fmt.Errorf("failed to decode pem block")
	}

	for block != nil {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("error parsing x509 certificate: %w", err)
			}
			certs = append(certs, cert)
			block, rest = pem.Decode(rest)
			if block == nil && len(rest) > 0 {
				return nil, fmt.Errorf("failed to decode pem block")
			}
		}
	}

	return certs, nil
}
