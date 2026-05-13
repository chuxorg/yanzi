# Demo Flow

Use this sequence for a reliable Yanzi demo.

## Story

1. create a project
2. capture working context
3. create a checkpoint
4. add more work after the checkpoint
5. simulate interruption
6. run `yanzi rehydrate`
7. resume from the rendered continuity

## Install Check

Verify the CLI first:

```bash
yanzi --version
```

If it is missing on macOS or Linux:

```bash
curl -fL -o /tmp/yanzi-install.sh https://raw.githubusercontent.com/chuxorg/yanzi/main/install.sh
test -s /tmp/yanzi-install.sh
sh /tmp/yanzi-install.sh
```

## Recommended Live Flow

Create and select a project:

```bash
yanzi project create demo-flow
yanzi project use demo-flow
```

Inline capture:

```bash
yanzi capture \
  --author "Ada" \
  --title "Auth planning" \
  --prompt "What do we need to verify in the reconnect flow?" \
  --response "Check refresh token expiry, clock skew, and retry handling." \
  --meta area=auth \
  --meta phase=planning
```

stdin capture:

```bash
echo "Need to validate refresh token edge cases after reconnect." \
  | yanzi capture --author "Ada" --title "Reconnect edge cases" \
    --response "Focus on leeway handling and stale local clocks." \
    --meta area=auth --meta phase=investigation
```

File capture:

```bash
yanzi capture \
  --author "Ada" \
  --title "Auth notes" \
  --prompt-file prompt.txt \
  --response-file response.txt \
  --meta area=auth --meta phase=notes
```

Checkpoint:

```bash
yanzi checkpoint create --summary "Auth reconnect baseline captured"
```

Post-checkpoint capture:

```bash
echo "Need to resume from the last stable auth checkpoint." \
  | yanzi capture --author "Ada" --title "Resume note" \
    --response "Rehydrate should show the reconnect investigation and next step." \
    --meta area=auth --meta phase=resume
```

Recovery:

```bash
yanzi rehydrate
yanzi rehydrate --format json
```

Exports:

```bash
yanzi export --format markdown
yanzi export --format html
```

## Operator Notes

- stdin currently supports prompt-side ingestion only
- responses remain explicit with `--response` or `--response-file`
- this is intentional so capture input stays deterministic
- `yanzi list` scopes to the active project by default
- `yanzi rehydrate` renders captures in chronological order from the last checkpoint forward
- `yanzi rehydrate --format json` is the structured recovery output for automation
