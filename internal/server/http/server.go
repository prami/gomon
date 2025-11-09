package http

import (
	"context"
	"net/http"
)

func NewServer(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux)

	return mux
}
