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

package notation

import (
	"encoding/json"
	"fmt"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/ratify-project/ratify-go"
	"github.com/ratify-project/ratify-verifier-go/notation"
	"github.com/ratify-project/ratify/v2/internal/verifier/factory"
)

const notationType = "notation"

type options struct {
	// TrustPolicyDocument is the trust policy document configured for the
	// Notation verifier to verify the signature. Required.
	TrustPolicyDocument *trustpolicy.Document `json:"trustPolicyDocument"`

	// TrustStorePath is the path to the trust store directory. Required.
	TrustStorePath string `json:"trustStorePath"`
}

func init() {
	factory.RegisterVerifierFactory(notationType, func(opts factory.NewVerifierOptions) (ratify.Verifier, error) {
		raw, err := json.Marshal(opts.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal verifier parameters: %w", err)
		}

		var params options
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verifier parameters: %w", err)
		}

		notationOpts := &notation.VerifierOptions{
			Name:           opts.Name,
			TrustPolicyDoc: params.TrustPolicyDocument,
			TrustStore:     truststore.NewX509TrustStore(dir.NewSysFS(params.TrustStorePath)),
		}

		return notation.NewVerifier(notationOpts)
	})
}
