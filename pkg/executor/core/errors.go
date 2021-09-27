package core

import "errors"

var (
	ReferrerStoreNotFound = errors.New("Cannot find any referrer stores for the given subject")
	ReferrersNotFound     = errors.New("no referrers found for this artifact")
)

type ReferenceParseError struct {
	Enclosed string
}

func (err ReferenceParseError) Error() string {
	return err.Enclosed
}
