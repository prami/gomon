ARG GO_VERSION=1.25.3
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Cache deps
FROM golang:${GO_VERSION} AS deps
WORKDIR /src
# Cache go mod download
COPY go.mod go.sum ./
RUN go mod download

# Dev hotreload (air)
FROM golang:${GO_VERSION} AS dev
WORKDIR /app

RUN go install github.com/air-verse/air@latest
# Copy deps
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY --from=deps /go/bin /go/bin

# Build
FROM golang:${GO_VERSION} AS build
WORKDIR /src

# ARG do wstrzyknięcia metadanych w binarkę
ARG VERSION=dev
ARG COMMIT=none
ARG BUILDDATE=unknown

# Copy deps
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY --from=deps /go/bin /go/bin

# Copy dir
COPY . .

# Build app
ENV CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath \
      -ldflags="-s -w \
        -X main.version=${VERSION} \
        -X main.commit=${COMMIT} \
        -X main.buildDate=${BUILDDATE}" \
      -o /out/gomon ./cmd/gomon

# Final (?)
FROM gcr.io/distroless/static-debian12:nonroot AS final
WORKDIR /app
COPY --from=build /out/gomon /app/gomon

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/gomon"]