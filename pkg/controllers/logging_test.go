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

package controllers

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestNewLogSink(t *testing.T) {
	logger := logrus.New()
	sink := NewLogrusSink(logger)

	if sink.logger != logger {
		t.Errorf("expected logger to be %v, got %v", logger, sink.logger)
	}
}

func TestInit(_ *testing.T) {
	logger := logrus.New()
	sink := NewLogrusSink(logger)
	sink.Init(logr.RuntimeInfo{})
}

func TestEnabled_UsesInfoLevel(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	sink := NewLogrusSink(logger)

	if !sink.Enabled(0) {
		t.Errorf("expected enabled to be true, got false")
	}
}

func TestEnabled_UsesDebugLevel(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	sink := NewLogrusSink(logger)

	if !sink.Enabled(1) {
		t.Errorf("expected enabled to be true, got false")
	}
}

func TestInfo(t *testing.T) {
	logger, hook := test.NewNullLogger()
	sink := NewLogrusSink(logger)

	sink.Info(0, "test", "property1", 100, "property2", "value2")

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.LastEntry().Message)
	assert.Equal(t, 100, hook.LastEntry().Data["property1"])
	assert.Equal(t, "value2", hook.LastEntry().Data["property2"])
}

func TestError(t *testing.T) {
	logger, hook := test.NewNullLogger()
	sink := NewLogrusSink(logger)

	err := fmt.Errorf("myerror")
	sink.Error(err, "test", "property1", 100, "property2", "value2")

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.LastEntry().Message)
	assert.Equal(t, 100, hook.LastEntry().Data["property1"])
	assert.Equal(t, "value2", hook.LastEntry().Data["property2"])
}

func TestError_WithDebug(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	sink := NewLogrusSink(logger)

	err := fmt.Errorf("myerror")
	sink.Error(err, "test", "property1", 100, "property2", "value2")

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.NotNil(t, hook.LastEntry().Data["stacktrace"])
}

func TestWithValues(t *testing.T) {
	logger, hook := test.NewNullLogger()
	sink := NewLogrusSink(logger)

	sink.WithValues("property1", 100, "property2", "value2").Info(0, "test")

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.LastEntry().Message)
	assert.Equal(t, 100, hook.LastEntry().Data["property1"])
	assert.Equal(t, "value2", hook.LastEntry().Data["property2"])
}

func TestWithName(t *testing.T) {
	logger, hook := test.NewNullLogger()
	sink := NewLogrusSink(logger)

	sink.WithName("myname").WithName("subname").Info(0, "test")

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, "[myname.subname] test", hook.LastEntry().Message)
}

func TestError_WithName(t *testing.T) {
	logger, hook := test.NewNullLogger()
	sink := NewLogrusSink(logger)

	err := fmt.Errorf("myerror")
	sink.WithName("myname").WithName("subname").Error(err, "test")

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "[myname.subname] test", hook.LastEntry().Message)
}

func TestError_WithValues(t *testing.T) {
	logger, hook := test.NewNullLogger()
	sink := NewLogrusSink(logger)

	err := fmt.Errorf("myerror")
	sink.WithValues("property1", 100, "property2", "value2").Error(err, "test")

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.LastEntry().Message)
	assert.Equal(t, 100, hook.LastEntry().Data["property1"])
	assert.Equal(t, "value2", hook.LastEntry().Data["property2"])
}
