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
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
)

type MockInt64Histogram struct {
	instrument.Int64Histogram
	Value      int64
	Attributes map[string]string
}

func (m *MockInt64Histogram) Record(ctx context.Context, incr int64, attrs ...attribute.KeyValue) {
	m.Value = incr
	for _, attr := range attrs {
		m.Attributes[string(attr.Key)] = attr.Value.AsString()
	}
}

func (m *MockInt64Counter) Add(ctx context.Context, incr int64, attrs ...attribute.KeyValue) {
	m.Value = incr
	for _, attr := range attrs {
		m.Attributes[string(attr.Key)] = attr.Value.AsString()
	}
}

type MockInt64Counter struct {
	instrument.Int64Counter
	Value      int64
	Attributes map[string]string
}

func TestReportVerificationRequest(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockDuration := &MockInt64Histogram{}
	verificationDuration = mockDuration
	ReportVerificationRequest(context.Background(), 5)
	if mockDuration.Value != 5 {
		t.Fatalf("ReportVerificationRequest() mockDuration.Value = %v, expected %v", mockDuration.Value, 5)
	}
}

func TestReportMutationRequest(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockDuration := &MockInt64Histogram{}
	mutationDuration = mockDuration
	ReportMutationRequest(context.Background(), 5)
	if mockDuration.Value != 5 {
		t.Fatalf("ReportMutationRequest() mockDuration.Value = %v, expected %v", mockDuration.Value, 5)
	}
}

func TestReportVerifierDuration(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockDuration := &MockInt64Histogram{Attributes: make(map[string]string)}
	verifierDuration = mockDuration
	ReportVerifierDuration(context.Background(), 5, "test_verifier", "test_subject")
	if mockDuration.Value != 5 {
		t.Fatalf("ReportVerifierDuration() mockDuration.Value = %v, expected %v", mockDuration.Value, 5)
	}
	if len(mockDuration.Attributes) != 2 {
		t.Fatalf("ReportVerifierDuration() len(mockDuration.Attributes) = %v, expected %v", len(mockDuration.Attributes), 2)
	}
	if mockDuration.Attributes["verifier"] != "test_verifier" {
		t.Fatalf("expected verifer attribute to be test_verifier but got %s", mockDuration.Attributes["verifier"])
	}
	if mockDuration.Attributes["subject"] != "test_subject" {
		t.Fatalf("expected subject attribute to be test_subject but got %s", mockDuration.Attributes["subject"])
	}
}

func TestReportSystemError(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockCounter := &MockInt64Counter{Attributes: make(map[string]string)}
	systemErrorCount = mockCounter
	ReportSystemError(context.Background(), "test_error")
	if mockCounter.Value != 1 {
		t.Fatalf("ReportSystemError() mockCounter.Value = %v, expected %v", mockCounter.Value, 1)
	}
	if len(mockCounter.Attributes) != 1 {
		t.Fatalf("ReportSystemError() len(mockCounter.Attributes) = %v, expected %v", len(mockCounter.Attributes), 1)
	}
	if mockCounter.Attributes["error"] != "test_error" {
		t.Fatalf("expected error attributes to be test_error but got %s", mockCounter.Attributes["error"])
	}
}
