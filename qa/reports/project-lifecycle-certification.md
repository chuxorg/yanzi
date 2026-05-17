## Run: 2026-05-16T20:09:36Z

- Scenario: project-lifecycle deterministic operational certification
- Environment: Darwin arm64
- Repository: /Users/developer/projects/chuxorg/yanzi-cli-qa
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/expected
  - Normalized: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/normalized
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: No Drift
- Result: PASS
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

```text
No drift detected.
```

---

## Run: 2026-05-16T20:09:42Z

- Scenario: project-lifecycle deterministic operational certification
- Environment: Darwin arm64
- Repository: /Users/developer/projects/chuxorg/yanzi-cli-qa
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/expected
  - Normalized: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/normalized
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: No Drift
- Result: PASS
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

```text
No drift detected.
```

---

## Run: 2026-05-17T17:02:44Z

- Scenario: project-lifecycle deterministic operational certification
- Environment: Darwin arm64
- Repository: /Users/developer/projects/chuxorg/yanzi-cli-qa
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/expected
  - Normalized: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/normalized
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: No Drift
- Result: PASS
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

```text
No drift detected.
```

---

## Run: 2026-05-17T17:19:27Z

- Scenario: project-lifecycle deterministic operational certification
- Environment: Darwin arm64
- Repository: /Users/developer/projects/chuxorg/yanzi-cli-qa
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/expected
  - Normalized: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/normalized
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: No Drift
- Result: PASS
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

```text
No drift detected.
```

---

## Run: 2026-05-17T18:03:01Z

- Scenario: project-lifecycle deterministic operational certification
- Environment: Darwin arm64
- Repository: /Users/developer/projects/chuxorg/yanzi-cli-qa
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/validate-release-convergence.sh (when candidate tag is provided)
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/expected
  - Normalized: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/normalized
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: No Drift
- Distribution Convergence Status: FAIL
- Promotable Candidate: no
- Candidate Tag: v2.9.1-rc1
- Candidate SHA: bceb106b0fa97d6574fa1aa5d419f489f3e935c4
- Installer Runtime Version: 
- Homebrew Lineage: tap:chuxorg/yanzi/yanzi
- Result: FAIL
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

```text
No drift detected.
```

### Convergence Findings

```text
candidate_tag=v2.9.1-rc1
candidate_sha=bceb106b0fa97d6574fa1aa5d419f489f3e935c4
FAIL installer: pinned install failed for v2.9.1-rc1
uninstall_check=pass
FAIL reinstall failed
homebrew_formula_version="stable": "2.7.0"
FAIL homebrew lineage mismatch: expected 2.9.1-rc1 got "stable": "2.7.0"
status=FAIL
promotable=no
status_file=/Users/developer/projects/chuxorg/yanzi-cli-qa/qa/reports/release-convergence-status.env
```

---

## Run: 2026-05-17T18:03:34Z

- Scenario: project-lifecycle deterministic operational certification
- Environment: Darwin arm64
- Repository: /Users/developer/projects/chuxorg/yanzi-cli-qa
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/validate-release-convergence.sh (when candidate tag is provided)
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/expected
  - Normalized: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/normalized
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: No Drift
- Distribution Convergence Status: FAIL
- Promotable Candidate: no
- Candidate Tag: v2.9.1-rc1
- Candidate SHA: bceb106b0fa97d6574fa1aa5d419f489f3e935c4
- Installer Runtime Version: 
- Homebrew Lineage: tap:chuxorg/yanzi/yanzi
- Result: FAIL
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

```text
No drift detected.
```

### Convergence Findings

```text
candidate_tag=v2.9.1-rc1
candidate_sha=bceb106b0fa97d6574fa1aa5d419f489f3e935c4
FAIL installer: pinned install failed for v2.9.1-rc1
uninstall_check=pass
FAIL reinstall failed
homebrew_formula_version=2.7.0
FAIL homebrew lineage mismatch: expected 2.9.1-rc1 got 2.7.0
status=FAIL
promotable=no
status_file=/Users/developer/projects/chuxorg/yanzi-cli-qa/qa/reports/release-convergence-status.env
```

---

## Run: 2026-05-17T18:50:44Z

- Scenario: project-lifecycle deterministic operational certification
- Environment: Darwin arm64
- Repository: /Users/developer/projects/chuxorg/yanzi-cli-qa
- Commands Executed:
  - qa/execution/run-project-lifecycle.sh
  - qa/execution/normalize-output.sh
  - qa/execution/compare-snapshots.sh
  - qa/execution/validate-release-convergence.sh (when candidate tag is provided)
  - qa/execution/generate-report.sh
- Snapshots Certified:
  - Expected: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/expected
  - Normalized: /Users/developer/projects/chuxorg/yanzi-cli-qa/qa/snapshots/project-lifecycle/normalized
- Normalization Applied:
  - timestamp tokens (<TIMESTAMP>)
  - generated identifier tokens (<ID32>, <ID64>, <ULID>)
  - machine/path tokens (<REPO_PATH>, <HOME_PATH>, <WORKSPACE_PATH>, <TMP_PATH>)
- Drift Findings Classification: No Drift
- Distribution Convergence Status: PASS
- Promotable Candidate: yes
- Candidate Tag: v2.9.1-rc1
- Candidate SHA: 688d3f372b2e2f9a644e70bc5bb602dd54758cb6
- Installer Runtime Version: yanzi v2.9.1-rc1
- Homebrew Lineage: tap:chuxorg/yanzi/yanzi
- Result: PASS
- Certification Notes: First human-reviewed deterministic baseline established for project-lifecycle scenario.

### Drift Findings

```text
No drift detected.
```

### Convergence Findings

```text
candidate_tag=v2.9.1-rc1
candidate_sha=688d3f372b2e2f9a644e70bc5bb602dd54758cb6
installer_version=yanzi v2.9.1-rc1
uninstall_check=pass
reinstall_version=yanzi v2.9.1-rc1
homebrew_formula_version=2.9.1-rc1
status=PASS
promotable=yes
status_file=/Users/developer/projects/chuxorg/yanzi-cli-qa/qa/reports/release-convergence-status.env
```

---

