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
	"fmt"
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
		setValue := attr.Value.AsString()
		if attr.Value.Type() == attribute.BOOL {
			setValue = fmt.Sprintf("%t", attr.Value.AsBool())
		}
		m.Attributes[string(attr.Key)] = setValue
	}
}

func (m *MockInt64Counter) Add(ctx context.Context, incr int64, attrs ...attribute.KeyValue) {
	m.Value = incr
	for _, attr := range attrs {
		setValue := attr.Value.AsString()
		if attr.Value.Type() == attribute.INT64 {
			setValue = fmt.Sprintf("%d", attr.Value.AsInt64())
		} else if attr.Value.Type() == attribute.BOOL {
			setValue = fmt.Sprintf("%t", attr.Value.AsBool())
		}
		m.Attributes[string(attr.Key)] = setValue
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
	ReportVerifierDuration(context.Background(), 5, "test_verifier", "test_subject", true, true)
	if mockDuration.Value != 5 {
		t.Fatalf("ReportVerifierDuration() mockDuration.Value = %v, expected %v", mockDuration.Value, 5)
	}
	if len(mockDuration.Attributes) != 4 {
		t.Fatalf("ReportVerifierDuration() len(mockDuration.Attributes) = %v, expected %v", len(mockDuration.Attributes), 2)
	}
	if mockDuration.Attributes["verifier"] != "test_verifier" {
		t.Fatalf("expected verifer attribute to be test_verifier but got %s", mockDuration.Attributes["verifier"])
	}
	if mockDuration.Attributes["subject"] != "test_subject" {
		t.Fatalf("expected subject attribute to be test_subject but got %s", mockDuration.Attributes["subject"])
	}
	if mockDuration.Attributes["error"] != "true" {
		t.Fatalf("expected error attribute to be true but got %s", mockDuration.Attributes["error"])
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

func TestReportRequestCount(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockCounter := &MockInt64Counter{Attributes: make(map[string]string)}
	registryRequestCount = mockCounter
	ReportRegistryRequestCount(context.Background(), 429, "test_registry_host")
	if mockCounter.Value != 1 {
		t.Fatalf("ReportRequestCount() mockCounter.Value = %v, expected %v", mockCounter.Value, 1)
	}
	if len(mockCounter.Attributes) != 2 {
		t.Fatalf("ReportRequestCount() len(mockCounter.Attributes) = %v, expected %v", len(mockCounter.Attributes), 2)
	}
	if mockCounter.Attributes["status_code"] != "429" {
		t.Fatalf("expected status_code attribute to be 429 but got %s", mockCounter.Attributes["status_code"])
	}
	if mockCounter.Attributes["registry_host"] != "test_registry_host" {
		t.Fatalf("expected registry_host attribute to be test_registry_host but got %s", mockCounter.Attributes["registry_host"])
	}
}

func TestReportAADExchangeDuration(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockDuration := &MockInt64Histogram{Attributes: make(map[string]string)}
	aadExchangeDuration = mockDuration
	ReportAADExchangeDuration(context.Background(), 500, "test_scope")
	if mockDuration.Value != 500 {
		t.Fatalf("ReportAADExchangeDuration() mockDuration.Value = %v, expected %v", mockDuration.Value, 500)
	}
	if len(mockDuration.Attributes) != 1 {
		t.Fatalf("ReportAADExchangeDuration() len(mockDuration.Attributes) = %v, expected %v", len(mockDuration.Attributes), 1)
	}
	if mockDuration.Attributes["resource_type"] != "test_scope" {
		t.Fatalf("expected resource_type attribute to be test_scope but got %s", mockDuration.Attributes["resource_type"])
	}
}

func TestReportACRExchangeDuration(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockDuration := &MockInt64Histogram{Attributes: make(map[string]string)}
	acrExchangeDuration = mockDuration
	ReportACRExchangeDuration(context.Background(), 500, "test_repo")
	if mockDuration.Value != 500 {
		t.Fatalf("ReportACRExchangeDuration() mockDuration.Value = %v, expected %v", mockDuration.Value, 500)
	}
	if len(mockDuration.Attributes) != 1 {
		t.Fatalf("ReportACRExchangeDuration() len(mockDuration.Attributes) = %v, expected %v", len(mockDuration.Attributes), 1)
	}
	if mockDuration.Attributes["repository"] != "test_repo" {
		t.Fatalf("expected repository attribute to be test_repo but got %s", mockDuration.Attributes["repository"])
	}
}

func TestReportAKVCertificateDuration(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockDuration := &MockInt64Histogram{Attributes: make(map[string]string)}
	akvCertificateDuration = mockDuration
	ReportAKVCertificateDuration(context.Background(), 500, "test_cert")
	if mockDuration.Value != 500 {
		t.Fatalf("ReportAKVCertificateDuration() mockDuration.Value = %v, expected %v", mockDuration.Value, 500)
	}
	if len(mockDuration.Attributes) != 1 {
		t.Fatalf("ReportAKVCertificateDuration() len(mockDuration.Attributes) = %v, expected %v", len(mockDuration.Attributes), 1)
	}
	if mockDuration.Attributes["certificate_name"] != "test_cert" {
		t.Fatalf("expected certificate_name attribute to be test_cert but got %s", mockDuration.Attributes["certificate_name"])
	}
}

func TestReportBlobCacheCount(t *testing.T) {
	if err := initStatsReporter(); err != nil {
		t.Fatalf("initStatsReporter() error = %v", err)
	}

	mockCounter := &MockInt64Counter{Attributes: make(map[string]string)}
	cacheBlobCount = mockCounter
	ReportBlobCacheCount(context.Background(), true)
	if mockCounter.Value != 1 {
		t.Fatalf("ReportBlobCacheCount() mockCounter.Value = %v, expected %v", mockCounter.Value, 1)
	}
	if len(mockCounter.Attributes) != 1 {
		t.Fatalf("ReportBlobCacheCount() len(mockCounter.Attributes) = %v, expected %v", len(mockCounter.Attributes), 1)
	}
	if mockCounter.Attributes["hit"] != "true" {
		t.Fatalf("expected hit attribute to be true but got %s", mockCounter.Attributes["hit"])
	}
}
