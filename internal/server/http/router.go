package http

import (
	"net/http"

	"go.opentelemetry.io/otel/metric"
)

func addRoutes(
	mux *http.ServeMux,
	meter metric.Meter,
) {
	mux.Handle("/", handleIndex(meter))
}
