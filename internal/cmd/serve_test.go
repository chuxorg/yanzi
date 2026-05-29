package cmd

import (
	"flag"
	"testing"
)

func TestServeDefaultHost(t *testing.T) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	host := fs.String("host", "127.0.0.1", "Host address to bind the HTTP server")
	if err := fs.Parse(nil); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if *host != "127.0.0.1" {
		t.Errorf("default host: got %q, want %q", *host, "127.0.0.1")
	}
}
