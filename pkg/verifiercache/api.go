package verifiercache

import (
	"context"
	"time"

	et "github.com/deislabs/ratify/pkg/executor/types"
)

type VerifierCache interface {
	GetVerifyResult(ctx context.Context, subjectRefString string) (et.VerifyResult, bool)
	SetVerifyResult(ctx context.Context, subjectRefString string, verifyResult et.VerifyResult, ttl time.Duration)
}
