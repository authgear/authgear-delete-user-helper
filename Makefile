GIT_HASH ?= git-$(shell git rev-parse --short=12 HEAD)

.PHONY: start
start:
	go run ./cmd/helper start

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.50.1
	go mod download
	go install github.com/google/wire/cmd/wire@v0.5.0

.PHONY: generate
generate:
	go generate ./pkg/... ./cmd/...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build:
	go build -o authgear-delete-user-helper ./cmd/helper

.PHONY: check-tidy
check-tidy:
	$(MAKE) fmt
	$(MAKE) generate
	go mod tidy
	git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: build-image
build-image:
	# Add --pull so that we are using the latest base image.
	docker build --pull --file ./cmd/helper/Dockerfile --tag quay.io/theauthgear/authgear-delete-user-helper:$(GIT_HASH) --build-arg GIT_HASH=$(GIT_HASH) .

.PHONY: push-image
push-image:
	docker push quay.io/theauthgear/authgear-delete-user-helper:$(GIT_HASH)
