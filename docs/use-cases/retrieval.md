# Targeted Retrieval

## Problem

Agents require only relevant context, not full history.

## Approach

Retrieve filtered context:

```bash
yanzi export --meta role=engineer --limit 10
```

## Result

Agents receive only the context they need.

## Note

Yanzi does not decide relevance. It only filters data.
