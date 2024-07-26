package mocks

import (
	"crypto"
	"crypto/x509"

	"github.com/ratify-project/ratify/pkg/keymanagementprovider"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/config"
)

type TestKeyManagementProviderFactory struct {
}

func (f *TestKeyManagementProviderFactory) Create(_ string, _ config.KeyManagementProviderConfig, _ string) (keymanagementprovider.KeyManagementProvider, error) {
	var certMap map[keymanagementprovider.KMPMapKey][]*x509.Certificate
	var keyMap map[keymanagementprovider.KMPMapKey]crypto.PublicKey
	return &TestKeyManagementProvider{certificates: certMap, keys: keyMap}, nil
}
