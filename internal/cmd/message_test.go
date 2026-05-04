package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMessageSendListPull(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	messagePath := filepath.Join(home, "ready.md")
	if err := os.WriteFile(messagePath, []byte("Status: READY\nProceed with execution."), 0o644); err != nil {
		t.Fatalf("write message file: %v", err)
	}

	if _, err := captureStdout(func() error {
		return RunMessage([]string{"send", "--to", "codex", "--from", "operator", "--channel", "execution", "--file", messagePath, "--title", "Execution Ready"})
	}); err != nil {
		t.Fatalf("RunMessage send: %v", err)
	}

	listOutput, err := captureStdout(func() error {
		return RunMessage([]string{"list", "--to", "codex", "--channel", "execution"})
	})
	if err != nil {
		t.Fatalf("RunMessage list: %v", err)
	}
	if !strings.Contains(listOutput, "Execution Ready") || !strings.Contains(listOutput, "channel=execution") {
		t.Fatalf("unexpected message list output: %q", listOutput)
	}

	pullOutput, err := captureStdout(func() error {
		return RunMessage([]string{"pull", "--to", "codex", "--channel", "execution"})
	})
	if err != nil {
		t.Fatalf("RunMessage pull: %v", err)
	}
	if !strings.Contains(pullOutput, "# Messages") || !strings.Contains(pullOutput, "Status: READY") || !strings.Contains(pullOutput, "operator -> codex") {
		t.Fatalf("unexpected message pull output: %q", pullOutput)
	}
}

func TestRunMessagePullNoMessages(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withCwd(t, home)
	writeTestConfig(t, home)

	output, err := captureStdout(func() error {
		return RunMessage([]string{"pull", "--to", "codex", "--channel", "execution"})
	})
	if err != nil {
		t.Fatalf("RunMessage pull: %v", err)
	}
	if !strings.Contains(output, "No messages found.") {
		t.Fatalf("expected no messages output: %q", output)
	}
}
