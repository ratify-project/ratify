package factory

import (
	"github.com/notaryproject/hora/pkg/executor"
	"github.com/notaryproject/hora/pkg/executor/core"
	sc "github.com/notaryproject/hora/pkg/referrerstore/config"
)

func CreateExecutorFromConfig(storesConfig sc.StoresConfig) (executor.Executor, error) {
	// TODO if cache is enabled, wrap with cache
	return core.Executor{}, nil
}
