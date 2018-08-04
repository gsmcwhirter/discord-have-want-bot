BUILD_DATE := `date -u +%Y%m%d`
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.0.1)
GIT_SHA := $(shell git rev-parse HEAD)
APP_NAME := have-want-bot
REPL_NAME := botrepl
PROJECT := github.com/gsmcwhirter/eso-discord
SERVER := discordbot@evogames.org:~/eso-discord/

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

build-debug: version generate
	$Q go build -v -ldflags "-X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) -race $(PROJECT)/cmd/$(APP_NAME)
	$Q go build -v -ldflags "-X main.AppName=$(REPL_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(REPL_NAME) -race $(PROJECT)/cmd/$(REPL_NAME)

build-release: version generate
	$Q GOOS=linux go build -v -ldflags "-s -w -X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) $(PROJECT)/cmd/$(APP_NAME)
	$Q GOOS=linux go build -v -ldflags "-s -w -X main.AppName=$(REPL_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(REPL_NAME) $(PROJECT)/cmd/$(REPL_NAME)

generate:
	$Q go generate ./...

build-release-bundles: build-release
	$Q gzip -k -f bin/$(APP_NAME)
	$Q cp bin/$(APP_NAME).gz bin/$(APP_NAME)-$(VERSION).gz

clean:  ## Remove compiled artifacts
	$Q rm bin/*

debug: generate vet build-debug  ## Debug build: create a dev build (enable race detection, don't strip symbols)

release: generate test build-release-bundles  ## Release build: create a release build (disable race detection, strip symbols)

deps:  ## Download dependencies
	$Q go get ./...
	$Q # for development and linting
	$Q go get -u github.com/alecthomas/gometalinter
	$Q # $Q gometalinter --install

test:  ## Run the tests
	$Q go test -cover ./pkg/...

version:  ## Print the version string and git sha that would be recorded if a release was built now
	$Q echo $(VERSION) $(GIT_SHA)

vet:  ## Run the linter
	$Q gometalinter -e S1008 --disable=gocyclo --disable=megacheck --disable=gas --disable=goconst --deadline 120s -s jsonapi $(GOPATH)/src/$(PROJECT)/cmd/...
	$Q gometalinter -e S1008 --disable=gocyclo --disable=megacheck --disable=gas --disable=goconst --deadline 120s -s jsonapi $(GOPATH)/src/$(PROJECT)/pkg/...

release-upload: release upload

upload:
	$Q scp ./bin/$(REPL_NAME) ./bin/$(APP_NAME).gz ./config.toml ./eso-have-want-bot.service ./install.sh $(SERVER)
	$Q scp ./bin/$(APP_NAME)-$(VERSION).gz $(SERVER)

help:  ## Show the help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ./Makefile
