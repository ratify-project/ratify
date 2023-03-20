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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	exconfig "github.com/deislabs/ratify/pkg/executor/config"
	"github.com/deislabs/ratify/pkg/executor/core"
	"github.com/deislabs/ratify/pkg/ocispecs"
	config "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/deislabs/ratify/pkg/policyprovider/types"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/mocks"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/opencontainers/go-digest"
)

const testArtifactType string = "test-type1"

func TestServer_Timeout_Failed(t *testing.T) {
	timeoutDuration := 6
	testImageName := "localhost:5000/net-monitor:v1"
	t.Run("server_timeout_fail", func(t *testing.T) {
		body := new(bytes.Buffer)

		_ = json.NewEncoder(body).Encode(externaldata.NewProviderRequest([]string{testImageName}))
		request := httptest.NewRequest(http.MethodPost, "/ratify/gatekeeper/v1/verify", bytes.NewReader(body.Bytes()))
		logrus.Infof("policies successfully created. %s", body.Bytes())

		responseRecorder := httptest.NewRecorder()

		testDigest := digest.FromString("test")
		configPolicy := config.PolicyEnforcer{
			ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
				testArtifactType: types.AnyVerifySuccess,
			}}
		store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
			{
				ArtifactType: testArtifactType,
			}},
			ResolveMap: map[string]digest.Digest{
				"v1": testDigest,
			},
		}
		ver := &core.TestVerifier{
			CanVerifyFunc: func(at string) bool {
				return at == testArtifactType
			},
			VerifyResult: func(artifactType string) bool {
				time.Sleep(time.Duration(timeoutDuration) * time.Second)
				return true
			},
		}

		ex := &core.Executor{
			PolicyEnforcer: configPolicy,
			ReferrerStores: []referrerstore.ReferrerStore{store},
			Verifiers:      []verifier.ReferenceVerifier{ver},
		}

		getExecutor := func() *core.Executor {
			return ex
		}

		server := &Server{
			GetExecutor: getExecutor,
			Context:     request.Context(),

			keyMutex: keyMutex{},
			cache:    newSimpleCache(DefaultCacheTTL, DefaultCacheMaxSize),
		}

		handler := contextHandler{
			context: server.Context,
			handler: processTimeout(server.verify, server.GetExecutor().GetVerifyRequestTimeout(), false),
		}

		handler.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusInternalServerError {
			t.Errorf("Want status '%d', got '%d'", http.StatusInternalServerError, responseRecorder.Code)
		}
	})
}

// TestServer_MultipleSubjects_Success tests multiple subjects are verified concurrently
func TestServer_MultipleSubjects_Success(t *testing.T) {
	testImageNames := []string{"localhost:5000/net-monitor:v1", "localhost:5000/net-monitor:v2"}
	t.Run("server_multiple_subjects_success", func(t *testing.T) {
		body := new(bytes.Buffer)

		if err := json.NewEncoder(body).Encode(externaldata.NewProviderRequest(testImageNames)); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/ratify/gatekeeper/v1/verify", bytes.NewReader(body.Bytes()))
		logrus.Infof("policies successfully created. %s", body.Bytes())

		responseRecorder := httptest.NewRecorder()

		testDigest := digest.FromString("test")
		configPolicy := config.PolicyEnforcer{
			ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
				testArtifactType: types.AnyVerifySuccess,
			}}
		store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
			{
				ArtifactType: testArtifactType,
			}},
			ResolveMap: map[string]digest.Digest{
				"v1": testDigest,
				"v2": testDigest,
			},
			ExtraSubject: testImageNames[0],
		}
		ver := &core.TestVerifier{
			CanVerifyFunc: func(at string) bool {
				return at == testArtifactType
			},
			VerifyResult: func(artifactType string) bool {
				return true
			},
		}

		ex := &core.Executor{
			PolicyEnforcer: configPolicy,
			ReferrerStores: []referrerstore.ReferrerStore{store},
			Verifiers:      []verifier.ReferenceVerifier{ver},
			Config: &exconfig.ExecutorConfig{
				VerificationRequestTimeout: nil,
				MutationRequestTimeout:     nil,
			},
		}

		getExecutor := func() *core.Executor {
			return ex
		}

		server := &Server{
			GetExecutor: getExecutor,
			Context:     request.Context(),

			keyMutex: keyMutex{},
			cache:    newSimpleCache(DefaultCacheTTL, DefaultCacheMaxSize),
		}

		handler := contextHandler{
			context: server.Context,
			handler: processTimeout(server.verify, server.GetExecutor().GetVerifyRequestTimeout(), false),
		}

		handler.ServeHTTP(responseRecorder, request)
		var respBody externaldata.ProviderResponse
		if err := json.NewDecoder(responseRecorder.Result().Body).Decode(&respBody); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		retFirstKey := respBody.Response.Items[0].Key
		if retFirstKey != testImageNames[1] {
			t.Fatalf("Expected first subject response to be %s but got %s", testImageNames[1], retFirstKey)
		}
	})
}

func TestServer_Mutation_Success(t *testing.T) {
	timeoutDuration := 6
	testImageNameTagged := "localhost:5000/net-monitor:v1"
	testDigest := digest.FromString("test")
	testImageNameDigested := fmt.Sprintf("localhost:5000/net-monitor@%s", testDigest)
	t.Run("server_timeout_fail", func(t *testing.T) {
		body := new(bytes.Buffer)

		if err := json.NewEncoder(body).Encode(externaldata.NewProviderRequest([]string{testImageNameTagged})); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/ratify/gatekeeper/v1/mutate", bytes.NewReader(body.Bytes()))
		logrus.Infof("policies successfully created. %s", body.Bytes())

		responseRecorder := httptest.NewRecorder()

		configPolicy := config.PolicyEnforcer{
			ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
				testArtifactType: types.AnyVerifySuccess,
			}}
		store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
			{
				ArtifactType: testArtifactType,
			}},
			ResolveMap: map[string]digest.Digest{
				"v1": testDigest,
			},
		}
		ver := &core.TestVerifier{
			CanVerifyFunc: func(at string) bool {
				return at == testArtifactType
			},
			VerifyResult: func(artifactType string) bool {
				time.Sleep(time.Duration(timeoutDuration) * time.Second)
				return true
			},
		}

		ex := &core.Executor{
			PolicyEnforcer: configPolicy,
			ReferrerStores: []referrerstore.ReferrerStore{store},
			Verifiers:      []verifier.ReferenceVerifier{ver},
		}

		getExecutor := func() *core.Executor {
			return ex
		}

		server := &Server{
			GetExecutor:       getExecutor,
			Context:           request.Context(),
			MutationStoreName: store.Name(),

			keyMutex: keyMutex{},
			cache:    newSimpleCache(DefaultCacheTTL, DefaultCacheMaxSize),
		}

		handler := contextHandler{
			context: server.Context,
			handler: processTimeout(server.mutate, server.GetExecutor().GetMutationRequestTimeout(), true),
		}

		handler.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusOK {
			t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
		}

		var respBody externaldata.ProviderResponse
		if err := json.NewDecoder(responseRecorder.Result().Body).Decode(&respBody); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		retFirstValue := respBody.Response.Items[0].Value
		if retFirstValue != testImageNameDigested {
			t.Fatalf("Expected mutation response to be %s but got %s", testImageNameDigested, retFirstValue)
		}
	})
}

func TestServer_MultipleRequestsForSameSubject_Success(t *testing.T) {
	testImageNames := []string{"localhost:5000/net-monitor:v1", "localhost:5000/net-monitor:v1"}
	t.Run("server_multiple_subjects_success", func(t *testing.T) {
		body := new(bytes.Buffer)

		if err := json.NewEncoder(body).Encode(externaldata.NewProviderRequest(testImageNames)); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/ratify/gatekeeper/v1/verify", bytes.NewReader(body.Bytes()))
		logrus.Infof("policies successfully created. %s", body.Bytes())

		responseRecorder := httptest.NewRecorder()

		testDigest := digest.FromString("test")
		configPolicy := config.PolicyEnforcer{
			ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
				testArtifactType: types.AnyVerifySuccess,
			}}
		store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
			{
				ArtifactType: testArtifactType,
			}},
			ResolveMap: map[string]digest.Digest{
				"v1": testDigest,
				"v2": testDigest,
			},
			ExtraSubject: testImageNames[0],
		}
		ver := &core.TestVerifier{
			CanVerifyFunc: func(at string) bool {
				return at == testArtifactType
			},
			VerifyResult: func(artifactType string) bool {
				return true
			},
		}

		ex := &core.Executor{
			PolicyEnforcer: configPolicy,
			ReferrerStores: []referrerstore.ReferrerStore{store},
			Verifiers:      []verifier.ReferenceVerifier{ver},
			Config: &exconfig.ExecutorConfig{
				VerificationRequestTimeout: nil,
				MutationRequestTimeout:     nil,
			},
		}

		getExecutor := func() *core.Executor {
			return ex
		}

		server := &Server{
			GetExecutor: getExecutor,
			Context:     request.Context(),

			keyMutex: keyMutex{},
			cache:    newSimpleCache(DefaultCacheTTL, DefaultCacheMaxSize),
		}

		handler := contextHandler{
			context: server.Context,
			handler: processTimeout(server.verify, server.GetExecutor().GetVerifyRequestTimeout(), false),
		}

		handler.ServeHTTP(responseRecorder, request)
		var respBody externaldata.ProviderResponse
		if err := json.NewDecoder(responseRecorder.Result().Body).Decode(&respBody); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		retFirstKey := respBody.Response.Items[0].Key
		if retFirstKey != testImageNames[1] {
			t.Fatalf("Expected first subject response to be %s but got %s", testImageNames[1], retFirstKey)
		}
	})
}

// TestServer_MultipleSubjects_Success tests multiple subjects are verified concurrently
func TestServer_Verify_ParseReference_Failure(t *testing.T) {
	testImageNames := []string{"&&"}
	t.Run("server_verify_parsereference_failure", func(t *testing.T) {
		body := new(bytes.Buffer)

		if err := json.NewEncoder(body).Encode(externaldata.NewProviderRequest(testImageNames)); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/ratify/gatekeeper/v1/verify", bytes.NewReader(body.Bytes()))
		logrus.Infof("policies successfully created. %s", body.Bytes())

		responseRecorder := httptest.NewRecorder()

		ex := &core.Executor{
			PolicyEnforcer: config.PolicyEnforcer{},
			ReferrerStores: []referrerstore.ReferrerStore{&mocks.TestStore{}},
			Verifiers:      []verifier.ReferenceVerifier{&core.TestVerifier{}},
			Config: &exconfig.ExecutorConfig{
				VerificationRequestTimeout: nil,
				MutationRequestTimeout:     nil,
			},
		}

		getExecutor := func() *core.Executor {
			return ex
		}

		server := &Server{
			GetExecutor: getExecutor,
			Context:     request.Context(),

			keyMutex: keyMutex{},
			cache:    newSimpleCache(DefaultCacheTTL, DefaultCacheMaxSize),
		}

		handler := contextHandler{
			context: server.Context,
			handler: processTimeout(server.verify, server.GetExecutor().GetVerifyRequestTimeout(), false),
		}

		handler.ServeHTTP(responseRecorder, request)
		var respBody externaldata.ProviderResponse
		if err := json.NewDecoder(responseRecorder.Result().Body).Decode(&respBody); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		retFirstKey := respBody.Response.Items[0].Key
		retFirstErr := respBody.Response.Items[0].Error
		if retFirstKey != testImageNames[0] {
			t.Fatalf("Expected first subject response to be %s but got %s", testImageNames[0], retFirstKey)
		}
		expectedErr := errors.Wrap(reference.ErrReferenceInvalidFormat, "failed to parse subject reference")
		if retFirstErr != expectedErr.Error() {
			t.Fatalf("Expected first subject error to be %s but got %s", expectedErr.Error(), retFirstErr)
		}
	})
}
