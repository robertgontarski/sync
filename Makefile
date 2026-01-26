.PHONY: build build-docker build-all test clean

BINARY_NAME=sync
IMAGE_NAME=sync
PLATFORMS=linux/amd64,linux/arm64,darwin/amd64,darwin/arm64,windows/amd64

build:
	go build -o $(BINARY_NAME) ./cmd/sync

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

build-docker:
	docker build -t $(IMAGE_NAME):latest .

build-all:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):latest .

build-push:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):latest --push .

dist:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/sync
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/sync
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/sync
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/sync
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/sync
