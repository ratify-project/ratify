package memory

import (
	"context"
	"time"

	et "github.com/deislabs/hora/pkg/executor/types"
)

type MemoryCache struct {
	syncMap *SyncMapWithExpiration
}

func (memoryCache MemoryCache) GetVerifyResult(ctx context.Context, subjectRefString string) (et.VerifyResult, bool) {
	item, ok := memoryCache.syncMap.GetEntry(subjectRefString)
	if !ok {
		return et.VerifyResult{}, false
	}
	return item.(et.VerifyResult), true
}

func (memoryCache MemoryCache) SetVerifyResult(ctx context.Context, subjectRefString string, verifyResult et.VerifyResult, ttl time.Duration) {
	memoryCache.syncMap.SetEntry(subjectRefString, verifyResult, ttl)
}
