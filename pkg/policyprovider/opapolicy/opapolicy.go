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

package opapolicy

// import (
// 	"context"

// 	"github.com/deislabs/ratify/pkg/common"
// 	"github.com/deislabs/ratify/pkg/ocispecs"
// 	"github.com/open-policy-agent/opa/rego"
// )

// // PolicyEnforcer describes different polices that are enforced during verification
// type OPAPolicyEnforcer struct {
// }

// func (enforcer OPAPolicyEnforcer) VerifyNeeded(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) bool {
// 	module := `
// 	package example.authz

// 	default allow = false

// 	allow {
// 		regex.match(".+.azurecr.io$", input.subject)
// 	}
// 	`

// 	query, err := rego.New(
// 		rego.Query("x = data.example.authz.allow"),
// 		rego.Module("example.rego", module),
// 	).PrepareForEval(ctx)

// 	if err != nil {
// 		return false
// 	}

// 	input := map[string]interface{}{
// 		"method":  "GET",
// 		"path":    []interface{}{"salary", "bob"},
// 		"subject": "stuff.azurecr.io",
// 	}

// 	results, err := query.Eval(ctx, rego.EvalInput(input))
// 	if err != nil || !results.Allowed() {
// 		return false
// 	}

// 	return results[0].Bindings["x"].(bool)
// }
