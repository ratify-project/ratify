package types

import (
	"encoding/json"
	"io"

	"github.com/deislabs/hora/pkg/verifier"
)

const (
	SpecVersion      string = "0.1.0"
	Version          string = "version"
	Name             string = "name"
	ArtifactTypes    string = "artifactTypes"
	NestedReferences string = "nestedReferences"
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
	ErrTryAgainLater               uint = 11
	ErrInternal                    uint = 999
)

type VerifierResult struct {
	IsSuccess bool     `json:"isSuccess"`
	Results   []string `json:"results"`
	Name      string   `json:"name"`
}

func GetVerifierResult(result []byte) (*verifier.VerifierResult, error) {
	vResult := VerifierResult{}
	if err := json.Unmarshal(result, &vResult); err != nil {
		return nil, err
	}
	return &verifier.VerifierResult{
		IsSuccess: vResult.IsSuccess,
		Results:   vResult.Results,
		Name:      vResult.Name,
	}, nil
}

func WriteVerifyResultResult(result *verifier.VerifierResult, w io.Writer) error {
	return json.NewEncoder(w).Encode(result)
}
