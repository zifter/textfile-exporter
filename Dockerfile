FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

RUN GOOS=linux GOARCH=amd64 go build .

FROM alpine:latest

COPY --from=builder /app/metrics.txt /metrics.txt
COPY --from=builder /app/textfile-exporter /textfile-exporter

EXPOSE 8080

ENTRYPOINT ["/textfile-exporter"]