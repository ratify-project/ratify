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
	"fmt"
	"net/http"
	"time"

	logstash "github.com/bshuster-repo/logrus-logstash-hook"
	re "github.com/deislabs/ratify/errors"
	dcontext "github.com/docker/distribution/context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ContextKey defines the key type used for the context.
type ContextKey string

// String returns the string representation of the context key.
func (c ContextKey) String() string {
	return string(c)
}

type componentType string

// Option is used for each component to add customized fields to the logger.
type Option struct {
	ComponentType componentType
}

// Config is the configuration for the logger.
type Config struct {
	Formatter      string            `json:"formatter,omitempty"`
	RequestHeaders map[string]string `json:"requestHeaders"`
}

const (
	// ContextKeyTraceID is the context key for the trace ID.
	ContextKeyTraceID = ContextKey("trace-id")
	// ContextKeyComponentType is the context key for the component type.
	ContextKeyComponentType = ContextKey("component-type")
	// Executor is the component type for the executor.
	Executor componentType = "executor"
	// Server is the component type for the Ratify http server.
	Server componentType = "server"
	// ReferrerStore is the component type for the referrer store.
	ReferrerStore componentType = "referrerStore"

	traceIDHeaderName = "traceIDHeaderName"
)

// InitLogger initializes the logger with the given configuration.
func InitLogger(ctx context.Context, r *http.Request, config Config) context.Context {
	return setTraceID(ctx, r, config.RequestHeaders)
}

// GetLogger returns a logger with provided values.
func GetLogger(ctx context.Context, opt Option) dcontext.Logger {
	ctx = context.WithValue(ctx, ContextKeyComponentType, opt.ComponentType)
	return dcontext.GetLogger(ctx, ContextKeyComponentType)
}

// setTraceID sets the trace ID in the context. If the trace ID is not present in the request headers, a new one is generated.
func setTraceID(ctx context.Context, r *http.Request, headers map[string]string) context.Context {
	traceID := ""
	if headers != nil {
		if _, ok := headers[traceIDHeaderName]; ok {
			label := headers[traceIDHeaderName]
			traceID = r.Header.Get(label)
		}
	}
	if traceID == "" {
		traceID = uuid.New().String()
	}
	ctx = context.WithValue(ctx, ContextKeyTraceID, traceID)
	return dcontext.WithLogger(ctx, dcontext.GetLogger(ctx, ContextKeyTraceID))
}

// SetFormatter sets the formatter for the logger.
func SetFormatter(formatter string) error {
	switch formatter {
	case "text", "":
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
			DisableQuote:    true,
		})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "logstash":
		logrus.SetFormatter(&logstash.LogstashFormatter{
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: time.RFC3339Nano,
			},
		})
	default:
		return re.ErrorCodeConfigInvalid.WithDetail(fmt.Sprintf("unsupported logging formatter: %s", formatter))
	}
	return nil
}
