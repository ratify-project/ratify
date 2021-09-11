package registry

import (
	"os"
	"sync"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
)

type AuthConfig struct {
	Username string
	Password string
}

type AuthProvider interface {
	Provide(resource string) (*AuthConfig, error)
}

type defaultProvider struct {
	mu sync.Mutex
}

var (
	DefaultAuthProvider AuthProvider = &defaultProvider{}
	anonymous                        = &AuthConfig{}
)

func (da *defaultProvider) Provide(resource string) (*AuthConfig, error) {
	da.mu.Lock()
	defer da.mu.Unlock()
	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return nil, err
	}

	cfg, err := cf.GetAuthConfig(resource)
	if err != nil {
		return nil, err
	}

	empty := types.AuthConfig{}
	if cfg == empty {
		return anonymous, nil
	}

	return &AuthConfig{
		Username: cfg.Username,
		Password: cfg.Password,
	}, nil
}
