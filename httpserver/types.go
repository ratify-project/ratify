package httpserver

import (
	"github.com/deislabs/ratify/pkg/executor/types"
)

const VerificationResultVersion = "0.1.0"

type VerificationResponse struct {
	Version         string        `json:"version"`
	IsSuccess       bool          `json:"isSuccess"`
	VerifierReports []interface{} `json:"verifierReports,omitempty"`
}

func fromVerifyResult(res types.VerifyResult) VerificationResponse {
	return VerificationResponse{
		Version:         VerificationResultVersion,
		IsSuccess:       res.IsSuccess,
		VerifierReports: res.VerifierReports,
	}
}
