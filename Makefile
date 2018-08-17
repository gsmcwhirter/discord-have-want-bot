BUILD_DATE := `date -u +%Y%m%d`
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.0.1)
GIT_SHA := $(shell git rev-parse HEAD)

APP_NAME := have-want-bot
REPL_NAME := have-want-repl
PROJECT := github.com/gsmcwhirter/discord-have-want-bot

SERVER := discordbot@evogames.org:~/eso-discord/
CONF_FILE := ./have-want-config.toml
SERVICE_FILE := ./eso-have-want-bot.service
INSTALLER := ./have-want-install.sh

# can specify V=1 on the line with `make` to get verbose output
V ?= 0
Q = $(if $(filter 1,$V),,@)

.DEFAULT_GOAL := help

build-debug: version generate
	$Q go build -v -ldflags "-X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) -race $(PROJECT)/cmd/$(APP_NAME)
	$Q go build -v -ldflags "-X main.AppName=$(REPL_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(REPL_NAME) -race $(PROJECT)/cmd/$(REPL_NAME)

build-release: version generate
	$Q GOOS=linux go build -v -ldflags "-s -w -X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) $(PROJECT)/cmd/$(APP_NAME)
	$Q GOOS=linux go build -v -ldflags "-s -w -X main.AppName=$(REPL_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(REPL_NAME) $(PROJECT)/cmd/$(REPL_NAME)

generate:  ## run a go generate
	$Q go generate ./...

build-release-bundles: build-release
	$Q gzip -k -f bin/$(APP_NAME)
	$Q cp bin/$(APP_NAME).gz bin/$(APP_NAME)-$(VERSION).gz
	$Q gzip -k -f bin/$(REPL_NAME)
	$Q cp bin/$(REPL_NAME).gz bin/$(REPL_NAME)-$(VERSION).gz

clean:  ## Remove compiled artifacts
	$Q rm bin/*

debug: generate test build-debug  ## Debug build: create a dev build (enable race detection, don't strip symbols)

release: generate test build-release-bundles  ## Release build: create a release build (disable race detection, strip symbols)

deps:  ## Download dependencies
	$Q go get ./...

test:  ## Run the tests
	$Q go test -cover ./...

version:  ## Print the version string and git sha that would be recorded if a release was built now
	$Q echo $(VERSION) $(GIT_SHA)

vet:  ## Run the linter
	$Q golint ./...
	$Q go vet ./...
	$Q gometalinter -D gas -D gocyclo -D goconst -e .pb.go -e _easyjson.go --warn-unmatched-nolint --enable-gc --deadline 180s ./...

release-upload: release upload

setup: deps generate  ## attempt to get everything set up to do a build (deps and generate)

upload:
	$Q scp $(CONF_FILE) $(SERVICE_FILE) $(INSTALLER) $(SERVER)
	$Q scp  ./bin/$(APP_NAME).gz ./bin/$(APP_NAME)-$(VERSION).gz $(SERVER)
	$Q scp ./bin/$(REPL_NAME).gz ./bin/$(REPL_NAME)-$(VERSION).gz $(SERVER)

help:  ## Show the help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ./Makefile
