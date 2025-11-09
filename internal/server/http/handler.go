package http

import (
	"fmt"
	"net/http"
)

func handleRoot() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "hello .. 👋👋👋")
		})
}
