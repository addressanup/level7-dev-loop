SHELL := /bin/sh

override PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GO_VERSION ?= $(shell sed -n '1p' "$(PROJECT_ROOT)/.go-version")
override GO_ROOT := $(PROJECT_ROOT)/.cache/toolchains/go$(GO_VERSION)
override GO := $(GO_ROOT)/bin/go
override GOFMT := $(GO_ROOT)/bin/gofmt
override TELEMETRY_DIR := $(PROJECT_ROOT)/.cache/go/telemetry

override UNAME_S := $(shell uname -s)
override UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
override HOST_GOOS := darwin
else ifeq ($(UNAME_S),Linux)
override HOST_GOOS := linux
else
$(error unsupported harness operating system: $(UNAME_S))
endif
ifneq ($(filter arm64 aarch64,$(UNAME_M)),)
override HOST_GOARCH := arm64
else ifneq ($(filter x86_64 amd64,$(UNAME_M)),)
override HOST_GOARCH := amd64
else
$(error unsupported harness architecture: $(UNAME_M))
endif

override L7_EXPECT_GO_VERSION := go$(GO_VERSION)
override L7_LOG_FORMAT := json
override L7_LOG_LEVEL := INFO
override L7_TELEMETRY := off
override L7_NETWORK := off
override CORE_MODULE_PATH := $(strip $(shell awk -F '\t' '$$1 == "core" && $$2 == "active" { print $$4 }' "$(PROJECT_ROOT)/harness/modules.lock.tsv"))
ifneq ($(words $(CORE_MODULE_PATH)),1)
$(error harness/modules.lock.tsv must contain exactly one active core module)
endif
override HARNESS_IMPORT_PATH := $(CORE_MODULE_PATH)/internal/harness
override HARNESS_IDENTITY_LDFLAGS := -X $(HARNESS_IMPORT_PATH).expectedGoVersion=$(L7_EXPECT_GO_VERSION) -X $(HARNESS_IMPORT_PATH).expectedGOOS=$(HOST_GOOS) -X $(HARNESS_IMPORT_PATH).expectedGOARCH=$(HOST_GOARCH)
override L7_CLI_VERSION := 0.1.0-dev
override CLI_PACKAGE := ./cmd/l7
override CLI_LDFLAGS := -buildid= -X main.version=$(L7_CLI_VERSION)

override GOENV := off
override GOTOOLCHAIN := local
override GOWORK := off
override GO111MODULE := on
override GOFLAGS :=
override GOROOT := $(GO_ROOT)
override CGO_ENABLED := 0
override GOOS := $(HOST_GOOS)
override GOARCH := $(HOST_GOARCH)
override GOEXPERIMENT :=
override GOFIPS140 := off
ifeq ($(HOST_GOARCH),amd64)
override GOAMD64 := v1
else ifeq ($(HOST_GOARCH),arm64)
override GOARM64 := v8.0
endif
override GOPATH := $(PROJECT_ROOT)/.cache/go/path
override GOBIN := $(PROJECT_ROOT)/.cache/go/bin
override GOCACHE := $(PROJECT_ROOT)/.cache/go/build
override GOMODCACHE := $(PROJECT_ROOT)/.cache/go/mod
override GOTMPDIR := $(PROJECT_ROOT)/.cache/go/tmp
override TMPDIR := $(GOTMPDIR)
override GOPROXY := off
override GOSUMDB := off
override GOPRIVATE :=
override GONOPROXY :=
override GONOSUMDB :=
override GOINSECURE :=
override GOVCS := *:off
override GOAUTH := off
override TEST_TELEMETRY_DIR := $(TELEMETRY_DIR)
override GIT_TERMINAL_PROMPT := 0
override LC_ALL := C
override TZ := UTC

export GOENV GOTOOLCHAIN GOWORK GO111MODULE GOFLAGS GOROOT CGO_ENABLED GOOS GOARCH GOEXPERIMENT GOFIPS140
export GOAMD64 GOARM64 GOPATH GOBIN GOCACHE GOMODCACHE GOTMPDIR TMPDIR GOPROXY GOSUMDB GOPRIVATE GONOPROXY
export GONOSUMDB GOINSECURE GOVCS GOAUTH TEST_TELEMETRY_DIR GIT_TERMINAL_PROMPT LC_ALL TZ
export L7_EXPECT_GO_VERSION L7_LOG_FORMAT L7_LOG_LEVEL L7_TELEMETRY L7_NETWORK

.PHONY: bootstrap prepare toolchain-check install build cli-build cli-cross-build cli-benchmark-check cli-actual-host-compile distribution distribution-check build-control-check policy-check ready-check l7-import-closure-check import-check candidate-check format-check technical-lint lint typecheck test reproducible cli-reproducible technical-verify verify ci

bootstrap:
	@./scripts/harness/bootstrap-go.sh "$(GO_VERSION)"

prepare:
	@"$(PROJECT_ROOT)/scripts/harness/prepare-cache.sh" "$(PROJECT_ROOT)" "$(GO_VERSION)"

toolchain-check: prepare
	@test -x "$(GO)" || { echo "missing pinned Go $(GO_VERSION); run: make bootstrap GO_VERSION=$(GO_VERSION)" >&2; exit 1; }
	@test "$$("$(GO)" env GOVERSION)" = "$(L7_EXPECT_GO_VERSION)"
	@test "$$("$(GO)" env GOROOT)" = "$(GO_ROOT)"
	@test "$$("$(GO)" env GOTOOLDIR)" = "$(GO_ROOT)/pkg/tool/$(HOST_GOOS)_$(HOST_GOARCH)"
	@test "$$("$(GO)" env GOOS)" = "$(HOST_GOOS)"
	@test "$$("$(GO)" env GOARCH)" = "$(HOST_GOARCH)"
	@test "$$("$(GO)" env GOHOSTOS)" = "$(HOST_GOOS)"
	@test "$$("$(GO)" env GOHOSTARCH)" = "$(HOST_GOARCH)"
	@test "$$("$(GO)" env GOTOOLCHAIN)" = "local"
	@test "$$("$(GO)" env GOWORK)" = "off"
	@test -z "$$("$(GO)" env GOFLAGS)"
	@test "$$("$(GO)" env CGO_ENABLED)" = "0"
	@test "$$("$(GO)" env GOPROXY)" = "off"
	@test "$$("$(GO)" env GOSUMDB)" = "off"
	@test "$$("$(GO)" env GOVCS)" = "*:off"
	@test "$$("$(GO)" env GOAUTH)" = "off"
	@test -z "$$("$(GO)" env GOEXPERIMENT)"
	@test "$$("$(GO)" env GOFIPS140)" = "off"
	@test "$$("$(GO)" env GOTELEMETRY)" = "off"
	@test "$$("$(GO)" env GOTELEMETRYDIR)" = "$(TELEMETRY_DIR)"
	@test "$$("$(GO)" env GOCACHE)" = "$(GOCACHE)"
	@test "$$("$(GO)" env GOMODCACHE)" = "$(GOMODCACHE)"
	@test "$$("$(GO)" env GOPATH)" = "$(GOPATH)"
	@test "$$("$(GO)" env GOTMPDIR)" = "$(GOTMPDIR)"
	@test "$$TMPDIR" = "$(GOTMPDIR)"
	@grep -Fq 'TEST_TELEMETRY_DIR' "$(GO_ROOT)/src/cmd/internal/telemetry/telemetry.go"
	@grep -Fq 'TEST_TELEMETRY_DIR' "$(GO_ROOT)/src/cmd/internal/telemetry/counter/counter.go"

# The Wave 1 CLI intentionally has no third-party production dependencies.
install: toolchain-check
	@"$(GO)" mod download
	@"$(GO)" mod verify
	@"$(GO)" mod tidy -diff

build: cli-build

cli-build: install
	@mkdir -p "$(PROJECT_ROOT)/build/bin"
	@"$(GO)" build -mod=readonly -trimpath -buildvcs=false -ldflags='$(CLI_LDFLAGS)' -o "$(PROJECT_ROOT)/build/bin/l7" $(CLI_PACKAGE)

cli-cross-build: install
	@mkdir -p "$(PROJECT_ROOT)/build/bin"
	@GOOS=darwin GOARCH=arm64 GOARM64=v8.0 "$(GO)" build -mod=readonly -trimpath -buildvcs=false -ldflags='$(CLI_LDFLAGS)' -o "$(PROJECT_ROOT)/build/bin/l7-darwin-arm64" $(CLI_PACKAGE)
	@GOOS=darwin GOARCH=amd64 GOAMD64=v1 "$(GO)" build -mod=readonly -trimpath -buildvcs=false -ldflags='$(CLI_LDFLAGS)' -o "$(PROJECT_ROOT)/build/bin/l7-darwin-amd64" $(CLI_PACKAGE)

cli-benchmark-check: toolchain-check
	@test -n "$(L7_BENCHMARK_BASE_ROOT)" || { echo 'L7_BENCHMARK_BASE_ROOT must name a separate base checkout' >&2; exit 1; }
	@./scripts/harness/check-cli-benchmarks.sh "$(GO)" "$(L7_BENCHMARK_BASE_ROOT)" "$(PROJECT_ROOT)"

distribution: install
	@"$(GO)" run -mod=readonly ./internal/harness/distribution --root "$(PROJECT_ROOT)" --output "$(PROJECT_ROOT)/build/distributions"

distribution-check: install
	@./scripts/harness/check-distribution.sh "$(GO)"

build-control-check: toolchain-check
	@"$(GO)" run -mod=readonly ./internal/harness/buildcontrol

policy-check: build-control-check

ready-check: toolchain-check
	@"$(GO)" run -mod=readonly ./internal/harness/buildcontrol --require-ready

l7-import-closure-check: toolchain-check
	@set -eu; \
	 validate_imports() { \
	   label=$$1; \
	   imports=$$2; \
	   allowed=$$3; \
	   for imported in $$imports; do \
	     accepted=false; \
	     for allowed_import in $$allowed; do \
	       if test "$$imported" = "$$allowed_import"; then accepted=true; break; fi; \
	     done; \
	     if test "$$accepted" != true; then \
	       printf 'l7-import-closure-check: %s imports non-allowlisted %s\n' "$$label" "$$imported" >&2; \
	       return 1; \
	     fi; \
	   done; \
	 }; \
	 domain_allowed=''; \
	 app_allowed='context $(CORE_MODULE_PATH)/internal/l7/domain strings unicode/utf8'; \
	 presentation_allowed='bytes encoding/json fmt $(CORE_MODULE_PATH)/internal/l7/domain strconv'; \
	 for target in '$(HOST_GOOS)/$(HOST_GOARCH)' 'darwin/arm64' 'darwin/amd64'; do \
	   goos=$${target%/*}; \
	   goarch=$${target#*/}; \
	   domain_imports=$$(GOOS="$$goos" GOARCH="$$goarch" GOARM64=v8.0 GOAMD64=v1 "$(GO)" list -mod=readonly -f '{{join .Imports " "}}' ./internal/l7/domain); \
	   app_imports=$$(GOOS="$$goos" GOARCH="$$goarch" GOARM64=v8.0 GOAMD64=v1 "$(GO)" list -mod=readonly -f '{{join .Imports " "}}' ./internal/l7/app); \
	   presentation_imports=$$(GOOS="$$goos" GOARCH="$$goarch" GOARM64=v8.0 GOAMD64=v1 "$(GO)" list -mod=readonly -f '{{join .Imports " "}}' ./internal/l7/presentation); \
	   validate_imports "internal/l7/domain ($$target)" "$$domain_imports" "$$domain_allowed"; \
	   validate_imports "internal/l7/app ($$target)" "$$app_imports" "$$app_allowed"; \
	   validate_imports "internal/l7/presentation ($$target)" "$$presentation_imports" "$$presentation_allowed"; \
	 done; \
	 if validate_imports 'non-domain repository probe' "$$app_allowed $(CORE_MODULE_PATH)/internal/evaluator" "$$app_allowed" >/dev/null 2>&1; then \
	   printf 'l7-import-closure-check: accepted non-domain repository probe\n' >&2; \
	   exit 1; \
	 fi; \
	 if validate_imports 'indirect filesystem probe' "$$app_allowed io/ioutil" "$$app_allowed" >/dev/null 2>&1; then \
	   printf 'l7-import-closure-check: accepted indirect filesystem probe\n' >&2; \
	   exit 1; \
	 fi; \
	 printf 'l7-import-closure-check: PASS (3 package closures; host plus darwin/arm64 and darwin/amd64; 2 negative probes)\n'

import-check: toolchain-check l7-import-closure-check
	@./scripts/harness/check-import-boundaries.sh "$(GO)"

candidate-check: policy-check import-check

format-check: toolchain-check
	@unformatted="$$(find . -type f -name '*.go' -not -path './.cache/*' -not -path './build/*' -print0 | xargs -0 "$(GOFMT)" -l)"; \
	 test -z "$$unformatted" || { printf 'unformatted Go files:\n%s\n' "$$unformatted" >&2; exit 1; }

technical-lint: install import-check format-check
	@for script in scripts/harness/*.sh; do sh -n "$$script"; done
	@"$(GO)" vet -mod=readonly ./...

lint: policy-check technical-lint

typecheck: install
	@"$(GO)" test -mod=readonly -trimpath -buildvcs=false -ldflags='$(HARNESS_IDENTITY_LDFLAGS)' -run '^$$' -count=1 ./...

cli-actual-host-compile: install
	@"$(GO)" test -mod=readonly -trimpath -buildvcs=false -tags l7_actual_provider -run '^$$' -count=1 ./...

test: install
	@"$(GO)" test -mod=readonly -trimpath -buildvcs=false -ldflags='$(HARNESS_IDENTITY_LDFLAGS)' -count=1 -shuffle=off -timeout=2m ./...

reproducible: install
	@set -eu; \
	 repro_root="$$(mktemp -d "$(PROJECT_ROOT)/.cache/repro/go$(GO_VERSION).XXXXXX")"; \
	 mkdir "$$repro_root/a-cache" "$$repro_root/b-cache"; \
	 GOCACHE="$$repro_root/a-cache" "$(GO)" test -mod=readonly -c -trimpath -buildvcs=false -ldflags='-buildid= $(HARNESS_IDENTITY_LDFLAGS)' -o "$$repro_root/harness-a.test" ./internal/harness; \
	 GOCACHE="$$repro_root/b-cache" "$(GO)" test -mod=readonly -c -trimpath -buildvcs=false -ldflags='-buildid= $(HARNESS_IDENTITY_LDFLAGS)' -o "$$repro_root/harness-b.test" ./internal/harness; \
	 cmp "$$repro_root/harness-a.test" "$$repro_root/harness-b.test"; \
	 if command -v sha256sum >/dev/null 2>&1; then sha256sum "$$repro_root/harness-a.test"; else shasum -a 256 "$$repro_root/harness-a.test"; fi

cli-reproducible: install
	@set -eu; \
	 repro_root="$$(mktemp -d "$(PROJECT_ROOT)/.cache/repro/cli-go$(GO_VERSION).XXXXXX")"; \
	 mkdir "$$repro_root/a-cache" "$$repro_root/b-cache"; \
	 GOCACHE="$$repro_root/a-cache" "$(GO)" build -mod=readonly -trimpath -buildvcs=false -ldflags='$(CLI_LDFLAGS)' -o "$$repro_root/l7-a" $(CLI_PACKAGE); \
	 GOCACHE="$$repro_root/b-cache" "$(GO)" build -mod=readonly -trimpath -buildvcs=false -ldflags='$(CLI_LDFLAGS)' -o "$$repro_root/l7-b" $(CLI_PACKAGE); \
	 cmp "$$repro_root/l7-a" "$$repro_root/l7-b"; \
	 if command -v sha256sum >/dev/null 2>&1; then sha256sum "$$repro_root/l7-a"; else shasum -a 256 "$$repro_root/l7-a"; fi

technical-verify: install technical-lint typecheck cli-actual-host-compile test reproducible cli-reproducible distribution-check

verify: policy-check technical-verify

ci: technical-verify
