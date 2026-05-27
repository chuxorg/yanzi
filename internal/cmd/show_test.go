package cmd

import (
	"strings"
	"testing"
)

func TestRunShowDisplaysFullIntentDetails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	record := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "Artifact detail",
		Prompt:     "show prompt",
		Response:   "show response",
		Meta:       []byte(`{"project":"alpha","profile":"engineer","custom":"x"}`),
	})

	output, err := captureStdout(func() error {
		return RunShow([]string{record.ID})
	})
	if err != nil {
		t.Fatalf("RunShow: %v", err)
	}

	expectedSubstrings := []string{
		"ID: " + record.ID,
		"Author: alice",
		"Source: cli",
		"Title: Artifact detail",
		"Prev_Hash: ",
		"Hash: " + record.Hash,
		"--- Prompt ---\nshow prompt",
		"--- Response ---\nshow response",
	}
	for _, substring := range expectedSubstrings {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output: %q", substring, output)
		}
	}
	for _, metaPart := range []string{`Meta: {`, `"project":"alpha"`, `"profile":"engineer"`, `"custom":"x"`} {
		if !strings.Contains(output, metaPart) {
			t.Fatalf("expected %q in output: %q", metaPart, output)
		}
	}
}

func TestRunShowMissingIntentReturnsCurrentError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	err := RunShow([]string{"missing-intent"})
	if err == nil {
		t.Fatal("expected missing intent error")
	}
	if got := err.Error(); got != "intent not found for ID missing-intent" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestRunShowIncludesDeletedIntentByID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	record := createTestIntentRecord(t, createIntentInput{
		Author:     "alice",
		SourceType: "cli",
		Title:      "Deleted detail",
		Prompt:     "deleted prompt",
		Response:   "deleted response",
	})

	if err := RunDelete([]string{record.ID}); err != nil {
		t.Fatalf("RunDelete: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunShow([]string{record.ID})
	})
	if err != nil {
		t.Fatalf("RunShow deleted: %v", err)
	}
	if !strings.Contains(output, "Title: Deleted detail") {
		t.Fatalf("expected deleted record in show output: %q", output)
	}
}
