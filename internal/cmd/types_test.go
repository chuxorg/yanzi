package cmd

import (
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
}
