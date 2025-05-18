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

package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/ratify-project/ratify-go"
	"oras.land/oras-go/v2/registry"
)

// verify handles the verification request from Gatekeeper.
func (s *server) verify(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	var providerRequest externaldata.ProviderRequest
	if err = json.Unmarshal(body, &providerRequest); err != nil {
		return fmt.Errorf("failed to unmarshal request body to provider request: %w", err)
	}

	results := make([]externaldata.Item, len(providerRequest.Request.Keys))
	for idx, key := range providerRequest.Request.Keys {
		item := externaldata.Item{
			Key: key,
		}
		opts := ratify.ValidateArtifactOptions{
			Subject: key,
		}
		result, err := s.executor.ValidateArtifact(ctx, opts)
		if err != nil {
			item.Error = err.Error()
		}
		item.Value = convertResult(result)
		results[idx] = item
	}

	return sendResponse(results, w, http.StatusOK, false)
}

func (s *server) mutate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	var providerRequest externaldata.ProviderRequest
	if err = json.Unmarshal(body, &providerRequest); err != nil {
		return fmt.Errorf("failed to unmarshal request body to provider request: %w", err)
	}
	results := make([]externaldata.Item, len(providerRequest.Request.Keys))
	for idx, key := range providerRequest.Request.Keys {
		results[idx] = s.resolveReference(ctx, key)
	}

	return sendResponse(results, w, http.StatusOK, true)
}

func (s *server) resolveReference(ctx context.Context, reference string) externaldata.Item {
	item := externaldata.Item{
		Key:   reference,
		Value: reference,
	}

	ref, err := registry.ParseReference(reference)
	if err != nil {
		item.Error = fmt.Sprintf("failed to parse reference: %v", err)
		return item
	}
	if _, err = ref.Digest(); err == nil {
		item.Value = ref.String()
		return item
	}

	desc, err := s.executor.Store.Resolve(ctx, ref.String())
	if err != nil {
		item.Error = fmt.Sprintf("failed to resolve reference: %v", err)
		return item
	}
	ref.Reference = desc.Digest.String()
	item.Value = ref.String()
	return item
}

func sendResponse(results []externaldata.Item, w http.ResponseWriter, respCode int, isMutation bool) error {
	response := externaldata.ProviderResponse{
		APIVersion: "externaldata.gatekeeper.sh/v1beta1",
		Kind:       "ProviderResponse",
		Response: externaldata.Response{
			Idempotent: true,
			Items:      results,
		},
	}

	// Only mutation webhook can be invoked multiple times and thus must be idempotent.
	response.Response.Idempotent = isMutation

	w.WriteHeader(respCode)
	return json.NewEncoder(w).Encode(response)
}
