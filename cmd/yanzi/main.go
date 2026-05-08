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
	case "delete":
		err = cmd.RunDelete(os.Args[2:])
	case "restore":
		err = cmd.RunRestore(os.Args[2:])
	case "mode":
		err = cmd.RunMode(os.Args[2:])
	case "project":
		err = cmd.RunProject(os.Args[2:])
	case "init":
		err = cmd.RunInit(os.Args[2:])
	case "intent":
		err = cmd.RunIntent(os.Args[2:])
	case "context":
		err = cmd.RunContext(os.Args[2:])
	case "pack":
		err = cmd.RunPack(os.Args[2:])
	case "bootstrap":
		err = cmd.RunBootstrap(os.Args[2:])
	case "rules":
		err = cmd.RunRules(os.Args[2:], version)
	case "types":
		err = cmd.RunTypes(os.Args[2:])
	case "message":
		err = cmd.RunMessage(os.Args[2:])
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
  capture  Create a new intent record.
  verify   Verify an intent by id.
  chain    Print an intent chain by id.
  list     List intent records.
  show     Show intent details by id.
  delete   Tombstone an intent or artifact by id.
  restore  Remove tombstone metadata by id.
  mode     Show or set runtime mode (local | http).
  project  Manage project context.
  init  Create or bind a project to the current directory.
  intent  Manage intent artifacts.
  context  Manage context artifacts.
  pack  Apply or export portable context packs.
  bootstrap  Load ordered context documents from .yanzi/bootstrap.yaml.
  rules  Manage rule metadata wrappers.
  types  List canonical artifact types and aliases.
  message  Manage thin message wrappers.
  checkpoint  Manage checkpoints.
  rehydrate  Rehydrate active project context.
  export  Export active project history.
  version  Print the CLI version.

global args:
  -h, --help, ?          Show help.
  --version              Print the CLI version.

capture args:
  --author <name>         Required author name.
  --prompt <text>         Prompt text (exclusive with --prompt-file).
  --prompt-file <path>    Prompt file path (exclusive with --prompt).
  --response <text>       Response text (exclusive with --response-file).
  --response-file <path>  Response file path (exclusive with --response).
  --title <title>         Optional title.
  --source <source>       Optional source type (default "cli").
  --profile <name>        Optional profile label.
  --prev-hash <hash>      Optional previous hash.
  --meta key=value        Optional metadata (repeatable).

verify args:
  <intent-id>             Intent id to verify.

chain args:
  <intent-id>             Intent id to chain.

list args:
  --author <name>         Optional author filter.
  --source <source>       Optional source filter.
  --profile <name>        Optional profile filter.
  --meta k=v              Optional meta filter (repeatable; exact match; AND).
  --include-deleted       Include tombstoned records.
  --limit <n>             Max records to return (default 20).

show args:
  <intent-id>             Intent id to show.

delete args:
  <intent-id>             Intent or artifact id to tombstone.
  --cascade               Also tombstone dependent chain records.
  --force                 Allow tombstoning checkpoint-referenced artifacts.

restore args:
  <intent-id>             Intent or artifact id to restore.

mode args:
  (no args)              Show current mode.
  local                  Set mode to local.
  http                   Set mode to http.

project args:
  create <name>         Create a new project.
  use <name>            Set the active project.
  current               Show the active project.
  list                  List projects.

init args:
  [name]                Create or reuse a project and bind ./.yanzi/project.

intent args:
  add --title "..."     Add an intent artifact.
  list                  List intent artifacts.

context args:
  add --type "..." --title "..." [--scope global|project]
                        Add a context artifact.
  list                  List visible context artifacts.
  show <id>             Show a context artifact by id.

pack args:
  apply <file>          Apply a YAML context pack.
  export --output <file>
                        Export visible context into a pack YAML and sidecar files.

bootstrap args:
  --dry-run             Validate .yanzi/bootstrap.yaml without loading documents.

rules args:
  add <file>            Capture a rules file with context metadata.
  list                  List rule captures only.
  export                Export rule captures only; supports --compose for markdown and html.

types args:
  list --json           Show canonical types and alias mappings.

message args:
  send                  Capture a message note with routing metadata.
  list                  List stored message notes.
  pull                  Pull stored message notes as markdown.



checkpoint args:
  create --summary "..." Create a checkpoint for the active project.
  list                   List checkpoints for the active project.

rehydrate args:
  (no args)             Rehydrate the active project context.
  --dry-run             Preview the checkpoint/context load without rehydrating.

export args:
  --format markdown     Export active project history to ./YANZI_LOG.md.
  --format json         Generates YANZI_LOG.json in project root.
  --format html         Generates YANZI_LOG.html in project root.
  --format claude-context Generates CLAUDE_CONTEXT.md in project root.
                        Default format when omitted.
  --type a,b            Optional context type filter list.
  --profile <name>      Optional profile filter.
  --meta key=value      Optional metadata filter (repeatable; exact match; AND).
  --fields a,b          Optional context field list.
  --order <field>       Deterministic order field: created_at|updated_at.
  --limit <n>           Optional result limit after filtering.
  --include-deleted     Include tombstoned records.

notes:
  mode set to http does not start libraryd.

examples:
  yanzi --help
  yanzi --version
  yanzi capture --author "Ada" --prompt-file prompt.txt --response-file response.txt --meta lang=go
  yanzi capture --author "Ada" --prompt "Hello" --response "World" --profile engineer
  yanzi capture --author "Ada" --prompt "Hello" --response "World"
  yanzi verify 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi chain 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi list --limit 10
  yanzi show 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi delete 01HZX9Q4X8N9JZ1K2G9N8M4V3P --cascade
  yanzi restore 01HZX9Q4X8N9JZ1K2G9N8M4V3P
  yanzi mode
  yanzi mode local
  yanzi mode http
  yanzi project create "alpha"
  yanzi project use "alpha"
  yanzi project current
  yanzi project list
  yanzi intent add --title "Clarify export scope" --content "Export only deterministic artifacts."
  yanzi intent list --type decision
  yanzi context add --type process_rule --title "Release rule" --file ./policy.md
  yanzi context add --type governance --title "Release rule" --file ./policy.md
  yanzi context list --scope project
  yanzi context show abc123def456
  yanzi bootstrap --dry-run
  yanzi rules add ./SYSTEM_RULES.md --scope global --priority critical
  yanzi rules add ./SYSTEM_RULES.md --profile engineer
  yanzi rules list --scope global
  yanzi rules list --profile engineer
  yanzi types list --json
  yanzi message send --to codex --from operator --channel execution --file ./ready.md
  yanzi message pull --to codex --channel execution
  yanzi rules export --format markdown --profile default
  yanzi rules export --format markdown --compose --profile engineer
  yanzi rules export --format html --compose --profile engineer
  yanzi checkpoint create --summary "Weekly snapshot"
  yanzi checkpoint list
  yanzi rehydrate --dry-run
  yanzi export --format markdown
  yanzi export --meta type=context --meta subtype=rules --format markdown
  yanzi export --format json
  yanzi export --format html
  yanzi export --format claude-context
  yanzi export --meta pack=vibe-coder --limit 3
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
