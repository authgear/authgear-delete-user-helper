.PHONY: start
start:
	go run ./cmd/helper start

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.50.1
	go mod download
	go install github.com/google/wire/cmd/wire

.PHONY: generate
generate:
	go generate ./pkg/... ./cmd/...
