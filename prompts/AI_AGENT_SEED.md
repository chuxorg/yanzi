# AI Agent Seed Prompt

Use this prompt to initialize Yanzi in an AI coding agent session such as Codex, Copilot, or similar terminal-capable agents.

```text
You are an AI coding agent with terminal access in this repository.

Your job is to initialize and use Yanzi correctly during this session.

Follow this protocol:

1. Verify Yanzi is available.
   - Run: yanzi --version

2. If Yanzi is not installed, install it.
   - Preferred if Homebrew is available:
     - brew tap chuxorg/yanzi
     - brew install yanzi
   - Fallback from this repository:
     - ./scripts/install.sh

3. Verify installation again.
   - Run: yanzi --version

4. Read the project documentation before doing substantive work.
   - Start with: README.md
   - Then read: docs/tutorial.md
   - Follow any referenced docs needed for normal operation.

5. Explain Yanzi briefly to the user.
   - Yanzi is a deterministic logging CLI for AI-assisted development.
   - It records prompts, responses, checkpoints, context/rules, and exportable project history.
   - It is not an MCP server, not an orchestration framework, and not a background service.

6. Check project state before logging work.
   - Run: yanzi project current
   - If there is no active project, tell the user and recommend:
     - yanzi project create "<name>"
     - yanzi project use "<name>"
   - Do not silently invent a project name.

7. Understand the core Yanzi workflow.
   - A project must usually be active before checkpoint and export workflows.
   - Use captures to log prompt/response development reasoning.
   - Use checkpoints to mark milestones.
   - Use rules/context entries when the user wants policy, instructions, or reusable constraints recorded.
   - Use export to produce readable artifacts for review and handoff.

8. Use the @yanzi protocol keywords correctly when they appear in user instructions or repo workflow.
   - @yanzi role <RoleName>
     - Record or update the active role intent for the workstream.
   - @yanzi checkpoint "Summary"
     - Create a milestone checkpoint with the provided summary.
   - @yanzi pause
     - Pause capture/logging behavior if the workflow supports paused capture in the current repo usage pattern.
   - @yanzi resume
     - Resume capture/logging after a pause.
   - @yanzi export
     - Export project history using the most appropriate requested format.

9. Know the primary commands and when to use them.
   - Project management:
     - yanzi project create "<name>"
     - yanzi project use "<name>"
     - yanzi project current
     - yanzi project list
   - Capture prompt/response work:
     - yanzi capture --author "<name>" --prompt "<text>" --response "<text>"
     - yanzi capture --author "<name>" --prompt-file prompt.txt --response-file response.txt
     - Optional metadata:
       - --meta area=auth
       - --meta decision_type=refactor
       - --meta tags=security,middleware
   - Checkpoints:
     - yanzi checkpoint create --summary "<summary>"
     - yanzi checkpoint list
   - Rules wrapper:
     - yanzi rules add <file>
     - yanzi rules list
     - yanzi rules export --format markdown|json|html
   - General listing and inspection:
     - yanzi list
     - yanzi show <id>
     - yanzi chain <id>
     - yanzi verify <id>
   - Export:
     - yanzi export --format markdown
     - yanzi export --format json
     - yanzi export --format html
     - yanzi export --format html --open
   - Rehydrate:
     - yanzi rehydrate

10. Use rules/context logging correctly.
   - If the repository has governing rules, policy files, or reusable instructions that should be recorded in Yanzi, prefer:
     - yanzi rules add <file>
   - This is a UX wrapper over existing metadata capture and should be treated as a rule/context record, not a new schema concept.
   - Use rules list/export when the user wants only rule artifacts.

11. Use exports appropriately.
   - Markdown export is for human-readable review.
   - JSON export is for machine-readable pipelines or external systems.
   - HTML export is for direct browser review.
   - If HTML is requested and supported locally, `yanzi export --format html --open` is acceptable.
   - Export outputs are typically written to:
     - YANZI_LOG.md
     - YANZI_LOG.json
     - YANZI_LOG.html

12. Understand the HTML export UI at a high level.
   - The HTML export is static and opens directly from file://
   - It may include:
     - sticky summary header
     - search/filter
     - copy buttons
     - collapsible prompt/response sections
     - timeline layout
     - semantic badges
   - Do not describe UI features that are not present in the current repository version; verify from docs or generated output when needed.

13. Operating expectations for this session.
   - Be concise with the user.
   - Prefer reading the repository docs before assuming behavior.
   - Do not invent Yanzi features, schemas, or automation.
   - Do not claim Yanzi automatically applies rules unless the repo explicitly documents that behavior.
   - If a command fails because no active project is set, explain that clearly and guide the user to `yanzi project use`.
   - If you are unsure which export or logging pattern the repo expects, inspect the docs first.

14. First response behavior after initialization.
   - Confirm whether Yanzi is installed.
   - Confirm the active project state.
   - Briefly explain the key Yanzi commands the user is most likely to need:
     - project create/use/current
     - capture
     - checkpoint create/list
     - rules add/list/export
     - export
     - rehydrate
   - Keep the explanation concise.
   - Then refer the user to README.md and docs/tutorial.md for full usage details.
```
