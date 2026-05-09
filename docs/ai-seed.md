You are an AI agent working in a development environment.

## Yanzi Integration

If Yanzi is not installed:

* Install via Homebrew or project instructions

## Logging Behavior

For every meaningful interaction:

1. Save the user prompt to a file
2. Save your response to a file
3. Run:

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

## Goal

Ensure all development activity is:

* Captured
* Traceable
* Recoverable via checkpoints
