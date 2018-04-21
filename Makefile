BUILD_DATE := `date -u +%Y%m%d`
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.0.1)
GIT_SHA := $(shell git rev-parse HEAD)
APP_NAME := discordbot
PROJECT := github.com/gsmcwhirter/eso-discord

# can specify V=1 on the line with `make` to get verbose output
V ?= 0
Q = $(if $(filter 1,$V),,@)

# Entrypoints:
# 	- `make` (show help)
# 	- `make all` (deps + debug)
# 	- `make deps` (get dependencies)
# 	- `make` / `make debug` (build debug)
#   - `make test` (run tests)
#	- `make release` (build release)
#	- `make release-upload` (to build release and upload at once)
#	- `make release && make upload-release-bundles` (to build and upload separately)

.DEFAULT_GOAL := help

all: deps debug  ## Download dependencies and do a debug build

build-debug: version proto
	$Q go build -v -ldflags "-X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) -race $(PROJECT)/cmd/$(APP_NAME)

build-release: version proto
	$Q go build -v -ldflags "-s -w -X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) $(PROJECT)/cmd/$(APP_NAME)

proto: pkg/storage/storage.go  ## Compile the protobuf files

pkg/storage/storage.go: pkg/storage/storage.proto
	protoc --go_out=pkg/storage --proto_path=pkg/storage pkg/storage/storage.proto

build-release-bundles: build-release
	$Q gzip -k -f bin/$(APP_NAME)
	$Q cp bin/$(APP_NAME).gz bin/$(APP_NAME)-$(VERSION).gz

clean:  ## Remove compiled artifacts
	$Q rm bin/*

debug: fmt vet build-debug  ## Debug build: create a dev build (enable race detection, don't strip symbols)

release: fmt vet test build-release-bundles  ## Release build: create a release build (disable race detection, strip symbols)

deps:  ## Download dependencies
	$Q # for development and linting
	$Q go get golang.org/x/tools/cmd/godoc
	$Q go get golang.org/x/tools/cmd/goimports
	$Q go get -u github.com/alecthomas/gometalinter
	# $Q gometalinter --install
	$Q go get -u github.com/coreos/bbolt/..
	$Q go get -u github.com/golang/protobuf/protoc-gen-go
	$Q go get -u github.com/gorilla/websocket

fmt:
	$Q for src in $(shell goimports -l .); \
		do \
			goimports -w $$src; \
		done

test:  ## Run the tests
	$Q go test -cover ./pkg/...

version:  ## Print the version string and git sha that would be recorded if a release was built now
	$Q echo $(VERSION)
	$Q echo $(GIT_SHA)

vet:  ## Run the linter
	$Q gometalinter -e S1008 $(GOPATH)/src/$(PROJECT)/cmd/...
	$Q gometalinter -e S1008 $(GOPATH)/src/$(PROJECT)/pkg/...

help:  ## Show the help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ./Makefile
