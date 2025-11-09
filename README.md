# gomon

Small Go service wired with OpenTelemetry metrics, an OTLP collector, and a Prometheus/Grafana front-end. Handy for verifying instrumentation patterns or trying out dashboard ideas before rolling them into real systems.

## Ingredients
- Go HTTP handler exposes one endpoint, adds custom counters/histograms, and streams runtime stats via `otelruntime` (`main.go`).
- OTLP gRPC exporter feeds the collector, which batches and re-exports as a Prometheus target (`otel-collector.yaml`).
- Prometheus scrapes the collector (`prometheus.yml`) and Grafana provisions that datasource and optional dashboards on startup.
- Docker Compose connects the services, runs the app through Air for hot reloads, and stores Prometheus/Grafana data on volumes.

## Quick start
```bash
docker compose up --build
```

When everything is running:
- App: http://localhost:8888 — each request increases `http_requests_total` and records latency.
- Prometheus: http://localhost:9090 — browse the collector-exported series (runtime metrics included).
- Grafana: http://localhost:3030 (admin/admin) — dashboards pick up the Prometheus datasource automatically.

## Reference points
- `main.go` – instrumentation setup, runtime metrics, graceful shutdown.
- `otel-collector.yaml` – OTLP receiver + Prometheus exporter.
- `prometheus.yml` – scrape configuration.
- `docker-compose.yml` – service topology and env wiring.
