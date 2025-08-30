.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -o . ./cmd/...

.PHONY: update
update:
	go get -u -t ./...
	go mod tidy
	go mod vendor

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0
	golangci-lint run