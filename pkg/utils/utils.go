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

package utils

import (
	"fmt"

	_ "crypto/sha256"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

// ParseDigest parses the given string and returns a validated Digest object.
func ParseDigest(digestStr string) (digest.Digest, error) {
	digest, err := digest.Parse(digestStr)
	if err != nil {
		return "", fmt.Errorf("The digest of the subject is invalid %s: %w", digestStr, err)
	}

	return digest, nil
}

// ParseSubjectReference parses the given subject and returns a valid reference
func ParseSubjectReference(subRef string) (common.Reference, error) {
	parseResult, err := reference.Parse(subRef)
	if err != nil {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference: %w", err)
	}

	var subjectRef common.Reference
	if named, ok := parseResult.(reference.Named); ok {
		subjectRef.Path = named.Name()
	} else {
		return common.Reference{}, fmt.Errorf("failed to parse subject reference Path")
	}

	if digested, ok := parseResult.(reference.Digested); ok {
		subjectRef.Digest = digested.Digest()
	}

	if tag, ok := parseResult.(reference.Tagged); ok {
		subjectRef.Tag = tag.Tag()
	}

	subjectRef.Original = subRef

	return subjectRef, nil
}
