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

package types

import (
	"encoding/json"
	"io"

	"github.com/ratify-project/ratify/pkg/verifier"
)

const (
	SpecVersion      string = "0.1.0"
	Version          string = "version"
	Name             string = "name"
	Type             string = "type"
	ArtifactTypes    string = "artifactTypes"
	NestedReferences string = "nestedReferences"
	Source           string = "source"
)

const (
	ErrUnknown                     uint = iota // 0
	ErrConfigParsingFailure                    // 1
	ErrInvalidVerifierConfig                   // 2
	ErrUnknownCommand                          // 3
	ErrMissingEnvironmentVariables             // 4
	ErrIOFailure                               // 5
	ErrVersionNotSupported                     // 6
	ErrArgsParsingFailure                      // 7
	ErrPluginCmdFailure                        // 8
	ErrInternalFailure             uint = 999
)

// VerifierResult describes the verification result returned from the verifier plugin
type VerifierResult struct {
	IsSuccess  bool        `json:"isSuccess"`
	Message    string      `json:"message"`
	Name       string      `json:"name"`
	Type       string      `json:"type,omitempty"`
	Extensions interface{} `json:"extensions"`
}

// GetVerifierResult encodes the given JSON data into verify result object
func GetVerifierResult(result []byte) (*verifier.VerifierResult, error) {
	vResult := VerifierResult{}
	if err := json.Unmarshal(result, &vResult); err != nil {
		return nil, err
	}
	return &verifier.VerifierResult{
		IsSuccess:  vResult.IsSuccess,
		Message:    vResult.Message,
		Name:       vResult.Name,
		Type:       vResult.Type,
		Extensions: vResult.Extensions,
	}, nil
}

// WriteVerifyResultResult writes the given result as JSON data to the writer w
func WriteVerifyResultResult(result *verifier.VerifierResult, w io.Writer) error {
	return json.NewEncoder(w).Encode(result)
}

// NewVerifierResult creates a new verifier result object from the given
// verifier.VerifierResult.
func NewVerifierResult(result verifier.VerifierResult) VerifierResult {
	return VerifierResult{
		IsSuccess:  result.IsSuccess,
		Message:    result.Message,
		Name:       result.Name,
		Extensions: result.Extensions,
	}
}
