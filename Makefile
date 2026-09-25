BINARY := gateway-monitor
BIN_DIR := bin
VERSION ?= $(shell git describe --tags --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/wustus/gateway-monitor/internal/version/version.Version=$(VERSION)

.PHONY: build build-linux test vet clean docker help

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) .

build-linux:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)

docker:
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY):$(VERSION) .
