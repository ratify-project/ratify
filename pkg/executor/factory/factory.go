package factory

import (
	"github.com/deislabs/hora/pkg/executor"
	"github.com/deislabs/hora/pkg/executor/core"
	sc "github.com/deislabs/hora/pkg/referrerstore/config"
)

func CreateExecutorFromConfig(storesConfig sc.StoresConfig) (executor.Executor, error) {
	// TODO if cache is enabled, wrap with cache
	return core.Executor{}, nil
}
