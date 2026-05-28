.PHONY: build run test docs docs-check docs-install docs-generate-api docs-serve docs-build docs-local

DOCGEN=go run github.com/princjef/gomarkdoc/cmd/gomarkdoc@v1.1.0
PIP?=pip3
MKDOCS?=mkdocs


build:
	@mkdir -p bin
	go build -o bin/yanzi ./cmd/yanzi

run:
	go run ./cmd/yanzi $(ARGS)

test:
	go test ./...

# Generated API doc targets (gomarkdoc):
#   docs             — regenerate docs/API.md (combined reference)
#   docs-check       — verify docs/API.md is up to date (CI-safe, no write)
#   docs-generate-api — regenerate docs/API.md + docs/api/cmd.md + docs/api/internal.md
#
# Hand-maintained files that must NOT be regenerated:
#   docs/api/index.md — REST API reference for /v0 HTTP endpoints (CAP-002)
#   docs/cli.md       — CLI reference (22-command surface)
#
# Regeneration cadence: run `make docs-generate-api` before each release
# or whenever cmd/ or internal/ package exports change (new commands, new types).
docs:
	$(DOCGEN) -o docs/API.md ./cmd/yanzi ./internal/...

docs-check:
	$(DOCGEN) --check -o docs/API.md ./cmd/yanzi ./internal/...

docs-install:
	$(PIP) install -r docs/requirements.txt

docs-generate-api:
	@command -v gomarkdoc >/dev/null 2>&1 || go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
	./scripts/generate_api_docs.sh

docs-serve: docs-generate-api
	$(MKDOCS) serve

docs-local: docs-install docs-generate-api
	$(MKDOCS) serve

docs-build: docs-generate-api
	$(MKDOCS) build --clean
