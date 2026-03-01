.PHONY: build run test docs docs-check docs-install docs-generate-api docs-serve docs-build

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

docs-build: docs-generate-api
	$(MKDOCS) build --clean
