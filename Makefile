.DEFAULT_GOAL := run
.PHONY: help run build uninstall install install-prod debug debug-warn debug-info \
        profile profile-cpu profile-heap profile-allocs dlv logs \
        lint lint-fix test test-one test-rerun fmt check-nerd-font fix-nerd-font \
        docs-prepare docs docs-build docs-preview docs-sponsors

# Pass extra arguments via ARGS, e.g.: make run ARGS="--debug"
ARGS ?=

# Match Taskfile's `lint`/`lint:fix` which clear GOEXPERIMENT
LINT_ENV := GOEXPERIMENT=

help:
	@echo "Available targets:"
	@echo "  run               Run (default)"
	@echo "  build             go build ."
	@echo "  uninstall         Remove the gh dash extension"
	@echo "  install           Build and install local as gh extension"
	@echo "  install-prod      Install latest released version"
	@echo "  debug             Run with --debug; tail with 'make logs'"
	@echo "  debug-warn        Run with --debug at LOG_LEVEL=warn"
	@echo "  debug-info        Run with --debug at LOG_LEVEL=info"
	@echo "  profile           Run with profiling enabled"
	@echo "  profile-cpu       10s CPU profile"
	@echo "  profile-heap      Heap profile"
	@echo "  profile-allocs    Allocations profile"
	@echo "  dlv               Start headless dlv on 127.0.0.1:43000"
	@echo "  logs              Tail ./debug.log"
	@echo "  lint              Run golangci-lint"
	@echo "  lint-fix          Run golangci-lint --fix"
	@echo "  test              Run prism tests (./...)"
	@echo "  test-one          Run a single test (gotip)"
	@echo "  test-rerun        Rerun the last test"
	@echo "  fmt               gofumpt -w on tracked Go files"
	@echo "  check-nerd-font   Find broken nerd-font characters"
	@echo "  fix-nerd-font     Fix broken nerd-font icons"
	@echo "  docs              Start docs dev server"
	@echo "  docs-build        Build docs production bundle"
	@echo "  docs-preview      Preview docs production build"
	@echo "  docs-sponsors     Update sponsors list"

run:
	go run . $(ARGS)

build:
	go build .

uninstall:
	gh ext remove dash

install: build
	-gh ext remove dash
	@echo "🕐 Installing local build..."
	gh ext install .
	gh dash --version

install-prod:
	-gh ext remove dash
	@echo "🕐 Installing latest version..."
	gh ext install dlvhdr/gh-dash
	gh dash --version

debug:
	@printf "~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n―――――――――――――――――――――――――――――――――――――――――――――――\n" > ./debug.log
	LOG_LEVEL=debug DEBUG=true go run . --debug $(ARGS)

debug-warn:
	@printf "~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n―――――――――――――――――――――――――――――――――――――――――――――――\n" > ./debug.log
	LOG_LEVEL=warn DEBUG=true go run . --debug $(ARGS)

debug-info:
	@printf "~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n―――――――――――――――――――――――――――――――――――――――――――――――\n" > ./debug.log
	LOG_LEVEL=info DEBUG=true go run . --debug $(ARGS)

profile:
	@printf "~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n~\n―――――――――――――――――――――――――――――――――――――――――――――――\n" > ./debug.log
	DASH_PROFILE=true LOG_LEVEL=debug DEBUG=true go run . --debug $(ARGS)

profile-cpu:
	go tool pprof -http :6061 'http://localhost:6060/debug/pprof/profile?seconds=10'

profile-heap:
	go tool pprof -http :6061 'http://localhost:6060/debug/pprof/heap'

profile-allocs:
	go tool pprof -http :6061 'http://localhost:6060/debug/pprof/allocs'

dlv:
	dlv debug --headless --api-version=2 --listen=127.0.0.1:43000 .

logs:
	rm -f ./debug.log
	touch ./debug.log
	tail -f ./debug.log

lint:
	$(LINT_ENV) golangci-lint run --path-mode=abs --config=".golangci.yml" --timeout=5m

lint-fix:
	$(LINT_ENV) golangci-lint run --path-mode=abs --config=".golangci.yml" --timeout=5m --fix

test:
	prism test $(ARGS) ./...

test-one:
	gotip

test-rerun:
	gotip --rerun

fmt:
	gofumpt -w $$(git ls-files '*.go')

check-nerd-font:
	nerdfix check $$(fd --extension go)

fix-nerd-font:
	nerdfix fix --format=json $$(fd --extension go)

docs-prepare:
	cd docs && pnpm i

docs:
	cd docs && pnpm dev $(ARGS)

docs-build: docs-prepare
	cd docs && pnpm build

docs-preview: docs-build
	cd docs && pnpm preview

docs-sponsors:
	which sponsors || go install github.com/goreleaser/sponsors@main
	sponsors generate --config gh://dlvhdr/sponsors/sponsors.yml sponsors.json
	sponsors apply sponsors.json gh://dlvhdr/sponsors/readme.tpl.md ./README.md
