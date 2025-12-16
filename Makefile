.PHONY: build test fmt vet lint clean

build:
	go build -o bin/yaml-versioner ./cmd/yaml-versioner

build-image:
	ko

test:
	go test -v -race -coverprofile=coverage.out ./...

coverage: test
	go tool cover -html=coverage.out

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/ coverage.out
