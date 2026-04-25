# Rehydrate

Rehydration reconstructs the current state of a project.

Steps:

1. Load active project
2. Locate latest checkpoint
3. Retrieve artifacts created after checkpoint
4. Return ordered sequence

Rehydrate does not summarize or reinterpret artifacts.

It simply reconstructs the artifact chain.

This guarantees deterministic behavior.