package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

func TestRunStatusTextIncludesContinuitySections(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedCheckpoint(t, db, "alpha", "2025-01-01T00:00:02Z", "checkpoint 1")
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "cap-1",
		CreatedAt: "2025-01-01T00:00:03Z",
		Project:   "alpha",
		Author:    "Ada",
		Title:     "Retry path",
		Prompt:    "Investigate retry path",
		Response:  "Need lock visibility",
	})
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "evt-1",
		CreatedAt: "2025-01-01T00:00:04Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "@yanzi pause",
		Response:  "true",
	})
	if _, err := yanzilibrary.CreateArtifact("alpha", yanzilibrary.ArtifactClassIntent, "task", "Open task", "Continue validation.", `{"status":"open","owner":"ops"}`); err != nil {
		t.Fatalf("CreateArtifact task: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunStatus([]string{"--recent", "4"})
	})
	if err != nil {
		t.Fatalf("RunStatus: %v", err)
	}

	for _, want := range []string{
		"Continuity Summary",
		"Mode: checkpoint",
		"Operational Metrics",
		"Protocol annotations: 0",
		"Recent Activity",
		"capture by Ada: Retry path",
		"Unresolved Work",
		"task: Open task",
		"owner=ops status=open",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
}

func TestRunStatusJSONIncludesSummaryAndOpenWork(t *testing.T) {
	workdir := t.TempDir()
	t.Setenv("HOME", workdir)
	withCwd(t, workdir)
	writeTestConfig(t, workdir)
	writeStateFile(t, workdir, "alpha")
	db := openTestDB(t, workdir)
	defer db.Close()

	seedProject(t, db, "alpha")
	seedIntentRecord(t, db, rehydrateSeedIntent{
		ID:        "cap-1",
		CreatedAt: "2025-01-01T00:00:01Z",
		Project:   "alpha",
		Author:    "Ada",
		Prompt:    "Need continuity",
		Response:  "Use fallback",
	})
	if _, err := yanzilibrary.CreateArtifact("alpha", yanzilibrary.ArtifactClassIntent, "change_request", "Pending change", "Tighten export summary.", `{"status":"in_review"}`); err != nil {
		t.Fatalf("CreateArtifact change_request: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunStatus([]string{"--format", "json"})
	})
	if err != nil {
		t.Fatalf("RunStatus json: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, output)
	}
	if payload["schema_version"] != float64(machineContractSchemaVersion) {
		t.Fatalf("unexpected schema version: %#v", payload["schema_version"])
	}
	if payload["kind"] != jsonKindStatus {
		t.Fatalf("unexpected kind: %#v", payload["kind"])
	}
	if strings.Index(output, "\"schema_version\"") > strings.Index(output, "\"kind\"") || strings.Index(output, "\"kind\"") > strings.Index(output, "\"project\"") {
		t.Fatalf("unexpected field ordering: %s", output)
	}
	if payload["continuity_mode"] != "fallback" {
		t.Fatalf("unexpected continuity mode: %#v", payload["continuity_mode"])
	}
	if payload["total_captures"] != float64(1) {
		t.Fatalf("unexpected capture count: %#v", payload["total_captures"])
	}
	if payload["latest_checkpoint"] != nil {
		t.Fatalf("did not expect latest checkpoint: %#v", payload["latest_checkpoint"])
	}
	unresolved, ok := payload["unresolved_work"].([]any)
	if !ok || len(unresolved) != 1 {
		t.Fatalf("unexpected unresolved work payload: %#v", payload["unresolved_work"])
	}
	recent, ok := payload["recent_activity"].([]any)
	if !ok || len(recent) != 1 {
		t.Fatalf("unexpected recent activity payload: %#v", payload["recent_activity"])
	}
	if latest := strings.Index(output, "\"latest_checkpoint\""); latest != -1 {
		t.Fatalf("did not expect latest_checkpoint field in fallback output: %s", output)
	}
}
