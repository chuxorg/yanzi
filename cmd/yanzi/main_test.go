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
	if !strings.Contains(output, "--meta key=value") {
		t.Fatalf("expected capture metadata help text, got: %s", output)
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
