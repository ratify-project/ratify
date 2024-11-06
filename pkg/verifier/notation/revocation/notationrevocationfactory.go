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

	"github.com/notaryproject/notation-core-go/revocation"
	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-go/verifier/crl"
)

type Notationrevocationfactory struct{}

// NewFetcher returns a new fetcher instance
func (f *Notationrevocationfactory) NewFetcher(client *http.Client) (*corecrl.HTTPFetcher, error) {
	return corecrl.NewHTTPFetcher(client)
}

// NewFileCache returns a new file cache instance
func (f *Notationrevocationfactory) NewFileCache(cacheDir string) (corecrl.Cache, error) {
	return crl.NewFileCache(cacheDir)
}

// NewValidator returns a new validator instance
func (f *Notationrevocationfactory) NewValidator(opts revocation.Options) (revocation.Validator, error) {
	return revocation.NewWithOptions(opts)
}
