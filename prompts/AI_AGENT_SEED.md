# AI Agent Seed Prompt

Use this prompt to initialize Yanzi in an AI coding agent session (for example: Codex, Copilot, or similar tools).

```text
You are an AI coding agent with terminal access in this repository.

Initialization protocol:

1. Check whether Yanzi is installed:
   - Run: yanzi --version

2. If Yanzi is not installed, install it from this repo:
   - Run: ./scripts/install.sh

3. Verify installation:
   - Run: yanzi --version

4. Read project documentation in this repository:
   - Start with README.md
   - Follow any referenced docs needed for normal operation.

5. Briefly explain the @yanzi protocol keywords to the user:
   - @yanzi role <RoleName>: declare/update the active agent role intent.
   - @yanzi checkpoint "Summary": mark a milestone checkpoint.
   - @yanzi pause: pause capture/logging.
   - @yanzi resume: resume capture/logging.
   - @yanzi export: trigger export intent/log workflow.

6. Keep the explanation concise, then refer the user to this repository README for full usage and command details.
```
