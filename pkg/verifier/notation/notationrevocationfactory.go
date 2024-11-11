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

package notation

import (
	"net/http"

	"github.com/notaryproject/notation-core-go/revocation"
	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verifier/crl"
)

type RevocationFactoryImpl struct {
	cacheRoot  string
	httpClient *http.Client
}

// NewRevocationFactoryImpl returns a new NewRevocationFactoryImpl instance
func NewRevocationFactoryImpl() RevocationFactory {
	return &RevocationFactoryImpl{
		cacheRoot:  dir.PathCRLCache,
		httpClient: &http.Client{},
	}
}

// NewFetcher returns a new fetcher instance
func (f *RevocationFactoryImpl) NewFetcher() (corecrl.Fetcher, error) {
	crlFetcher, err := corecrl.NewHTTPFetcher(f.httpClient)
	if err != nil {
		return nil, err
	}
	crlFetcher.Cache, err = newFileCache(f.cacheRoot)
	if err != nil {
		return nil, err
	}
	return crlFetcher, nil
}

// NewValidator returns a new validator instance
func (f *RevocationFactoryImpl) NewValidator(opts revocation.Options) (revocation.Validator, error) {
	return revocation.NewWithOptions(opts)
}

// newFileCache returns a new file cache instance
func newFileCache(root string) (*crl.FileCache, error) {
	cacheRoot, err := dir.CacheFS().SysPath(root)
	if err != nil {
		return nil, err
	}
	return crl.NewFileCache(cacheRoot)
}
