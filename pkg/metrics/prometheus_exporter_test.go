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
	"net/http"
	"testing"
	"time"
)

func TestInitPrometheusExporter(t *testing.T) {
	if err := initPrometheusExporter(8888); err != nil {
		t.Fatalf("initPrometheusExporter() error = %v", err)
	}
	time.Sleep(2 * time.Second)
	r, err := http.NewRequest("GET", "http://localhost:8888/metrics", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("http.DefaultClient.Do() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("http.DefaultClient.Do() resp.StatusCode = %v, expected %v", resp.StatusCode, http.StatusOK)
	}
}
