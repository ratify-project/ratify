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

	"github.com/deislabs/ratify/errors"
	ctxUtils "github.com/deislabs/ratify/internal/context"
	"github.com/deislabs/ratify/internal/logger"
	"github.com/deislabs/ratify/pkg/cache"
	"github.com/deislabs/ratify/pkg/executor"
	"github.com/deislabs/ratify/pkg/executor/types"
	"github.com/deislabs/ratify/pkg/metrics"
	"github.com/deislabs/ratify/pkg/referrerstore"
	pkgUtils "github.com/deislabs/ratify/pkg/utils"
	"github.com/deislabs/ratify/utils"

	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
)

const apiVersion = "externaldata.gatekeeper.sh/v1alpha1"

// verify validates provided images against the configured policy.
// The image key could be either a standalone image(repo:tag) or an image within a specific namespace([namespace]repo:tag).
// e.g.
// 1. docker.io/library/nginx:latest an image without a namespace would be evaluated by cluster-wide policy.
// 2. [ratify]docker.io/library/nginx:latest an image with a namespace would be evaluated by namespaced policy.
func (server *Server) verify(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()
	sanitizedMethod := utils.SanitizeString(r.Method)
	sanitizedURL := utils.SanitizeURL(*r.URL)
	logger.GetLogger(ctx, server.LogOption).Debugf("start request %s %s", sanitizedMethod, sanitizedURL)

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
	for _, key := range providerRequest.Request.Keys {
		wg.Add(1)
		go func(key string, ctx context.Context) {
			defer wg.Done()
			routineStartTime := time.Now()
			returnItem := externaldata.Item{
				Key: key,
			}
			defer func() {
				mu.Lock()
				results = append(results, returnItem)
				mu.Unlock()
			}()
			requestKey, err := pkgUtils.ParseRequestKey(key)
			if err != nil {
				returnItem.Error = err.Error()
				return
			}
			subjectReference, err := pkgUtils.ParseSubjectReference(requestKey.Subject)
			if err != nil {
				returnItem.Error = err.Error()
				return
			}
			ctx = ctxUtils.SetContextWithNamespace(ctx, requestKey.Namespace)

			if subjectReference.Digest.String() == "" {
				logger.GetLogger(ctx, server.LogOption).Warn("Digest should be used instead of tagged reference. The resolved digest may not point to the same signed artifact, since tags are mutable.")
			}
			resolvedSubjectReference := subjectReference.Original
			unlock := server.keyMutex.Lock(resolvedSubjectReference)
			defer unlock()

			logger.GetLogger(ctx, server.LogOption).Infof("verifying subject %v", resolvedSubjectReference)
			var result types.VerifyResult
			found := false
			cacheHit := false
			var cacheResponse string
			cacheProvider := cache.GetCacheProvider()
			if cacheProvider != nil {
				cacheResponse, found = cacheProvider.Get(ctx, fmt.Sprintf(cache.CacheKeyVerifyHandler, resolvedSubjectReference))
			}
			if found && cacheResponse != "" {
				if err := json.Unmarshal([]byte(cacheResponse), &result); err != nil {
					err = errors.ErrorCodeDataDecodingFailure.WithError(err).WithDetail(fmt.Sprintf("unable to unmarshal cache entry for subject %v", resolvedSubjectReference))
					logger.GetLogger(ctx, server.LogOption).Warn(err)
				} else {
					cacheHit = true
					logger.GetLogger(ctx, server.LogOption).Debugf("cache hit for subject %v", resolvedSubjectReference)
				}
			}
			if !cacheHit {
				verifyParameters := executor.VerifyParameters{
					Subject: resolvedSubjectReference,
				}

				if result, err = server.GetExecutor().VerifySubject(ctx, verifyParameters); err != nil {
					returnItem.Error = errors.ErrorCodeExecutorFailure.WithError(err).WithComponentType(errors.Executor).Error()
					return
				}

				if cacheProvider != nil {
					logger.GetLogger(ctx, server.LogOption).Debugf("cache miss for subject %v", resolvedSubjectReference)
					if !cacheProvider.SetWithTTL(ctx, fmt.Sprintf(cache.CacheKeyVerifyHandler, resolvedSubjectReference), result, server.CacheTTL) {
						logger.GetLogger(ctx, server.LogOption).Warnf("unable to insert cache entry for subject %v", resolvedSubjectReference)
					}
				}

				if res, err := json.MarshalIndent(result, "", "  "); err == nil {
					logger.GetLogger(ctx, server.LogOption).Infof("verify result for subject %s: %s", resolvedSubjectReference, string(res))
				}
			}

			returnItem.Value = fromVerifyResult(result, server.GetExecutor().PolicyEnforcer.GetPolicyType(ctx))
			logger.GetLogger(ctx, server.LogOption).Debugf("verification: execution time for image %s: %dms", resolvedSubjectReference, time.Since(routineStartTime).Milliseconds())
		}(utils.SanitizeString(key), ctx)
	}
	wg.Wait()
	elapsedTime := time.Since(startTime).Milliseconds()
	logger.GetLogger(ctx, server.LogOption).Debugf("verification: execution time for request: %dms", elapsedTime)
	metrics.ReportVerificationRequest(ctx, elapsedTime)
	return sendResponse(&results, "", w, http.StatusOK, false)
}

func (server *Server) mutate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()
	sanitizedMethod := utils.SanitizeString(r.Method)
	sanitizedURL := utils.SanitizeURL(*r.URL)
	logger.GetLogger(ctx, server.LogOption).Debugf("start request %s %s", sanitizedMethod, sanitizedURL)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.ErrorCodeBadRequest.WithError(err).WithDetail("unable to read request body")
	}

	// parse request body
	var providerRequest externaldata.ProviderRequest
	if err = json.Unmarshal(body, &providerRequest); err != nil {
		return errors.ErrorCodeBadRequest.WithError(err).WithDetail("unable to unmarshal request body")
	}

	results := make([]externaldata.Item, 0)
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, image := range providerRequest.Request.Keys {
		wg.Add(1)
		go func(image string) {
			defer wg.Done()
			routineStartTime := time.Now()
			logger.GetLogger(ctx, server.LogOption).Infof("mutating image %v", image)
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
				err = errors.ErrorCodeReferenceInvalid.WithError(err).WithDetail(fmt.Sprintf("failed to parse image reference %s", image))
				logger.GetLogger(ctx, server.LogOption).Error(err)
				returnItem.Error = err.Error()
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
					err := errors.ErrorCodeReferrerStoreFailure.WithDetail(fmt.Sprintf("failed to mutate image reference %s: could not find matching store %s", image, server.MutationStoreName)).WithComponentType(errors.ReferrerStore)
					logger.GetLogger(ctx, server.LogOption).Error(err)
					returnItem.Error = err.Error()
					return
				}
				descriptor, err := selectedStore.GetSubjectDescriptor(ctx, parsedReference)
				if err != nil {
					err = errors.ErrorCodeGetSubjectDescriptorFailure.NewError(errors.ReferrerStore, selectedStore.Name(), errors.EmptyLink, err, fmt.Sprintf("failed to get subject descriptor for image %s", image), errors.HideStackTrace)
					returnItem.Error = err.Error()
					return
				}
				returnItem.Value = fmt.Sprintf("%s@%s", parsedReference.Path, descriptor.Digest.String())
			}
			logger.GetLogger(ctx, server.LogOption).Debugf("mutation: execution time for image %s: %dms", image, time.Since(routineStartTime).Milliseconds())
		}(utils.SanitizeString(image))
	}
	wg.Wait()
	elapsedTime := time.Since(startTime).Milliseconds()
	logger.GetLogger(ctx, server.LogOption).Debugf("mutation: execution time for request: %dms", elapsedTime)
	metrics.ReportMutationRequest(ctx, elapsedTime)
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
		metrics.ReportSystemError(context.Background(), systemErr) // this context should not be tied to the lifetime of the request
		response.Response.SystemError = systemErr
	}

	w.WriteHeader(respCode)
	return json.NewEncoder(w).Encode(response)
}

func processTimeout(h ContextHandler, duration time.Duration, isMutation bool) ContextHandler {
	return func(handlerContext context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), duration)
		defer cancel()

		ctx = logger.InitContext(ctx, r)

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
