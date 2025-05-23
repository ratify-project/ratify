/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package executor

import (
	"fmt"

	"github.com/notaryproject/ratify-go"
	"github.com/notaryproject/ratify/v2/internal/policyenforcer"
	policyFactory "github.com/notaryproject/ratify/v2/internal/policyenforcer/factory"
	"github.com/notaryproject/ratify/v2/internal/store"
	"github.com/notaryproject/ratify/v2/internal/verifier"
	"github.com/notaryproject/ratify/v2/internal/verifier/factory"
)

// Options contains the configuration options to create a new executor.
type Options struct {
	// Verifiers contains the configuration options for the verifiers. Required.
	Verifiers []factory.NewVerifierOptions `json:"verifiers"`

	// Stores contains the configuration options for the stores. Required.
	Stores store.PatternOptions `json:"stores"`

	// Policy contains the configuration options for the policy enforcer.
	// Optional.
	Policy *policyFactory.NewPolicyEnforcerOptions `json:"policyEnforcer,omitempty"`
}

// NewExecutor creates a new [ratify.Executor] instance based on the provided
// options.
func NewExecutor(opts *Options) (*ratify.Executor, error) {
	if opts == nil {
		return nil, fmt.Errorf("executor options cannot be nil")
	}
	verifiers, err := verifier.NewVerifiers(opts.Verifiers)
	if err != nil {
		return nil, err
	}

	storeMux, err := store.NewStore(opts.Stores)
	if err != nil {
		return nil, err
	}

	policy, err := policyenforcer.NewPolicyEnforcer(opts.Policy)
	if err != nil {
		return nil, err
	}

	return ratify.NewExecutor(storeMux, verifiers, policy)
}
