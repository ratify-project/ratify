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

	e "github.com/deislabs/ratify/pkg/executor"
	"github.com/sirupsen/logrus"
)

func (server *Server) verify(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	logrus.Infof("start request %v %v", r.Method, r.URL)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	subject := string(body)

	logrus.Infof("subject for request %v %v is %v", r.Method, r.URL, subject)

	verifyParameters := e.VerifyParameters{
		Subject: subject,
	}

	result, err := server.Executor.VerifySubject(ctx, verifyParameters)

	if err != nil {
		return err
	}

	logrus.Infof("request %v %v completed successfully", r.Method, r.URL)
	res, err := json.MarshalIndent(result, "", "  ")
	if err == nil {
		fmt.Println(string(res))
	}

	if result.IsSuccess {
		return serveJSON(w, "true")
	} else {
		return serveJSON(w, "false")
	}
}

func serveJSON(w http.ResponseWriter, result string) error {
	if err := json.NewEncoder(w).Encode(result); err != nil {
		return err
	}
	return nil
}
