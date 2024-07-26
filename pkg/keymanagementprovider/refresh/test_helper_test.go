package refresh

import (
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/factory"
	"github.com/ratify-project/ratify/pkg/keymanagementprovider/mocks"
)

func init() {
	// Register the mock KeyManagementProviderFactory
	factory.Register("test-kmp", &mocks.TestKeyManagementProviderFactory{})
}
