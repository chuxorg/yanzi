package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var mdLinkRE = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

type result struct {
	file   string
	issue  string
	detail string
}

func main() {
	root := flag.String("root", ".", "repository root")
	checkExternal := flag.Bool("check-external", false, "attempt HTTP checks for external URLs")
	timeout := flag.Duration("timeout", 5*time.Second, "external URL timeout")
	flag.Parse()

	files, err := findMarkdownFiles(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: find markdown files: %v\n", err)
		os.Exit(1)
	}

	failures := make([]result, 0)
	warnings := make([]result, 0)

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, result{file: path, issue: "read_error", detail: err.Error()})
			continue
		}
		links := mdLinkRE.FindAllStringSubmatch(string(data), -1)
		for _, match := range links {
			raw := sanitizeLinkTarget(match[1])
			if raw == "" || strings.HasPrefix(raw, "#") {
				continue
			}
			if strings.HasPrefix(raw, "mailto:") {
				if _, err := url.Parse(raw); err != nil {
					failures = append(failures, result{file: path, issue: "malformed_url", detail: raw})
				}
				continue
			}
			if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
				u, err := url.ParseRequestURI(raw)
				if err != nil || u.Scheme == "" || u.Host == "" {
					failures = append(failures, result{file: path, issue: "malformed_url", detail: raw})
					continue
				}
				if *checkExternal {
					if err := checkURL(raw, *timeout); err != nil {
						warnings = append(warnings, result{file: path, issue: "external_unreachable", detail: raw + " :: " + err.Error()})
					}
				}
				continue
			}

			target := strings.Split(raw, "#")[0]
			if target == "" {
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), target))
			if _, err := os.Stat(resolved); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					failures = append(failures, result{file: path, issue: "broken_internal_link", detail: raw})
					continue
				}
				failures = append(failures, result{file: path, issue: "stat_error", detail: raw + " :: " + err.Error()})
			}
		}
	}

	for _, warn := range warnings {
		fmt.Printf("WARN: %s: %s (%s)\n", warn.file, warn.issue, warn.detail)
	}
	for _, fail := range failures {
		fmt.Printf("FAIL: %s: %s (%s)\n", fail.file, fail.issue, fail.detail)
	}

	if len(failures) > 0 {
		fmt.Printf("FAIL: docs validation found %d failure(s), %d warning(s)\n", len(failures), len(warnings))
		os.Exit(1)
	}
	fmt.Printf("PASS: docs validation checked %d markdown files with %d warning(s)\n", len(files), len(warnings))
}

func sanitizeLinkTarget(target string) string {
	target = strings.TrimSpace(target)
	target = strings.TrimPrefix(target, "<")
	target = strings.TrimSuffix(target, ">")
	return strings.TrimSpace(target)
}

func findMarkdownFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func checkURL(raw string, timeout time.Duration) error {
	client := &httpClient{timeout: timeout}
	status, err := client.head(raw)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("status %d", status)
	}
	return nil
}
