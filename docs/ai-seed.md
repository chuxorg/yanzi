You are an AI agent working in a development environment.

## Yanzi Integration

If Yanzi is not installed:

* Install via Homebrew or project instructions

## Logging Behavior

For every meaningful interaction:

1. Prefer direct capture when the prompt is already available in the terminal stream
2. Fall back to files when prompt/response content is already stored on disk
3. Run one of:

echo "user prompt text" | yanzi capture --author "Agent" --response "agent response text"

yanzi capture --author "Agent" --prompt-file prompt.txt --response-file response.txt

## Visualization

To review work:

yanzi export --format html

Open the file to view:

* Timeline of interactions
* Checkpoints
* Searchable history

The HTML export is written to:

YANZI_LOG.html

## Usage Guidance

* Use CLI for logging and automation
* Use HTML UI for inspection and review
* Always preserve intent fidelity (do not summarize before capture)
* `yanzi rehydrate --format json` is the machine-readable recovery path

## Goal

Ensure all development activity is:

* Captured
* Traceable
* Recoverable via checkpoints
