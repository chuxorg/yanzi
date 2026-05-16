#!/usr/bin/env bats

setup() {
  ROOT_DIR="$(cd "${BATS_TEST_DIRNAME}/../../.." && pwd)"
  export PATH="$ROOT_DIR/qa/vendor/bin:$PATH"
  load "$ROOT_DIR/qa/vendor/bats-support/load.bash"
  load "$ROOT_DIR/qa/vendor/bats-assert/load.bash"
}

@test "yanzi --help returns success and command usage" {
  run "$ROOT_DIR/yanzi" --help

  assert_success
  assert_output --partial "usage:"
  assert_output --partial "yanzi"
  refute_output --partial "panic"
  refute_output --partial "segmentation fault"
}
