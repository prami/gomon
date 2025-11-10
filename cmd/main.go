package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	httpsrv "github.com/prami/gomon/internal/server/http"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc"
)

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	mp, shutdownOTel, err := initTelemetry(ctx, "otel-collector:4317", "gomon", "dev")
	if err != nil {
		return fmt.Errorf("otel init: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = shutdownOTel(ctx)
	}()

	// Start otel runtime metrics (goroutines, GC, mem, threads, etc.)
	err = otelruntime.Start(
		otelruntime.WithMinimumReadMemStatsInterval(10*time.Second),
		otelruntime.WithMeterProvider(mp),
	)
	if err != nil {
		return fmt.Errorf("runtime metrics: %w", err)
	}

	srv := httpsrv.NewServer(
		ctx,
		mp.Meter("gomon/http"),
	)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", "8080"),
		Handler: srv,
	}

	// Run the server in a goroutine so that it can be gracefully shut down
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	// Wait for an interrupt signal to gracefully shut down the server
	var wg sync.WaitGroup
	wg.Go(func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	})
	wg.Wait()

	return nil
}

func initTelemetry(ctx context.Context, endpoint, service, env string) (
	metric.MeterProvider, func(context.Context) error, error,
) {

	dctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exp, err := otlpmetricgrpc.New(dctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			semconv.DeploymentEnvironmentName(env),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	reader := sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(10*time.Second))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)

	shutdown := func(ctx context.Context) error {
		return mp.Shutdown(ctx)
	}
	return mp, shutdown, nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
