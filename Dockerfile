FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o textfile-exporter .

FROM alpine:latest

COPY --from=builder /app/textfile-exporter /textfile-exporter

EXPOSE 8080

ENTRYPOINT ["/textfile-exporter"]