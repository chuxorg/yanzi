#!/usr/bin/env bash
set -euo pipefail

go run ./test/qa/docvalidate --root . "$@"
