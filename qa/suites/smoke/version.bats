#!/usr/bin/env bats

setup() {
  ROOT_DIR="$(cd "${BATS_TEST_DIRNAME}/../../.." && pwd)"
  export PATH="$ROOT_DIR/qa/vendor/bin:$PATH"
  load "$ROOT_DIR/qa/vendor/bats-support/load.bash"
  load "$ROOT_DIR/qa/vendor/bats-assert/load.bash"
}

@test "yanzi --version returns success and expected marker" {
  run "$ROOT_DIR/yanzi" --version

  assert_success
  assert_output --partial "yanzi"
  refute_output --partial "panic"
  refute_output --partial "segmentation fault"
}
