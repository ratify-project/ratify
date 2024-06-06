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

package httpserver

import (
	"github.com/ratify-project/ratify/pkg/executor/types"
	pt "github.com/ratify-project/ratify/pkg/policyprovider/types"
)

const (
	VerificationResultVersion = "0.1.0"
	// Starting from this version, the verification result can be
	// evaluated by Ratify embedded OPA engine.
	ResultVersionSupportingRego = "1.0.0"
)

type VerificationResponse struct {
	Version         string        `json:"version"`
	IsSuccess       bool          `json:"isSuccess"`
	VerifierReports []interface{} `json:"verifierReports,omitempty"`
}

func fromVerifyResult(res types.VerifyResult, policyType string) VerificationResponse {
	version := VerificationResultVersion
	if policyType == pt.RegoPolicy {
		version = ResultVersionSupportingRego
	}
	return VerificationResponse{
		Version:         version,
		IsSuccess:       res.IsSuccess,
		VerifierReports: res.VerifierReports,
	}
}
