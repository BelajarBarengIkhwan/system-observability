package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func newHTTPExporter(url string, ctx context.Context) (exp trace.SpanExporter) {
	insecure := otlptracehttp.WithInsecure()
	endpoint := otlptracehttp.WithEndpoint(url)
	exp, err := otlptracehttp.New(ctx, insecure, endpoint)
	if err != nil {
		panic(err)
	}
	return
}

func newTraceProvider(exp trace.SpanExporter, serviceName string) *trace.TracerProvider {
	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	if err != nil {
		panic(err)
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource),
	)
}

func NewHTTPTelemetryProvider(url, serviceName string, ctx context.Context) *trace.TracerProvider {
	exp := newHTTPExporter(url, ctx)
	return newTraceProvider(exp, serviceName)
}

func NewTelemetryPropagators() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}
