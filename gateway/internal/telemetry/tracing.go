package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Tracing struct {
	provider *sdktrace.TracerProvider
	enabled  bool
}

func NewTracing(
	ctx context.Context,
) (*Tracing, error) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	if os.Getenv(
		"SENTRYMESH_TRACING_ENABLED",
	) != "1" {
		return &Tracing{}, nil
	}

	serviceName := os.Getenv(
		"OTEL_SERVICE_NAME",
	)

	if serviceName == "" {
		serviceName =
			"sentrymesh-gateway"
	}

	customResource :=
		resource.NewSchemaless(
			attribute.String(
				"service.name",
				serviceName,
			),
			attribute.String(
				"service.version",
				"0.1.0",
			),
		)

	res, err :=
		resource.Merge(
			resource.Default(),
			customResource,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"create tracing resource: %w",
			err,
		)
	}

	exporter, err :=
		otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"create OTLP trace exporter: %w",
			err,
		)
	}

	sampleRatio := 0.05

	if raw := os.Getenv(
		"SENTRYMESH_TRACE_SAMPLE_RATIO",
	); raw != "" {
		parsed, err := strconv.ParseFloat(
			raw,
			64,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"parse trace sample ratio: %w",
				err,
			)
		}

		if parsed < 0 || parsed > 1 {
			return nil, fmt.Errorf(
				"trace sample ratio must be between 0 and 1",
			)
		}

		sampleRatio = parsed
	}

	log.Printf(
		"OpenTelemetry trace sample ratio: %.4f",
		sampleRatio,
	)

	provider :=
		sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(
				exporter,
			),
			sdktrace.WithResource(
				res,
			),
			sdktrace.WithSampler(
				sdktrace.ParentBased(
					sdktrace.TraceIDRatioBased(
						sampleRatio,
					),
				),
			),
		)

	otel.SetTracerProvider(
		provider,
	)

	return &Tracing{
		provider: provider,
		enabled:  true,
	}, nil
}

func (t *Tracing) Enabled() bool {
	return t != nil &&
		t.enabled
}

func (t *Tracing) Shutdown(
	ctx context.Context,
) error {
	if t == nil ||
		t.provider == nil {
		return nil
	}

	return t.provider.Shutdown(
		ctx,
	)
}
