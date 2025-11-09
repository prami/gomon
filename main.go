package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.22.0"
	"google.golang.org/grpc"
)

func setupMeter(ctx context.Context) (func(context.Context) error, metric.Meter) {
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("otel-collector:4317"),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		log.Fatal(err)
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("myservice"),
			attribute.String("env", "dev"),
		),
	)
	reader := sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(2*time.Second))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)
	return mp.Shutdown, mp.Meter("myservice")
}

func main() {
	ctx := context.Background()
	shutdown, meter := setupMeter(ctx)
	defer func() { _ = shutdown(ctx) }()

	reqs, _ := meter.Int64Counter("http_requests_total")
	lat, _ := meter.Float64Histogram("request_duration_seconds")

	// Start otel runtime metrics (goroutines, GC, mem, threads, etc.)
	_ = otelruntime.Start(
		otelruntime.WithMinimumReadMemStatsInterval(10*time.Second),
		otelruntime.WithMeterProvider(otel.GetMeterProvider()),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer lat.Record(r.Context(), time.Since(start).Seconds())
		reqs.Add(r.Context(), 1)
		fmt.Fprintln(w, "hello 👋👋👋")
	})

	log.Println("listening -g on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
