# Seeding Context

## Problem

Projects require consistent rules and context before work begins.

## Approach

Capture shared rules explicitly:

```bash
yanzi rules add ./SYSTEM_RULES.md --scope global --priority critical
yanzi rules export --format markdown
```

## Result

All agents start with shared rules and context.

## Note

Yanzi does not enforce structure. Teams choose which rules to capture and export.
