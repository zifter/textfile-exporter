# textfile-exporter

A lightweight HTTP server that reads metrics from a text file and exposes them in [Prometheus exposition format](https://prometheus.io/docs/instrumenting/exposition_formats/).

Useful for exporting deployment metadata, environment variables, or any static/slowly-changing values to Prometheus without writing a full exporter.

## How it works

1. You write metrics to a file in Prometheus text format (e.g. from a deploy script or CI/CD pipeline)
2. `textfile-exporter` serves that file at `/metrics`
3. Prometheus scrapes the endpoint as usual

## Quick start

```bash
# Write metrics to a file
cat > metrics.txt << EOF
# HELP deploy_info Deployment metadata
# TYPE deploy_info gauge
deploy_info{version="1.2.3",env="prod",commit="abc123"} 1

# HELP deploy_timestamp Unix timestamp of last deployment
# TYPE deploy_timestamp gauge
deploy_timestamp 1713436800
EOF

# Run the exporter
./textfile-exporter
# or
docker run -v $(pwd)/metrics.txt:/metrics.txt -p 8080:8080 textfile-exporter
```

```
$ curl localhost:8080/metrics
# HELP deploy_info Deployment metadata
# TYPE deploy_info gauge
deploy_info{version="1.2.3",env="prod",commit="abc123"} 1
...
```

## Configuration

All settings are configured via environment variables:

| Variable | Default | Description |
|---|---|---|
| `SERVE_ADDR` | `:8080` | Address and port to listen on |
| `METRICS_FILE_PATH` | `metrics.txt` | Path to the metrics file |
| `METRICS_ENDPOINT` | `/metrics` | HTTP path to expose metrics at |
| `REFRESH_INTERVAL` | `0` | How often to reload the file (`0` = load once at startup). Accepts Go duration strings: `30s`, `1m`, `5m` |
| `LOG_OUTPUT` | `stdout` | Log destination (`stdout` or `stderr`) |

## Docker

```bash
docker build -t textfile-exporter .

# Mount your metrics file and run
docker run \
  -v /path/to/metrics.txt:/metrics.txt \
  -e METRICS_FILE_PATH=/metrics.txt \
  -e REFRESH_INTERVAL=30s \
  -p 8080:8080 \
  textfile-exporter
```

## Use case: deploy-time metrics

Generate `metrics.txt` as part of your deployment pipeline:

```bash
#!/bin/bash
cat > /var/metrics/deploy.txt << EOF
# HELP deploy_info Current deployment info
# TYPE deploy_info gauge
deploy_info{version="${APP_VERSION}",env="${ENVIRONMENT}",commit="${GIT_COMMIT}"} 1

# HELP deploy_timestamp Unix timestamp of last successful deployment
# TYPE deploy_timestamp gauge
deploy_timestamp $(date +%s)

# HELP build_number CI build number
# TYPE build_number gauge
build_number ${BUILD_NUMBER}
EOF
```

Run `textfile-exporter` with `REFRESH_INTERVAL=30s` so it picks up the new file after each deploy without restarting.

## Prometheus config

```yaml
scrape_configs:
  - job_name: deploy_metrics
    static_configs:
      - targets: ['localhost:8080']
```

## Metrics file format

Standard [Prometheus text format](https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format):

```
# HELP metric_name Description of the metric
# TYPE metric_name gauge
metric_name{label="value"} 42
metric_name_with_timestamp{label="value"} 42 1713436800000
```

Supported types: `gauge`, `counter`, `untyped`.

## Endpoints

| Path | Description |
|---|---|
| `/metrics` | Prometheus metrics (configurable via `METRICS_ENDPOINT`) |
| `/` | Health check — returns `200 OK` |
