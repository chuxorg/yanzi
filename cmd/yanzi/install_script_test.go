package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallScriptSucceedsWithLocalReleaseAsset(t *testing.T) {
	home := t.TempDir()
	installDir := filepath.Join(home, "bin")
	releaseJSON := createLocalReleaseFixture(t, home)

	output := runInstallScript(t, filepath.Join("..", "..", "install.sh"), map[string]string{
		"HOME":                       home,
		"PATH":                       os.Getenv("PATH"),
		"SHELL":                      "/bin/sh",
		"YANZI_INSTALL_DIR":          installDir,
		"YANZI_INSTALL_RELEASES_API": "file://" + releaseJSON,
		"YANZI_INSTALL_OS":           "Linux",
		"YANZI_INSTALL_ARCH":         "amd64",
	})

	if !strings.Contains(output, "yanzi v9.9.9") {
		t.Fatalf("expected installed version output, got %q", output)
	}
	if !strings.Contains(output, "installed") {
		t.Fatalf("expected installer status output, got %q", output)
	}
	if _, err := os.Stat(filepath.Join(installDir, "yanzi")); err != nil {
		t.Fatalf("expected installed binary: %v", err)
	}
}

func TestScriptsInstallWrapperSucceeds(t *testing.T) {
	home := t.TempDir()
	installDir := filepath.Join(home, "bin")
	releaseJSON := createLocalReleaseFixture(t, home)

	output := runInstallScript(t, filepath.Join("..", "..", "scripts", "install.sh"), map[string]string{
		"HOME":                       home,
		"PATH":                       os.Getenv("PATH"),
		"SHELL":                      "/bin/sh",
		"YANZI_INSTALL_DIR":          installDir,
		"YANZI_INSTALL_RELEASES_API": "file://" + releaseJSON,
		"YANZI_INSTALL_OS":           "Linux",
		"YANZI_INSTALL_ARCH":         "amd64",
	})

	if !strings.Contains(output, "yanzi v9.9.9") {
		t.Fatalf("expected wrapper install version output, got %q", output)
	}
	if !strings.Contains(output, "installed") {
		t.Fatalf("expected wrapper install success, got %q", output)
	}
}

func TestInstallScriptFailsWhenReleaseMetadataIsEmpty(t *testing.T) {
	home := t.TempDir()
	releaseJSON := filepath.Join(home, "release.json")
	if err := os.WriteFile(releaseJSON, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write release json: %v", err)
	}

	output, err := runInstallScriptExpectError(filepath.Join("..", "..", "install.sh"), map[string]string{
		"HOME":                       home,
		"PATH":                       os.Getenv("PATH"),
		"YANZI_INSTALL_DIR":          filepath.Join(home, "bin"),
		"YANZI_INSTALL_RELEASES_API": "file://" + releaseJSON,
		"YANZI_INSTALL_OS":           "Linux",
		"YANZI_INSTALL_ARCH":         "amd64",
	})
	if err == nil {
		t.Fatal("expected install failure")
	}
	if !strings.Contains(output, "release metadata did not include any downloadable assets") {
		t.Fatalf("unexpected error output: %q", output)
	}
}

func TestInstallScriptFailsWithoutCurl(t *testing.T) {
	home := t.TempDir()
	pathDir := filepath.Join(home, "pathbin")
	if err := os.MkdirAll(pathDir, 0o755); err != nil {
		t.Fatalf("mkdir pathbin: %v", err)
	}

	for _, name := range []string{"uname", "mktemp", "chmod", "grep", "awk", "sed", "head", "mv", "mkdir", "rm", "tar"} {
		target, err := exec.LookPath(name)
		if err != nil {
			t.Fatalf("lookpath %s: %v", name, err)
		}
		if err := os.Symlink(target, filepath.Join(pathDir, name)); err != nil {
			t.Fatalf("symlink %s: %v", name, err)
		}
	}

	output, err := runInstallScriptExpectError(filepath.Join("..", "..", "install.sh"), map[string]string{
		"HOME":              home,
		"PATH":              pathDir,
		"YANZI_INSTALL_DIR": filepath.Join(home, "bin"),
	})
	if err == nil {
		t.Fatal("expected missing curl failure")
	}
	if !strings.Contains(output, "required tool 'curl' is not available") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestInstallScriptFailsOnUnsupportedArchitecture(t *testing.T) {
	home := t.TempDir()
	output, err := runInstallScriptExpectError(filepath.Join("..", "..", "install.sh"), map[string]string{
		"HOME":               home,
		"PATH":               os.Getenv("PATH"),
		"YANZI_INSTALL_DIR":  filepath.Join(home, "bin"),
		"YANZI_INSTALL_OS":   "Linux",
		"YANZI_INSTALL_ARCH": "mips64",
	})
	if err == nil {
		t.Fatal("expected unsupported architecture failure")
	}
	if !strings.Contains(output, "unsupported architecture") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func createLocalReleaseFixture(t *testing.T, root string) string {
	t.Helper()

	assetDir := filepath.Join(root, "assets")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatalf("mkdir asset dir: %v", err)
	}
	binaryPath := filepath.Join(assetDir, "yanzi-linux-amd64")
	fakeBinary := "#!/usr/bin/env sh\nif [ \"$1\" = \"--version\" ]; then\n  echo \"yanzi v9.9.9\"\n  exit 0\nfi\necho \"yanzi help\"\n"
	if err := os.WriteFile(binaryPath, []byte(fakeBinary), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	releaseJSON := filepath.Join(root, "release.json")
	payload := "{\n  \"tag_name\": \"v9.9.9\",\n  \"browser_download_url\": \"file://" + filepath.ToSlash(binaryPath) + "\"\n}\n"
	if runtime.GOOS == "windows" {
		payload = strings.ReplaceAll(payload, "\\", "/")
	}
	if err := os.WriteFile(releaseJSON, []byte(payload), 0o644); err != nil {
		t.Fatalf("write release json: %v", err)
	}
	return releaseJSON
}

func runInstallScript(t *testing.T, script string, env map[string]string) string {
	t.Helper()

	output, err := runInstallScriptExpectError(script, env)
	if err != nil {
		t.Fatalf("install script failed: %v\noutput=%s", err, output)
	}
	return output
}

func runInstallScriptExpectError(script string, env map[string]string) (string, error) {
	cmd := exec.Command("/bin/sh", script)
	cmd.Env = flattenEnv(env)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func flattenEnv(overrides map[string]string) []string {
	envMap := map[string]string{}
	for _, entry := range os.Environ() {
		key, value, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		envMap[key] = value
	}
	for key, value := range overrides {
		envMap[key] = value
	}
	env := make([]string, 0, len(envMap))
	for key, value := range envMap {
		env = append(env, key+"="+value)
	}
	return env
}
