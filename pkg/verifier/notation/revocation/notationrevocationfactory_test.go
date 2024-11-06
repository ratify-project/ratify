// Copyright The Ratify Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package revocation

import (
	"net/http"
	"testing"

	"github.com/notaryproject/notation-core-go/revocation"
	"github.com/stretchr/testify/assert"
)

func TestNewFetcher(t *testing.T) {
	factory := &Notationrevocationfactory{}
	client := &http.Client{}

	fetcher, err := factory.NewFetcher(client)
	assert.NoError(t, err)
	assert.NotNil(t, fetcher)
}

func TestNewFileCache(t *testing.T) {
	factory := &Notationrevocationfactory{}
	cacheDir := "/cache"

	cache, err := factory.NewFileCache(cacheDir)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
}

func TestNewValidator(t *testing.T) {
	factory := &Notationrevocationfactory{}
	opts := revocation.Options{}

	validator, err := factory.NewValidator(opts)
	assert.NoError(t, err)
	assert.NotNil(t, validator)
}
