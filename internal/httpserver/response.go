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
	"encoding/json"

	"github.com/notaryproject/ratify-go"
	"github.com/sirupsen/logrus"
)

// verificationResult is a rendered view of [ratify.VerificationResult].
type verificationResult struct {
	VerifierName string `json:"verifierName"`
	Description  string `json:"description,omitempty"`
	Detail       string `json:"detail,omitempty"`
	ErrorReason  string `json:"errorReason,omitempty"`
}

// validationReport is a rendered view of [ratify.ValidationReport].
type validationReport struct {
	Subject         string                `json:"subject"`
	Artifact        string                `json:"artifact"`
	Results         []*verificationResult `json:"results,omitempty"`
	ArtifactReports []*validationReport   `json:"artifactReports,omitempty"`
}

// result is a rendered view of [ratify.ValidationResult].
type result struct {
	Succeeded       bool                `json:"succeeded"`
	ArtifactReports []*validationReport `json:"artifactReports"`
}

func convertResult(src *ratify.ValidationResult) *result {
	if src == nil {
		return nil
	}
	result := &result{
		Succeeded:       src.Succeeded,
		ArtifactReports: convertValidationReports(src.ArtifactReports),
	}

	return result
}

func convertValidationReports(src []*ratify.ValidationReport) []*validationReport {
	if src == nil {
		return nil
	}
	reports := make([]*validationReport, len(src))
	for idx, report := range src {
		reports[idx] = convertValidationReport(report)
	}
	return reports
}

func convertValidationReport(src *ratify.ValidationReport) *validationReport {
	if src == nil {
		return nil
	}
	report := &validationReport{
		Subject:         src.Subject,
		Artifact:        src.Artifact.Digest.String(),
		ArtifactReports: convertValidationReports(src.ArtifactReports),
	}

	if len(src.Results) > 0 {
		report.Results = make([]*verificationResult, len(src.Results))
		for idx, result := range src.Results {
			report.Results[idx] = convertVerificationResult(result)
		}
	}

	return report
}

func convertVerificationResult(src *ratify.VerificationResult) *verificationResult {
	if src == nil {
		return nil
	}
	result := &verificationResult{
		Description: src.Description,
	}
	if src.Verifier != nil {
		result.VerifierName = src.Verifier.Name()
	}
	if src.Err != nil {
		result.ErrorReason = src.Err.Error()
	}
	if src.Detail != nil {
		detail, err := json.Marshal(src.Detail)
		if err != nil {
			logrus.Errorf("failed to marshal detail: %v", err)
			return nil
		}
		result.Detail = string(detail)
	}
	return result
}
