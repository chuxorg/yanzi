.PHONY: build run test docs docs-check docs-install docs-generate-api docs-serve docs-build docs-local

DOCGEN=go run github.com/princjef/gomarkdoc/cmd/gomarkdoc@v1.1.0
PIP?=pip3
MKDOCS?=mkdocs
PYTHON?=python3
ARGS?=--help
DOCS_HOST?=127.0.0.1
DOCS_PORT?=8000
DOCS_ADDR?=$(DOCS_HOST):$(DOCS_PORT)


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
	$(MKDOCS) serve -a $(DOCS_ADDR)

docs-local: docs-install docs-generate-api
	@PORT=`DOCS_HOST=$(DOCS_HOST) DOCS_PORT=$(DOCS_PORT) $(PYTHON) -c 'exec("""import os\nimport socket\nhost = os.environ[\"DOCS_HOST\"]\nstart = int(os.environ[\"DOCS_PORT\"])\nfor port in range(start, start + 20):\n    with socket.socket() as sock:\n        try:\n            sock.bind((host, port))\n        except OSError:\n            continue\n    print(port)\n    break\nelse:\n    raise SystemExit(\"no free docs port found\")\n""")'` && \
	echo "Serving docs at http://$(DOCS_HOST):$$PORT" && \
	$(MKDOCS) serve -a $(DOCS_HOST):$$PORT

docs-build: docs-generate-api
	$(MKDOCS) build --clean
