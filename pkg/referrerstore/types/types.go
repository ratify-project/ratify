package types

import (
	"encoding/json"
	"io"

	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore"
)

const (
	SpecVersion string = "0.1.0"
	Version     string = "version"
	Name        string = "name"
)

const (
	ErrUnknown                     uint = iota // 0
	ErrConfigParsingFailure                    // 1
	ErrInvalidStoreConfig                      // 2
	ErrUnknownCommand                          // 3
	ErrMissingEnvironmentVariables             // 4
	ErrIOFailure                               // 5
	ErrVersionNotSupported                     // 6
	ErrArgsParsingFailure                      // 7
	ErrPluginCmdFailure                        // 8
	ErrTryAgainLater               uint = 11
	ErrInternal                    uint = 999
)

// TODO Versioned Referrers result (base don the verion serialize/deserialize the result)
type ReferrersResult struct {
	Referrers []ocispecs.ReferenceDescriptor `json:"referrers"`
	NextToken string                         `json:"nextToken"`
}

func GetListReferrersResult(result []byte) (referrerstore.ListReferrersResult, error) {
	listResult := ReferrersResult{}
	if err := json.Unmarshal(result, &listResult); err != nil {
		return referrerstore.ListReferrersResult{}, err
	}
	return referrerstore.ListReferrersResult{
		Referrers: listResult.Referrers,
		NextToken: listResult.NextToken,
	}, nil
}

func GetReferenceManifestResult(result []byte) (ocispecs.ReferenceManifest, error) {
	manifest := ocispecs.ReferenceManifest{}
	if err := json.Unmarshal(result, &manifest); err != nil {
		return ocispecs.ReferenceManifest{}, err
	}
	return manifest, nil
}

func WriteListReferrersResult(result *referrerstore.ListReferrersResult, w io.Writer) error {
	return json.NewEncoder(w).Encode(ReferrersResult{
		Referrers: result.Referrers,
		NextToken: result.NextToken,
	})
}

func WriteReferenceManifestResult(result *ocispecs.ReferenceManifest, w io.Writer) error {
	return json.NewEncoder(w).Encode(result)
}
