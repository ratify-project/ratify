package executor

import (
	"context"

	"github.com/deislabs/ratify/pkg/executor/types"
)

type VerifyParameters struct {
	Subject        string   `json:"subjectReference"`
	ReferenceTypes []string `json:"referenceTypes,omitempty"`
}

type Executor interface {
	VerifySubject(ctx context.Context, verifyParameters VerifyParameters) (types.VerifyResult, error)
}
