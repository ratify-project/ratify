package core

import (
	"context"
	"time"

	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/verifiercache"
)

type ExecutorWithCache struct {
	base                   executor.Executor
	verifierCache          verifiercache.VerifierCache
	verfierCacheItemExpiry time.Duration
}

func (executor ExecutorWithCache) VerifySubject(ctx context.Context, verifyParameters executor.VerifyParameters) (types.VerifyResult, error) {
	// check the cache for the existence of item
	cachedResult, ok := executor.verifierCache.GetVerifyResult(ctx, verifyParameters.Subject)

	if ok {
		return cachedResult, nil
	}

	result, err := executor.base.VerifySubject(ctx, verifyParameters)

	if err == nil {
		executor.verifierCache.SetVerifyResult(ctx, verifyParameters.Subject, result, executor.verfierCacheItemExpiry)
	}

	return result, err
}
