package verifier

import (
	"context"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
)

type VerifierResult struct {
	Subject       string           `json:"subject,omitempty"`
	IsSuccess     bool             `json:"isSuccess,omitempty"`
	Name          string           `json:"name,omitempty"`
	Results       []string         `json:"results,omitempty"`
	NestedResults []VerifierResult `json:"nestedResults,omitempty"`
}

type ReferenceVerifier interface {
	Name() string
	CanVerify(ctx context.Context, referenceDescriptor ocispecs.ReferenceDescriptor) bool
	Verify(ctx context.Context,
		subjectReference common.Reference,
		referenceDescriptor ocispecs.ReferenceDescriptor,
		referrerStore referrerstore.ReferrerStore,
		executor executor.Executor) (VerifierResult, error)
}
