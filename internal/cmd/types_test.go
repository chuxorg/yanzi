package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRunTypesListJSON(t *testing.T) {
	output, err := captureStdout(func() error {
		return RunTypes([]string{"list", "--json"})
	})
	if err != nil {
		t.Fatalf("RunTypes list: %v", err)
	}
	if !strings.Contains(output, "\"context\"") || !strings.Contains(output, "\"governance\": \"process_rule\"") {
		t.Fatalf("unexpected types json output: %q", output)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode types json: %v", err)
	}
	if payload["schema_version"] != float64(machineContractSchemaVersion) {
		t.Fatalf("unexpected schema version: %#v", payload["schema_version"])
	}
	if payload["kind"] != jsonKindArtifactTypes {
		t.Fatalf("unexpected kind: %#v", payload["kind"])
	}
	if strings.Index(output, "\"schema_version\"") > strings.Index(output, "\"kind\"") || strings.Index(output, "\"kind\"") > strings.Index(output, "\"intent\"") {
		t.Fatalf("unexpected field ordering: %s", output)
	}
}
