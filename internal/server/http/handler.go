package http

import (
	"fmt"
	rand "math/rand/v2"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func handleIndex(meter metric.Meter) http.Handler {
	reqs, _ := meter.Int64Counter("http.server.requests")
	lat, _ := meter.Float64Histogram("http.server.duration")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Sleep 0-50ms
			time.Sleep(time.Duration(rand.IntN(51)) * time.Millisecond)

			fmt.Fprintln(w, "hello ... 👋👋👋")

			d := time.Since(start).Seconds()
			lat.Record(r.Context(), d,
				metric.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.route", "/"), // lub pattern jeśli masz
				),
			)
			reqs.Add(r.Context(), 1,
				metric.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.route", "/"),
					attribute.Int("http.status_code", http.StatusOK),
				),
			)
		})
}
