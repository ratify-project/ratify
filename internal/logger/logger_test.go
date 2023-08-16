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

package logger

import (
	"context"
	"errors"
	"net/http"
	"testing"

	logstash "github.com/bshuster-repo/logrus-logstash-hook"
	dcontext "github.com/docker/distribution/context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	traceIDName       = "x-trace-id"
	testTraceID       = "test-trace-id"
	testComponentType = componentType("test")
)

var (
	testHeader = http.Header{}
)

func init() {
	testHeader.Set(traceIDName, testTraceID)
}

func TestInitLogger(t *testing.T) {
	testCases := []struct {
		name            string
		config          Config
		r               *http.Request
		expectedTraceID string
	}{
		{
			name:   "no headers set",
			config: Config{},
			r:      &http.Request{},
		},
		{
			name: "headers do not contain traceIDHeader",
			config: Config{
				RequestHeaders: make(map[string]string),
			},
			r: &http.Request{},
		},
		{
			name: "request does not contain traceIDHeader",
			config: Config{
				RequestHeaders: map[string]string{
					traceIDHeaderName: traceIDName,
				},
			},
			r: &http.Request{
				Header: http.Header{},
			},
		},
		{
			name: "request has its own traceIDHeader",
			config: Config{
				RequestHeaders: map[string]string{
					traceIDHeaderName: traceIDName,
				},
			},
			r: &http.Request{
				Header: testHeader,
			},
			expectedTraceID: testTraceID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := InitLogger(context.Background(), tc.r, tc.config)
			traceID := dcontext.GetStringValue(ctx, ContextKeyTraceID)
			if traceID == "" {
				t.Fatalf("expected non-empty traceID, but got empty one")
			}
			if tc.expectedTraceID != "" && traceID != tc.expectedTraceID {
				t.Fatalf("expected traceID %s, but got %s", tc.expectedTraceID, traceID)
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	opt := Option{
		ComponentType: testComponentType,
	}
	logger := GetLogger(context.Background(), opt)
	entry := logger.WithError(errors.New("test"))
	ct := entry.Data["component-type"]
	if ct != testComponentType {
		t.Fatalf("expected component type %s, but got %s", testComponentType, ct)
	}
}

func TestSetFormatter(t *testing.T) {
	t.Run("TextFormatter", func(t *testing.T) {
		err := SetFormatter("text")
		assert.NoError(t, err)

		// Assert that the formatter is set to TextFormatter
		assert.IsType(t, &logrus.TextFormatter{}, logrus.StandardLogger().Formatter)
	})

	t.Run("JSONFormatter", func(t *testing.T) {
		err := SetFormatter("json")
		assert.NoError(t, err)

		// Assert that the formatter is set to JSONFormatter
		assert.IsType(t, &logrus.JSONFormatter{}, logrus.StandardLogger().Formatter)
	})

	t.Run("LogstashFormatter", func(t *testing.T) {
		err := SetFormatter("logstash")
		assert.NoError(t, err)

		// Assert that the formatter is set to LogstashFormatter
		_, ok := logrus.StandardLogger().Formatter.(*logstash.LogstashFormatter)
		assert.True(t, ok)
	})

	t.Run("UnsupportedFormatter", func(t *testing.T) {
		err := SetFormatter("unsupported")
		assert.Error(t, err)

		// Assert that an error is returned for unsupported formatter
		expectedErrMsg := "unsupported logging formatter: unsupported"
		assert.Contains(t, err.Error(), expectedErrMsg)
	})
}
