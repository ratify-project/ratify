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
	"errors"
	"fmt"
	"testing"
)

func TestInitMetricsExporter_InvalidPortExporter(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		exporter    string
		expectedErr error
	}{
		{
			name:        "invalid port negative",
			port:        -1,
			exporter:    "prometheus",
			expectedErr: fmt.Errorf("invalid port %v", -1),
		},
		{
			name:        "invalid port positive",
			port:        68000,
			exporter:    "prometheus",
			expectedErr: fmt.Errorf("invalid port %v", 68000),
		},
		{
			name:        "invalid exporter",
			port:        8888,
			exporter:    "invalid",
			expectedErr: fmt.Errorf("unsupported metrics backend %v", "invalid"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitMetricsExporter(tt.exporter, tt.port); errors.Is(err, tt.expectedErr) {
				t.Errorf("InitMetricsExporter() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}
