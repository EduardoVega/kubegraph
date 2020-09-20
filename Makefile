GO ?= go
GIT_COMMIT=$(shell git rev-parse HEAD 2> /dev/null || true)
GO_VERSION=$(shell $(GO) version | cut -d" " -f3)
OS_ARCH=$(shell $(GO) env GOARCH)
DATE=$(shell date '+%d %h %Y %H:%M:%S')

.DEFAULT: build

build: version build_linux clean_version

build_linux:
	$(GO) build -o kubegraph cmd/kubegraph/main.go

install:
	install -D -m0755 kubegraph /usr/local/bin/kubegraph

version:
	mv pkg/cmd/version.go pkg/cmd/.version.go
	sed -e "s/^const Commit.*/const Commit = \"$(GIT_COMMIT)\"/g" \
		-e "s/^const GoVersion.*/const GoVersion = \"$(GO_VERSION)\"/g" \
		-e "s/^const OSArch.*/const OSArch = \"$(OS_ARCH)\"/g" \
		-e "s/^const Date.*/const Date = \"$(DATE)\"/g" pkg/cmd/.version.go > pkg/cmd/version.go

clean_version:
	mv pkg/cmd/.version.go pkg/cmd/version.go

test: unittest

unittest:
	$(GO) test -v -cover ./...

lint:
	$(GO) vet ./...
	$(GO) fmt ./...