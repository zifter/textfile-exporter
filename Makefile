BINARY     := textfile-exporter
IMAGE      := textfile-exporter
METRICS    := metrics.txt

.PHONY: all build run clean test check vet fmt docker-build docker-run

all: build

build:
	go build -o $(BINARY) .

run: build
	METRICS_FILE_PATH=$(METRICS) REFRESH_INTERVAL=5s ./$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test -v ./...

check: vet fmt test
	go mod verify

vet:
	go vet ./...

fmt:
	@gofmt -l . | grep -q . && (echo "run: gofmt -w ." && exit 1) || true

docker-build:
	docker build -t $(IMAGE) .

docker-run: docker-build
	docker run --rm \
		-v $(PWD)/$(METRICS):/metrics.txt \
		-e METRICS_FILE_PATH=/metrics.txt \
		-e REFRESH_INTERVAL=5s \
		-p 8080:8080 \
		$(IMAGE)
