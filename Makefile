# Copyright 2021 The Perses Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
BINARY           ?= poc-cuelang
GO               ?= go
GOCI             ?= golangci-lint
GOFMT            ?= $(GO)fmt
GOARCH           ?= amd64
pkgs              = $$($(GO) list ./...)
COMMIT           := $(shell git status >/dev/null && git rev-parse HEAD)
BRANCH           := $(shell git status >/dev/null && git rev-parse --abbrev-ref HEAD)
DATE             := $(shell date +%Y-%m-%d)
COVERAGE_PROFILE := test-coverage-profile.out
PKG_LDFLAGS      := github.com/prometheus/common/version
LDFLAGS          := -ldflags "-X ${PKG_LDFLAGS}.Version=${VERSION} -X ${PKG_LDFLAGS}.Revision=${COMMIT} -X ${PKG_LDFLAGS}.BuildDate=${DATE} -X ${PKG_LDFLAGS}.Branch=${BRANCH}"

all: fmt build test

.PHONY: checkstyle
checkstyle:
	@echo ">> checking code style"
	$(GOCI) run --build-tags=integration --modules-download-mode=vendor -E goconst -E unconvert -E gosec -E revive -E unparam -E govet -E gocyclo -E unused

.PHONY: checkformat
checkformat:
	@echo ">> checking code format"
	! $(GOFMT) -d $$(find . -name '*.go' -print) | grep '^' ;\

.PHONY: fmt
fmt:
	@echo ">> format code"
	$(GO) fmt $(pkgs)

.PHONY: test
test:
	@echo ">> running all tests"
	$(GO) test -count=1 -v $(pkgs)

.PHONY: build
build:
	@echo ">> build the binary"
	CGO_ENABLED=0 GOARCH=${GOARCH} $(GO) build ${LDFLAGS} -o ./bin/${BINARY}
