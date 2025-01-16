package common

import (
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Metrics struct {
	Resquests metric.Int64Counter
	Seconds   metric.Float64Histogram
}

func NewMetrics(appName string) (*Metrics, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	meter := provider.Meter(appName)

	requst, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		return nil, err
	}

	seconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		return nil, err
	}
	return &Metrics{
		Resquests: requst,
		Seconds:   seconds,
	}, nil
}
