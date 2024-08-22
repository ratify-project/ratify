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
	"context"
	"time"

	"github.com/ratify-project/ratify/internal/logger"
	"github.com/ratify-project/ratify/pkg/executor/types"
	pt "github.com/ratify-project/ratify/pkg/policyprovider/types"
)

const (
	VerificationResultVersion = "0.1.0"
	ResultVersion0_2_0        = "0.2.0"
	// Starting from this version, the verification result can be
	// evaluated by Ratify embedded OPA engine.
	ResultVersionSupportingRego = "1.0.0"
	ResultVersion1_1_0          = "1.1.0"
)

type VerificationResponse struct {
	Version         string        `json:"version"`
	IsSuccess       bool          `json:"isSuccess"`
	TraceID         string        `json:"traceID,omitempty"`
	Timestamp       string        `json:"timestamp,omitempty"`
	VerifierReports []interface{} `json:"verifierReports,omitempty"`
}

func fromVerifyResult(ctx context.Context, res types.VerifyResult, policyType string) VerificationResponse {
	version := ResultVersion0_2_0
	if policyType == pt.RegoPolicy {
		version = ResultVersion1_1_0
	}
	return VerificationResponse{
		Version:         version,
		IsSuccess:       res.IsSuccess,
		Timestamp:       time.Now().Format(time.RFC3339Nano),
		TraceID:         logger.GetTraceID(ctx),
		VerifierReports: res.VerifierReports,
	}
}
