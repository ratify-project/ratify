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

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

type componentType string

type Option struct {
	ComponentType componentType
}

type Config struct {
	Formatter      string            `json:"formatter,omitempty"`
	RequestHeaders map[string]string `json:"requestHeaders"`
}

const (
	ContextKeyTraceID                     = ContextKey("trace-id")
	ContextKeyComponentType               = ContextKey("component-type")
	Executor                componentType = "executor"
	Server                  componentType = "server"
	ReferrerStore           componentType = "referrerStore"

	traceIDHeaderName = "traceIDHeaderName"
)

func InitLogger(ctx context.Context, r *http.Request, config Config) context.Context {
	return setTraceID(ctx, r, config.RequestHeaders)
}

func GetLogger(ctx context.Context, opt Option) dcontext.Logger {
	ctx = context.WithValue(ctx, ContextKeyComponentType, opt.ComponentType)
	return dcontext.GetLogger(ctx, ContextKeyComponentType)
}

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

func SetFormatter(formatter string) error {
	switch formatter {
	case "text", "":
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
			DisableQuote:    true,
		})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "logstash":
		logrus.SetFormatter(&logstash.LogstashFormatter{
			Formatter: &logrus.JSONFormatter{},
		})
	default:
		err := re.ErrorCodeConfigInvalid.WithDetail(fmt.Sprintf("unsupported logging formatter: %s", formatter))
		logrus.Error(err)
		return err
	}
	return nil
}
