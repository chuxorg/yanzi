# Yanzi Tutorial

This tutorial walks through the normal Yanzi workflow from project setup to capture, export, and reviewing the HTML UI.

## Start With the AI Agent Seed Prompt

If you are using Yanzi with an AI coding agent, first give the agent the seed prompt from:

- [AI_AGENT_SEED.md](/Users/developer/projects/chuxorg/chux-yanzi-cli/prompts/AI_AGENT_SEED.md)

That file tells the agent how to initialize Yanzi, verify installation, and understand the `@yanzi` protocol lines.

Recommended workflow:

1. Open the seed prompt file.
2. Copy its contents.
3. Paste it into your chosen AI agent at the start of the session.
4. Then continue with the repository README and this tutorial.

## What Yanzi Is

Yanzi is a deterministic logging CLI for AI-assisted development.

It helps you preserve the decision trail behind your work, not just the final code diff. Instead of losing prompt/response context across chats and sessions, Yanzi records:

- captures
- checkpoints
- role/meta events
- optional metadata tied to a decision

The result is a project history you can review, export, rehydrate, and share.

## Yanzi Projects and Intent

A Yanzi **project** is the top-level container for a stream of related development activity.

You usually create one project per repository or major workstream. Once a project is active, captures and checkpoints are associated with that project and can be exported together.

A Yanzi **intent** is the recorded unit of development reasoning. In practice, the most common intent is a capture:

- prompt: what was asked
- response: what the AI or operator produced
- author/role/source: who or what created it
- metadata: extra context such as `area=auth` or `decision_type=refactor`
- hash: the immutable content fingerprint

Checkpoints are milestone entries that mark a meaningful boundary in the project timeline.

## Install and Verify

If Yanzi is not already installed:

```sh
./scripts/install.sh
```

Verify the CLI:

```sh
yanzi version
```

If you want to use the repo-local binary instead:

```sh
go build -o ./yanzi ./cmd/yanzi
./yanzi version
```

## Create and Select a Project

Create a project:

```sh
yanzi project create "alpha"
```

Then make it active:

```sh
yanzi project use "alpha"
```

Confirm the active project:

```sh
yanzi project current
```

This step matters because export and checkpoint commands operate on the active project.

## Capture Development Work

Capture a prompt/response pair inline:

```sh
yanzi capture \
  --author "Ada" \
  --prompt "Summarize the auth refactor plan" \
  --response "Split login, session, and token validation into separate handlers."
```

Capture with metadata:

```sh
yanzi capture \
  --author "Ada" \
  --prompt "What changed?" \
  --response "Moved token parsing into middleware." \
  --meta area=auth \
  --meta decision_type=refactor \
  --meta tags=middleware,security
```

Capture from files:

```sh
yanzi capture \
  --author "Ada" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta area=docs
```

## Use Checkpoints

Add a milestone checkpoint when you reach a stable boundary:

```sh
yanzi checkpoint create --summary "Auth middleware refactor complete"
```

List checkpoints:

```sh
yanzi checkpoint list
```

Checkpoints become strong boundary markers in the HTML export and help with later review and rehydration.

## Export Your Project History

Yanzi supports three export formats:

```sh
yanzi export --format markdown
yanzi export --format json
yanzi export --format html
```

Outputs:

- `YANZI_LOG.md`
- `YANZI_LOG.json`
- `YANZI_LOG.html`

If you want to open the HTML export immediately:

```sh
yanzi export --format html --open
```

## View the HTML UI

Open the exported file directly in a browser:

```sh
open YANZI_LOG.html
```

Or use the built-in flag:

```sh
yanzi export --format html --open
```

The HTML export is:

- single-file
- fully static
- usable from `file://`
- dependency-free

## How to Read the HTML Export

The HTML export is designed for scanning project history quickly while preserving the underlying data.

### Header

The sticky header shows:

- project name
- export timestamp
- CLI version
- total events
- total captures
- total checkpoints

This stays visible while scrolling so you keep context during longer reviews.

### Search and Filter

Near the top of the page there is a search input.

It filters across full event content, including:

- prompt text
- response text
- role
- source
- hash
- checkpoint summary
- timestamp
- badge text

The visible match counter updates as you type.

### Timeline

Events are rendered in chronological order on a vertical timeline.

- captures appear as regular timeline entries
- checkpoints appear as stronger boundary entries
- meta events appear as lighter timeline events

This lets you scan the sequence of work without changing the underlying export data.

### Capture Cards

Capture cards usually show:

- capture ID
- role
- human-readable timestamp
- hash
- optional metadata table
- prompt section
- response section

Collapsed prompt/response sections show preview snippets. Expanding them reveals the full content.

### Checkpoint Cards

Checkpoint cards show:

- checkpoint ID
- summary
- timestamp
- checkpoint-specific visual treatment

These act as project-history boundary markers.

### Semantic Badges

Badges help explain what kind of event you are looking at.

Common examples:

- `Capture`
- `Checkpoint`
- `Prompt`
- `Response`
- `Hash`
- `Metadata`
- `Role: ...`
- `Source: ...`
- `Boundary`
- `Rehydration Anchor`

Badges clarify existing data only; they do not invent new metadata.

### Timestamps

Visible timestamps are formatted for readability in the UI.

The original raw timestamp is still preserved in the tooltip via the `title` attribute, so you can inspect the canonical exported value on hover.

### Copy Buttons

The UI includes copy buttons for key fields such as:

- prompt
- response
- capture ID
- checkpoint ID
- hash

This is useful when you want to quote or inspect an entry without manually selecting large blocks.

## Typical End-to-End Flow

```sh
yanzi project create "alpha"
yanzi project use "alpha"

yanzi capture \
  --author "Ada" \
  --prompt "Plan the docs rework" \
  --response "Add a tutorial, update README links, and document the HTML export UI."

yanzi checkpoint create --summary "Tutorial docs drafted"

yanzi export --format html --open
```

## Useful Supporting Commands

Current project:

```sh
yanzi project current
```

List projects:

```sh
yanzi project list
```

Rehydrate the active project:

```sh
yanzi rehydrate
```

## Practical Tips

- Always run `yanzi project use "<name>"` after creating a new project.
- Use metadata consistently if you want more useful exports later.
- Add checkpoints at meaningful review boundaries, not on every small change.
- Use HTML export for human review, JSON export for integration, and Markdown for quick plain-text logs.

## See Also

- [Agent Bootstrap](./agent-bootstrap.md)
- [Home](./index.md)
- [README.md](../README.md)
