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

package policyevaluation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/rego"
)

// OpaEngine is an OPA engine implementing PolicyEvaluator interface.
type OpaEngine struct {
	query rego.PreparedEvalQuery
}

const query = "data.ratify.policy.valid"

// NewOpaEngine creates a new OPA engine.
func NewOpaEngine(policy string) (*OpaEngine, error) {
	engine := &OpaEngine{}
	if err := engine.UpdatePolicy(policy); err != nil {
		return nil, err
	}
	return engine, nil
}

// UpdatePolicy updates the policy of the engine.
func (oa *OpaEngine) UpdatePolicy(policy string) error {
	trimmedPolicy := strings.TrimSpace(policy)
	if trimmedPolicy == "" {
		return errors.New("policy is empty")
	}

	query, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", trimmedPolicy),
	).PrepareForEval(context.Background())
	if err != nil {
		return err
	}

	oa.query = query
	return nil
}

// Evaluate evaluates the policy with the given input.
func (oe *OpaEngine) Evaluate(ctx context.Context, input map[string]interface{}) (bool, error) {
	results, err := oe.query.Eval(ctx, rego.EvalInput(input))

	if err != nil {
		return false, err
	} else if len(results) == 0 {
		return false, errors.New("no results returned from OPA")
	} else {
		result, ok := results[0].Expressions[0].Value.(bool)
		if !ok {
			return false, fmt.Errorf("unexpected result type: %v", results[0].Expressions[0].Value)
		}
		return result, nil
	}
}
