package http

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/metric"
)

func NewServer(
	ctx context.Context,
	meter metric.Meter,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		meter,
	)

	return mux
}
