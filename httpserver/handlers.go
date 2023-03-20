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
	"sync"
	"time"

	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/referrerstore"
	pkgUtils "github.com/deislabs/ratify/pkg/utils"
	"github.com/deislabs/ratify/utils"

	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/sirupsen/logrus"
)

const apiVersion = "externaldata.gatekeeper.sh/v1alpha1"

func (server *Server) verify(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()
	sanitizedMethod := utils.SanitizeString(r.Method)
	sanitizedURL := utils.SanitizeURL(*r.URL)
	logrus.Infof("start request %s %s", sanitizedMethod, sanitizedURL)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %w", err)
	}
	defer r.Body.Close()

	// parse request body
	var providerRequest externaldata.ProviderRequest
	if err = json.Unmarshal(body, &providerRequest); err != nil {
		return fmt.Errorf("unable to unmarshal request body: %w", err)
	}

	results := make([]externaldata.Item, 0)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	// iterate over all keys
	for _, subject := range providerRequest.Request.Keys {
		wg.Add(1)
		go func(subject string) {
			defer wg.Done()
			routineStartTime := time.Now()
			returnItem := externaldata.Item{
				Key: subject,
			}
			defer func() {
				mu.Lock()
				results = append(results, returnItem)
				mu.Unlock()
			}()
			subjectReference, err := pkgUtils.ParseSubjectReference(subject)
			if err != nil {
				returnItem.Error = err.Error()
				return
			}
			resolvedSubjectReference := subjectReference.Original
			unlock := server.keyMutex.Lock(resolvedSubjectReference)
			defer unlock()

			logrus.Infof("verifying subject %v", resolvedSubjectReference)
			var result types.VerifyResult
			res := server.cache.get(resolvedSubjectReference)
			if res != nil {
				logrus.Debugf("cache hit for subject %v", resolvedSubjectReference)
				result = *res
			} else {
				logrus.Debugf("cache miss for subject %v", resolvedSubjectReference)
				verifyParameters := executor.VerifyParameters{
					Subject: resolvedSubjectReference,
				}

				if result, err = server.GetExecutor().VerifySubject(ctx, verifyParameters); err != nil {
					returnItem.Error = err.Error()
					return
				}
				server.cache.set(resolvedSubjectReference, &result)

				if res, err := json.MarshalIndent(result, "", "  "); err == nil {
					fmt.Println(string(res))
				}
			}

			returnItem.Value = fromVerifyResult(result)
			logrus.Debugf("verification: execution time for image %s: %dms", resolvedSubjectReference, time.Since(routineStartTime).Milliseconds())
		}(utils.SanitizeString(subject))
	}
	wg.Wait()
	logrus.Debugf("verification: execution time for request: %dms", time.Since(startTime).Milliseconds())
	return sendResponse(&results, "", w, http.StatusOK, false)
}

func (server *Server) mutate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()
	sanitizedMethod := utils.SanitizeString(r.Method)
	sanitizedURL := utils.SanitizeURL(*r.URL)
	logrus.Infof("start request %s %s", sanitizedMethod, sanitizedURL)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %w", err)
	}

	// parse request body
	var providerRequest externaldata.ProviderRequest
	if err = json.Unmarshal(body, &providerRequest); err != nil {
		return fmt.Errorf("unable to unmarshal request body: %w", err)
	}

	results := make([]externaldata.Item, 0)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, image := range providerRequest.Request.Keys {
		wg.Add(1)
		go func(image string) {
			defer wg.Done()
			routineStartTime := time.Now()
			logrus.Infof("mutating image %v", image)
			returnItem := externaldata.Item{
				Key:   image,
				Value: image,
			}
			defer func() {
				mu.Lock()
				results = append(results, returnItem)
				mu.Unlock()
			}()
			parsedReference, err := pkgUtils.ParseSubjectReference(image)
			if err != nil {
				errMessage := fmt.Sprintf("failed to mutate image reference %s: %v", image, err)
				logrus.Error(errMessage)
				returnItem.Error = errMessage
				return
			}

			if parsedReference.Digest == "" {
				var selectedStore referrerstore.ReferrerStore
				for _, store := range server.GetExecutor().ReferrerStores {
					if store.Name() == server.MutationStoreName {
						selectedStore = store
						break
					}
				}
				if selectedStore == nil {
					errMessage := fmt.Sprintf("failed to mutate image reference %s: could not find matching store %s", image, server.MutationStoreName)
					logrus.Error(errMessage)
					returnItem.Error = errMessage
					return
				}
				descriptor, err := selectedStore.GetSubjectDescriptor(ctx, parsedReference)
				if err != nil {
					errMessage := fmt.Sprintf("failed to mutate image reference %s: %v", image, err)
					logrus.Error(errMessage)
					returnItem.Error = errMessage
					return
				}
				returnItem.Value = fmt.Sprintf("%s@%s", parsedReference.Path, descriptor.Digest.String())
			}
			logrus.Debugf("mutation: execution time for image %s: %dms", image, time.Since(routineStartTime).Milliseconds())
		}(utils.SanitizeString(image))
	}
	wg.Wait()
	logrus.Debugf("mutation: execution time for request: %dms", time.Since(startTime).Milliseconds())
	return sendResponse(&results, "", w, http.StatusOK, true)
}

func sendResponse(results *[]externaldata.Item, systemErr string, w http.ResponseWriter, respCode int, isMutation bool) error {
	response := externaldata.ProviderResponse{
		APIVersion: apiVersion,
		Kind:       "ProviderResponse",
	}

	// only mutation webhook can be invoked multiple times and thus must be idempotent
	response.Response.Idempotent = isMutation

	if results != nil {
		response.Response.Items = *results
	} else {
		response.Response.SystemError = systemErr
	}

	w.WriteHeader(respCode)
	return json.NewEncoder(w).Encode(response)
}

func processTimeout(h ContextHandler, duration time.Duration, isMutation bool) ContextHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), duration)
		defer cancel()

		r = r.WithContext(ctx)

		processDone := make(chan bool)
		var err error
		go func() {
			err = h(ctx, w, r)
			processDone <- true
		}()

		select {
		case <-ctx.Done():
			err = fmt.Errorf("operation timed out after duration %v", duration)
		case <-processDone:
		}

		if err != nil {
			return sendResponse(nil, fmt.Sprintf("operation failed with error %v", err), w, http.StatusInternalServerError, isMutation)
		}

		return nil
	}
}
