package factory

import (
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/core"
	sc "github.com/deislabs/ratify/pkg/referrerstore/config"
)

func CreateExecutorFromConfig(storesConfig sc.StoresConfig) (executor.Executor, error) {
	// TODO if cache is enabled, wrap with cache
	return core.Executor{}, nil
}
