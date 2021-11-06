VERSION := $(shell git describe --tags)
GIT_HASH := $(shell git rev-parse --short HEAD )
#GOPATH=$(shell pwd)/vendor:$(shell pwd)

GO_VERSION        ?= $(shell go version)
GO_VERSION_NUMBER ?= $(word 3, $(GO_VERSION))
LDFLAGS = -ldflags "-X main.Version=${VERSION} -X main.GitHash=${GIT_HASH} -X main.GoVersion=${GO_VERSION_NUMBER}"

.PHONY: build
build:
	go build ${LDFLAGS} -v -o target/zconsole-exporter .

.PHONY: build-release
build-release: build-release-amd64 

.PHONY: build-release-amd64
build-release-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o=zconsole-exporter.linux.amd64 .      


.PHONY: test
test:
	go test -v -race ./...

.PHONY: get-dependencies
get-dependencies:
	go get -v -t -d ./...

.PHONY: vet
vet:
	go vet ./...

test-output:
	$(shell echo $$GO_VERSION_NUMBER)

.PHONY: fmt-fix
fmt-fix:
	goimports -w -l .

.PHONY: setup-env
setup-env:
	go get golang.org/x/tools/cmd/goimports
	go install golang.org/x/tools/cmd/goimports