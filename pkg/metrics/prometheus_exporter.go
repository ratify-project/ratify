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

package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	readHeaderTimeout = 5 * time.Second
)

func initPrometheusExporter(port int) error {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		server := &http.Server{
			Addr:              fmt.Sprintf(":%v", port),
			ReadHeaderTimeout: readHeaderTimeout,
		}
		err := server.ListenAndServe()
		if err != nil {
			logrus.Errorf("listen and server error: %v", err)
		}
	}()

	return nil
}
