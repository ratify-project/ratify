package utils

import (
	"fmt"

	_ "crypto/sha256"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

func ParseDigest(digestStr string) (digest.Digest, error) {

	digest, err := digest.Parse(digestStr)
	if err != nil {
		return "", fmt.Errorf("The digest of the subject is invalid %s %v", digestStr, err)
	}

	return digest, nil
}

func ParseSubjectReference(subRef string) (common.Reference, error) {
	parseResult, err := reference.Parse(subRef)
	if err != nil {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference %v", err)
	}

	var subjectRef common.Reference

	if named, ok := parseResult.(reference.Named); ok {
		subjectRef.Path = named.Name()
	} else {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference Path")
	}

	if digested, ok := parseResult.(reference.Digested); ok {
		subjectRef.Digest = digested.Digest()
	} else {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference digest")
	}

	if tag, ok := parseResult.(reference.Tagged); ok {
		subjectRef.Tag = tag.Tag()
	}

	subjectRef.Original = subRef

	return subjectRef, nil
}
