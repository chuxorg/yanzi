package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/chuxorg/yanzi/internal/config"
)

func TestIsHelpArg(t *testing.T) {
	cases := map[string]bool{
		"-h":     true,
		"--help": true,
		"?":      true,
		"help":   false,
		"":       false,
	}
	for input, want := range cases {
		if got := isHelpArg(input); got != want {
			t.Fatalf("isHelpArg(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestFormatMode(t *testing.T) {
	if got := formatMode(config.Config{Mode: config.ModeLocal}); got != "local" {
		t.Fatalf("expected local, got %q", got)
	}
	cfg := config.Config{Mode: config.ModeHTTP, BaseURL: "https://example.com"}
	if got := formatMode(cfg); got != "http (https://example.com)" {
		t.Fatalf("expected http mode, got %q", got)
	}
}

func TestUsagePrintsHelp(t *testing.T) {
	output := captureStderr(t, func() {
		usage()
	})
	if !strings.Contains(output, "usage:") {
		t.Fatalf("expected usage header, got: %s", output)
	}
	if !strings.Contains(output, "capture  Create a new intent") {
		t.Fatalf("expected command description, got: %s", output)
	}
	if !strings.Contains(output, "export  Export active project history.") {
		t.Fatalf("expected export command description, got: %s", output)
	}
	if !strings.Contains(output, "init  Create or bind a project to the current directory.") {
		t.Fatalf("expected init command description, got: %s", output)
	}
	if !strings.Contains(output, "intent  Manage intent artifacts.") {
		t.Fatalf("expected intent command description, got: %s", output)
	}
	if !strings.Contains(output, "context  Manage context artifacts.") {
		t.Fatalf("expected context command description, got: %s", output)
	}
	if !strings.Contains(output, "pack  Apply or export portable context packs.") {
		t.Fatalf("expected pack command description, got: %s", output)
	}
	if !strings.Contains(output, "bootstrap  Load ordered context documents") {
		t.Fatalf("expected bootstrap command description, got: %s", output)
	}
	if !strings.Contains(output, "rules  Manage rule metadata wrappers.") {
		t.Fatalf("expected rules command description, got: %s", output)
	}
	if !strings.Contains(output, "types  List canonical artifact types and aliases.") {
		t.Fatalf("expected types command description, got: %s", output)
	}
	if !strings.Contains(output, "message  Manage thin message wrappers.") {
		t.Fatalf("expected message command description, got: %s", output)
	}
	if !strings.Contains(output, "--profile <name>") {
		t.Fatalf("expected profile help text, got: %s", output)
	}
	if !strings.Contains(output, "--meta key=value") {
		t.Fatalf("expected capture metadata help text, got: %s", output)
	}
	if !strings.Contains(output, "Prompt may also be piped on stdin.") {
		t.Fatalf("expected stdin capture help text, got: %s", output)
	}
	if !strings.Contains(output, "--format claude-context") {
		t.Fatalf("expected claude-context help text, got: %s", output)
	}
	if !strings.Contains(output, "list and checkpoint list use tab-separated columns for deterministic parsing.") {
		t.Fatalf("expected parsing stability note, got: %s", output)
	}
	if !strings.Contains(output, "rehydrate --format json is the machine-readable continuity output.") {
		t.Fatalf("expected rehydrate json note, got: %s", output)
	}
	if !strings.Contains(output, "echo \"Need auth review\" | yanzi capture") {
		t.Fatalf("expected stdin capture example, got: %s", output)
	}
	exportSection := output[strings.Index(output, "export args:"):]
	if strings.Contains(exportSection, "--fields a,b") || strings.Contains(exportSection, "--order <field>") || strings.Contains(exportSection, "--limit <n>") {
		t.Fatalf("did not expect removed export flags in export help text, got: %s", exportSection)
	}
	if !strings.Contains(exportSection, "--format claude-context Generates CLAUDE_CONTEXT.md in project root.\n                        Required.") {
		t.Fatalf("expected required export format help text, got: %s", output)
	}
}

func TestRunUnknownCommandReturnsExitCodeOne(t *testing.T) {
	output := captureStderr(t, func() {
		if got := run([]string{"doesnotexist"}); got != 1 {
			t.Fatalf("expected exit code 1, got %d", got)
		}
	})
	if !strings.Contains(output, "unknown command: doesnotexist") {
		t.Fatalf("expected unknown command error, got: %s", output)
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = old
	}()

	fn()
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}
