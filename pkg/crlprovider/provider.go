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

package crlprovider

import (
	"context"
	"net/http"
	"sync"
	"time"

	corecrl "github.com/notaryproject/notation-core-go/revocation/crl"
)

// static concurrency-safe map to store errors while fetching CRL from CRL management provider.
// layout:
//
// map["<namespace>/<name>"] = error
var CRLErrMap sync.Map

// GetCRL gets the CRL from the provider
func GetCRL(ctx context.Context, url string, timeout time.Duration) (*corecrl.Bundle, error) {
	if err, ok := CRLErrMap.Load(url); ok && err != nil {
		return nil, err.(error)
	}
	newFetcher, err := corecrl.NewHTTPFetcher(&http.Client{Timeout: timeout})
	if err != nil {
		return nil, err
	}
	bundle, err := newFetcher.Fetch(ctx, url)
	if err != nil {
		SetCRLError(url, err)
		return nil, err
	}
	CRLErrMap.Delete(url)
	return bundle, nil
}

// SetCRLError sets the error while fetching CRL from CRL management provider.
func SetCRLError(url string, err error) {
	CRLErrMap.Store(url, err)
}
