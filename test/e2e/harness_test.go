package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type runResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type cliHarness struct {
	t       *testing.T
	binPath string
	homeDir string
	workDir string
}

var (
	buildOnce sync.Once
	buildPath string
	buildErr  error
)

func newCLIHarness(t *testing.T) *cliHarness {
	t.Helper()
	bin, err := buildYanziBinary(t)
	if err != nil {
		t.Fatalf("build yanzi binary: %v", err)
	}

	homeDir := t.TempDir()
	workDir := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	return &cliHarness{t: t, binPath: bin, homeDir: homeDir, workDir: workDir}
}

func buildYanziBinary(t *testing.T) (string, error) {
	t.Helper()
	buildOnce.Do(func() {
		tmp := t.TempDir()
		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
		buildPath = filepath.Join(tmp, "yanzi"+ext)
		cmd := exec.Command("go", "build", "-o", buildPath, "./cmd/yanzi")
		cmd.Dir = repoRoot(t)
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = fmt.Errorf("go build failed: %w\n%s", err, strings.TrimSpace(string(out)))
			return
		}
	})
	return buildPath, buildErr
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func (h *cliHarness) run(args ...string) runResult {
	h.t.Helper()
	cmd := exec.Command(h.binPath, args...)
	cmd.Dir = h.workDir
	cmd.Env = append(os.Environ(),
		"HOME="+h.homeDir,
		"USERPROFILE="+h.homeDir,
		"XDG_CONFIG_HOME="+filepath.Join(h.homeDir, ".config"),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			h.t.Fatalf("execute command %v: %v", args, err)
		}
	}
	return runResult{ExitCode: code, Stdout: stdout.String(), Stderr: stderr.String()}
}

func (h *cliHarness) requireExitCode(got runResult, code int) {
	h.t.Helper()
	if got.ExitCode != code {
		h.t.Fatalf("unexpected exit code %d want %d\nstdout:\n%s\nstderr:\n%s", got.ExitCode, code, got.Stdout, got.Stderr)
	}
}

func (h *cliHarness) requireContains(text, needle string) {
	h.t.Helper()
	if !strings.Contains(text, needle) {
		h.t.Fatalf("missing substring %q in:\n%s", needle, text)
	}
}

func (h *cliHarness) requireFile(path string) string {
	h.t.Helper()
	b, err := os.ReadFile(filepath.Join(h.workDir, path))
	if err != nil {
		h.t.Fatalf("read file %s: %v", path, err)
	}
	return string(b)
}

func normalizeExportSnapshot(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`Generated: .*`),
		regexp.MustCompile(`Exported: .*`),
		regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9:\.-]+Z`),
		regexp.MustCompile(`[a-f0-9]{64}`),
		regexp.MustCompile(`[a-f0-9]{32}`),
		regexp.MustCompile(`id: [a-zA-Z0-9_-]{20,}`),
	}
	for _, re := range patterns {
		s = re.ReplaceAllString(s, "<normalized>")
	}
	return strings.TrimSpace(s) + "\n"
}

func compareOrUpdateSnapshot(t *testing.T, relPath, got string) {
	t.Helper()
	abs := filepath.Join(repoRoot(t), relPath)
	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		if err := os.WriteFile(abs, []byte(got), 0o644); err != nil {
			t.Fatalf("update snapshot %s: %v", relPath, err)
		}
		return
	}
	want, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("read snapshot %s: %v", relPath, err)
	}
	if string(want) != got {
		t.Fatalf("snapshot mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", relPath, string(want), got)
	}
}
