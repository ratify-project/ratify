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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/deislabs/ratify/pkg/executor/core"
	"github.com/deislabs/ratify/pkg/ocispecs"
	config "github.com/deislabs/ratify/pkg/policyprovider/configpolicy"
	"github.com/sirupsen/logrus"

	"github.com/deislabs/ratify/pkg/policyprovider/types"
	"github.com/deislabs/ratify/pkg/referrerstore"
	"github.com/deislabs/ratify/pkg/referrerstore/mocks"
	"github.com/deislabs/ratify/pkg/verifier"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/opencontainers/go-digest"
)

func TestServer_Timeout_Failed(t *testing.T) {
	timeoutDuration := 6
	testImageName := "localhost:5000/net-monitor:v1"
	t.Run("server_timeout_fail", func(t *testing.T) {
		body := new(bytes.Buffer)

		json.NewEncoder(body).Encode(externaldata.NewProviderRequest([]string{testImageName}))
		request := httptest.NewRequest(http.MethodPost, "/ratify/gatekeeper/v1/verify", bytes.NewReader(body.Bytes()))
		logrus.Infof("policies successfully created. %s", body.Bytes())

		responseRecorder := httptest.NewRecorder()

		testDigest := digest.FromString("test")
		configPolicy := config.PolicyEnforcer{
			ArtifactTypePolicies: map[string]types.ArtifactTypeVerifyPolicy{
				"test-type1": types.AnyVerifySuccess,
			}}
		store := &mocks.TestStore{References: []ocispecs.ReferenceDescriptor{
			{
				ArtifactType: "test-type1",
			}},
			ResolveMap: map[string]digest.Digest{
				"v1": testDigest,
			},
		}
		ver := &core.TestVerifier{
			CanVerifyFunc: func(at string) bool {
				return at == "test-type1"
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
		}

		handler := contextHandler{
			context: server.Context,
			handler: processTimeout(server.verify, server.GetExecutor().GetVerifyRequestTimeout()),
		}

		handler.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusInternalServerError {
			t.Errorf("Want status '%d', got '%d'", http.StatusInternalServerError, responseRecorder.Code)
		}
	})
}
