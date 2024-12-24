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
	"sync"

	"github.com/notaryproject/notation-core-go/revocation"
	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
	"github.com/notaryproject/notation-go/dir"
	"github.com/ratify-project/ratify/config"
	re "github.com/ratify-project/ratify/errors"
)

type CRLHandler struct {
	CacheDisabled bool
	Fetcher       corecrl.Fetcher
	httpClient    *http.Client
}

var fetcherOnce sync.Once

// NewCRLHandler returns a new NewCRLHandler instance. Enable cache by default.
func NewCRLHandler() RevocationFactory {
	return &CRLHandler{CacheDisabled: false, httpClient: &http.Client{}}
}

func CreateCRLHandlerFromConfig(cf config.CRLConfig) RevocationFactory {
	return &CRLHandler{CacheDisabled: cf.CacheDisabled, httpClient: &http.Client{}}
}

// NewFetcher creates a new instance of a Fetcher if it doesn't already exist.
// If a Fetcher instance is already present, it returns the existing instance.
// The method also configures the cache for the Fetcher.
// Returns an instance of corecrl.Fetcher or an error if the Fetcher creation fails.
func (h *CRLHandler) NewFetcher() (corecrl.Fetcher, error) {
	var err error
	fetcherOnce.Do(func() {
		h.Fetcher, err = CreateCRLFetcher(h.httpClient, dir.PathCRLCache)
		if err == nil {
			h.configureCache()
		}
	})
	if err != nil {
		return nil, err
	}
	// Check if the fetcher is nil, return an error if it is.
	// one possible edge case is that an error happened in the first call,
	// the following calls will not get the error since the sync.Once block will be skipped.
	if h.Fetcher == nil {
		return nil, re.ErrorCodeConfigInvalid.WithDetail("failed to create CRL fetcher")
	}
	return h.Fetcher, nil
}

// NewValidator returns a new validator instance
func (h *CRLHandler) NewValidator(opts revocation.Options) (revocation.Validator, error) {
	return revocation.NewWithOptions(opts)
}

// configureCache disables the cache for the HTTPFetcher if caching is not enabled.
// If the EnableCache field is set to false, this method sets the Cache field of the
// HTTPFetcher to nil, effectively disabling caching for HTTP fetch operations.
func (h *CRLHandler) configureCache() {
	if h.CacheDisabled {
		if httpFetcher, ok := h.Fetcher.(*corecrl.HTTPFetcher); ok {
			httpFetcher.Cache = nil
		}
	}
}
