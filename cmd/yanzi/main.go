package main

import (
	"fmt"
	"os"

	"github.com/chuxorg/yanzi/internal/cmd"
	"github.com/chuxorg/yanzi/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi/internal/library"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	initialized, err := yanzilibrary.Initialize()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if initialized {
		fmt.Println("Yanzi initialized at ~/.yanzi")
	}

	if os.Args[1] == "--version" {
		if err := printVersion(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if isHelpArg(os.Args[1]) {
		usage()
		return
	}

	err = nil
	switch os.Args[1] {
	case "capture":
		err = cmd.RunCapture(os.Args[2:])
	case "verify":
		err = cmd.RunVerify(os.Args[2:])
	case "chain":
		err = cmd.RunChain(os.Args[2:])
	case "list":
		err = cmd.RunList(os.Args[2:])
	case "show":
		err = cmd.RunShow(os.Args[2:])
	case "mode":
		err = cmd.RunMode(os.Args[2:])
	case "project":
		err = cmd.RunProject(os.Args[2:])
	case "checkpoint":
		err = cmd.RunCheckpoint(os.Args[2:])
	case "rehydrate":
		err = cmd.RunRehydrate(os.Args[2:])
	case "export":
		err = cmd.RunExport(os.Args[2:], version)
	case "version":
		if err := printVersion(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  yanzi <command> [args]

commands:
  capture  Create a new intent record via the library API.
  verify   Verify an intent by id.
  chain    Print an intent chain by id.
  list     List intent records.
  show     Show intent details by id.
  mode     Show or set runtime mode (local | http).
  project  Manage project context.
  checkpoint  Manage checkpoints.
  rehydrate  Rehydrate active project context.
  export  Export active project history.
  version  Print the CLI version.

capture args:
  --author <name>         Required author name.
  --prompt <text>         Prompt text (exclusive with --prompt-file).
  --prompt-file <path>    Prompt file path (exclusive with --prompt).
  --response <text>       Response text (exclusive with --response-file).
  --response-file <path>  Response file path (exclusive with --response).
  --title <title>         Optional title.
  --source <source>       Optional source type (default "cli").
  --prev-hash <hash>      Optional previous hash.
  --meta key=value        Optional metadata (repeatable).

verify args:
  <intent-id>             Intent id to verify.

chain args:
  <intent-id>             Intent id to chain.

list args:
  --author <name>         Optional author filter.
  --source <source>       Optional source filter.
  --meta k=v              Optional meta filter (repeatable; exact match; AND).
  --limit <n>             Max records to return (default 20).

show args:
  <intent-id>             Intent id to show.

mode args:
  (no args)              Show current mode.
  local                  Set mode to local.
  http                   Set mode to http.

project args:
  create <name>         Create a new project.
  use <name>            Set the active project.
  current               Show the active project.
  list                  List projects.

checkpoint args:
  create --summary "..." Create a checkpoint for the active project.
  list                   List checkpoints for the active project.

rehydrate args:
  (no args)             Rehydrate the active project context.

export args:
  --format markdown     Export active project history to ./YANZI_LOG.md.
  --format json         Generates YANZI_LOG.json in project root.
  --format html         Generates YANZI_LOG.html in project root.

notes:
  mode set to http does not start libraryd.

examples:
  yanzi capture --author "Ada" --prompt-file prompt.txt --response-file response.txt --meta lang=go
  yanzi capture --author "Ada" --prompt "Hello" --response "World"
  yanzi verify 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi chain 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi list --limit 10
  yanzi show 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi mode
  yanzi mode local
  yanzi mode http
  yanzi project create "alpha"
  yanzi project use "alpha"
  yanzi project current
  yanzi project list
  yanzi checkpoint create --summary "Weekly snapshot"
  yanzi checkpoint list
  yanzi rehydrate
  yanzi export --format markdown
  yanzi export --format json
  yanzi export --format html
  yanzi version`)
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "?"
}

func printVersion() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Printf("yanzi %s\n", version)
	fmt.Printf("mode: %s\n", formatMode(cfg))
	return nil
}

func formatMode(cfg config.Config) string {
	switch cfg.Mode {
	case config.ModeHTTP:
		return fmt.Sprintf("http (%s)", cfg.BaseURL)
	default:
		return "local"
	}
}
