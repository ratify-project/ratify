package types

type VerifyResult struct {
	IsSuccess       bool          `json:"isSuccess"`
	VerifierReports []interface{} `json:"verifierReports,omitempty"`
}
