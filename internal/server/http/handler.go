package http

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/metric"
)

func handleIndex(meter metric.Meter) http.Handler {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "hello ... 👋👋👋")
		},
	)
}
