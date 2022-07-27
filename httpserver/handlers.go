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
	"io/ioutil"
	"net/http"
	"time"

	e "github.com/deislabs/ratify/pkg/executor"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"github.com/sirupsen/logrus"
)

const apiVersion = "externaldata.gatekeeper.sh/v1alpha1"

func (server *Server) verify(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	logrus.Infof("start request %v %v", r.Method, r.URL)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %v", err)
	}

	// parse request body
	var providerRequest externaldata.ProviderRequest
	err = json.Unmarshal(body, &providerRequest)
	if err != nil {
		return fmt.Errorf("unable to unmarshal request body: %v", err)
	}

	results := make([]externaldata.Item, 0)
	// iterate over all keys
	for _, subject := range providerRequest.Request.Keys {
		// TODO: Enable caching:  Providers should add a caching mechanism to avoid extra calls to external data sources.
		logrus.Infof("subject for request %v %v is %v", r.Method, r.URL, subject)

		verifyParameters := e.VerifyParameters{
			Subject: subject,
		}

		result, err := server.GetExecutor().VerifySubject(ctx, verifyParameters)

		if err != nil {
			return err
		}

		res, err := json.MarshalIndent(result, "", "  ")
		if err == nil {
			fmt.Println(string(res))
		}

		results = append(results, externaldata.Item{
			Key:   subject,
			Value: result,
		})
	}
	return sendResponse(&results, "", w, http.StatusOK)
}

func sendResponse(results *[]externaldata.Item, systemErr string, w http.ResponseWriter, respCode int) error {
	response := externaldata.ProviderResponse{
		APIVersion: apiVersion,
		Kind:       "ProviderResponse",
	}

	if results != nil {
		response.Response.Items = *results
	} else {
		response.Response.SystemError = systemErr
	}

	w.WriteHeader(respCode)
	return json.NewEncoder(w).Encode(response)
}

func processTimeout(h ContextHandler, duration time.Duration) ContextHandler {
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
			return sendResponse(nil, fmt.Sprintf("validate operation failed with error %v", err), w, http.StatusInternalServerError)
		}

		return nil
	}
}
