GO ?= go
GIT_COMMIT=$(shell git rev-parse HEAD 2> /dev/null || true)
GO_VERSION=$(shell $(GO) version | cut -d" " -f3)
DATE=$(shell date '+%d %h %Y %H:%M:%S')
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS ?= \
	-X kubegraph/pkg/cmd.Commit=$(GIT_COMMIT) \
	-X kubegraph/pkg/cmd.GoVersion=$(GO_VERSION) \
	-X 'kubegraph/pkg/cmd.Date=$(DATE)' \
	-X kubegraph/pkg/cmd.Branch=$(BRANCH)

.DEFAULT: build

build:
	$(GO) build -ldflags "$(LDFLAGS) -X 'kubegraph/pkg/cmd.OSArch=$(shell go env GOOS)/$(shell go env GOARCH)'" -o kubegraph cmd/kubegraph/main.go

build_linux:
	env GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -X kubegraph/pkg/cmd.OSArch=linux/amd64" -o kubegraph cmd/kubegraph/main.go

build_win:
	env GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -X kubegraph/pkg/cmd.OSArch=windows/amd64" -o kubegraph cmd/kubegraph/main.go

build_darwin:
	env GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -X kubegraph/pkg/cmd.OSArch=darwin/amd64" -o kubegraph cmd/kubegraph/main.go

install:
	install -D -m0755 kubegraph /usr/local/bin/kubegraph

test: unittest

unittest:
	$(GO) test -v -cover ./...

validate: gofmt lint

install_golanci_lint: 
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(shell go env GOPATH)/bin" v1.31.0

lint:
	golangci-lint run --timeout=5m0s

# https://github.com/containers/podman/blob/master/Makefile
gofmt:
	find . -name '*.go' -type f \
		-not \( \
			-name '.golangci.yml' -o \
			-name 'Makefile' -o \
			-path './vendor/*' -prune -o \
			-path './contrib/*' -prune \
		\) -exec gofmt -d -e -s -w {} \+
	git diff --exit-code

.PHONY: build build_linux build_darwin build_win test unittest validate install_golanci_lint lint gofmt install
