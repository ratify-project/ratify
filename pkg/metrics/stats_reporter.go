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

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
)

var (
	// defines the metrics that are collected
	verificationDuration instrument.Int64Histogram
	mutationDuration     instrument.Int64Histogram
	verifierDuration     instrument.Int64Histogram
	systemErrorCount     instrument.Int64Counter
	registryRequestCount instrument.Int64Counter

	// Azure Metrics
	aadExchangeDuration instrument.Int64Histogram
	acrExchangeDuration instrument.Int64Histogram
)

const (
	scope = "github.com/deislabs/ratify"

	// metric names
	metricNameVerificationDuration = "ratify_verification_request"
	metricNameMutationDuration     = "ratify_mutation_request"
	metricNameVerifierDuration     = "ratify_verifier_duration"
	metricNameSystemErrorCount     = "ratify_system_error_count"
	metricNameRegistryRequestCount = "ratify_registry_request_count"

	// Azure Metrics
	metricNameAADExchangeDuration = "ratify_aad_exchange_duration"
	metricNameACRExchangeDuration = "ratify_acr_exchange_duration"
)

// NewStatsReporter creates a new StatsReporter
func initStatsReporter() error {
	var err error
	// defines the data-stream transformation of histogram metrics
	// the boundaries of the histogram are based on historical data observed and attempt to minimize the error of quantile estimation
	// NOTE: https://prometheus.io/docs/practices/histograms/#errors-of-quantile-estimation
	views := []sdkmetric.View{
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name:  metricNameVerificationDuration,
				Scope: instrumentation.Scope{Name: scope},
			},
			sdkmetric.Stream{
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: []float64{0, 10, 30, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1400, 1600, 1800, 2000, 2300, 2600, 4000, 4400, 4900},
				},
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name:  metricNameMutationDuration,
				Scope: instrumentation.Scope{Name: scope},
			},
			sdkmetric.Stream{
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: []float64{0, 10, 30, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1400, 1600, 1800},
				},
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name:  metricNameVerifierDuration,
				Scope: instrumentation.Scope{Name: scope},
			},
			sdkmetric.Stream{
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: []float64{0, 10, 50, 100, 200, 300, 400, 600, 800, 1100, 1500, 2000},
				},
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name:  metricNameAADExchangeDuration,
				Scope: instrumentation.Scope{Name: scope},
			},
			sdkmetric.Stream{
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: []float64{0, 10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1200},
				},
			},
		),
		sdkmetric.NewView(
			sdkmetric.Instrument{
				Name:  metricNameACRExchangeDuration,
				Scope: instrumentation.Scope{Name: scope},
			},
			sdkmetric.Stream{
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: []float64{0, 10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1200},
				},
			},
		)}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(MetricReader), sdkmetric.WithView(views...))
	meter := provider.Meter(scope)
	verificationDuration, err = meter.Int64Histogram(metricNameVerificationDuration, instrument.WithUnit("millisecond"), instrument.WithDescription("verification request duration in ms"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	mutationDuration, err = meter.Int64Histogram(metricNameMutationDuration, instrument.WithUnit("millisecond"), instrument.WithDescription("mutation request duration in ms"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	verifierDuration, err = meter.Int64Histogram(metricNameVerifierDuration, instrument.WithUnit("millisecond"), instrument.WithDescription("verifier duration in ms"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	systemErrorCount, err = meter.Int64Counter(metricNameSystemErrorCount, instrument.WithDescription("system error count"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	registryRequestCount, err = meter.Int64Counter(metricNameRegistryRequestCount, instrument.WithDescription("registry request count"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	aadExchangeDuration, err = meter.Int64Histogram(metricNameAADExchangeDuration, instrument.WithUnit("millisecond"), instrument.WithDescription("AAD exchange duration in ms"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	acrExchangeDuration, err = meter.Int64Histogram(metricNameACRExchangeDuration, instrument.WithUnit("millisecond"), instrument.WithDescription("ACR exchange duration in ms"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func ReportVerificationRequest(ctx context.Context, duration int64) {
	if verificationDuration != nil {
		verificationDuration.Record(ctx, duration)
	}
}

func ReportMutationRequest(ctx context.Context, duration int64) {
	if mutationDuration != nil {
		mutationDuration.Record(ctx, duration)
	}
}

func ReportVerifierDuration(ctx context.Context, duration int64, veriferName string, subjectReference string, success bool, isError bool) {
	if verifierDuration != nil {
		verifierDuration.Record(ctx, duration,
			attribute.KeyValue{
				Key:   "verifier",
				Value: attribute.StringValue(veriferName),
			},
			attribute.KeyValue{
				Key:   "subject",
				Value: attribute.StringValue(subjectReference),
			},
			attribute.KeyValue{
				Key:   "success",
				Value: attribute.BoolValue(success),
			},
			attribute.KeyValue{
				Key:   "error",
				Value: attribute.BoolValue(isError),
			},
		)
	}
}

func ReportSystemError(ctx context.Context, errorString string) {
	if systemErrorCount != nil {
		systemErrorCount.Add(ctx, 1, attribute.KeyValue{Key: "error", Value: attribute.StringValue(errorString)})
	}
}

func ReportRegistryRequestCount(ctx context.Context, statusCode int, registryHost string) {
	if registryRequestCount != nil {
		registryRequestCount.Add(ctx, 1, attribute.KeyValue{Key: "status_code", Value: attribute.IntValue(statusCode)}, attribute.KeyValue{Key: "registry_host", Value: attribute.StringValue(registryHost)})
	}
}

func ReportAADExchangeDuration(ctx context.Context, duration int64, resourceType string) {
	if aadExchangeDuration != nil {
		aadExchangeDuration.Record(ctx, duration, attribute.KeyValue{Key: "resource_type", Value: attribute.StringValue(resourceType)})
	}
}

func ReportACRExchangeDuration(ctx context.Context, duration int64, repository string) {
	if acrExchangeDuration != nil {
		acrExchangeDuration.Record(ctx, duration, attribute.KeyValue{Key: "repository", Value: attribute.StringValue(repository)})
	}
}
