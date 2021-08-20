package utils

import (
	"fmt"
	"strings"

	_ "crypto/sha256"

	"github.com/notaryproject/hora/pkg/common"
	"github.com/opencontainers/go-digest"
)

func ParseSubjectReference(subRef string) (common.Reference, error) {
	referenceParts := strings.Split(subRef, "@")

	if len(referenceParts) != 2 {
		return common.Reference{}, fmt.Errorf("Subject reference %s should be a reference with Digest", subRef)
	}

	digest, err := digest.Parse(referenceParts[1])
	if err != nil {
		return common.Reference{}, fmt.Errorf("The digest of the subject is invalid %s %v", referenceParts[1], err)
	}

	// TODO switch to using distribution Reference object with Named and Path
	return common.Reference{
		Path:     referenceParts[0],
		Digest:   digest,
		Original: subRef,
	}, nil
}

func ParseDigest(digestStr string) (digest.Digest, error) {

	digest, err := digest.Parse(digestStr)
	if err != nil {
		return "", fmt.Errorf("The digest of the subject is invalid %s %v", digestStr, err)
	}

	return digest, nil
}

/*func ParseSubjectReference(subRef string) (common.Reference, error) {
	parseResult, err := reference.Parse(subRef)
	if err != nil {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference %v", err)
	}

	var subjectRef common.Reference

	if named, ok := parseResult.(reference.Named); ok {
		subjectRef.Path = named
	} else {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference Path")
	}

	if digested, ok := parseResult.(reference.Digested); ok {
		subjectRef.Digest = digested.Digest()
	} else {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference digest")
	}

	return subjectRef, nil
}*/
