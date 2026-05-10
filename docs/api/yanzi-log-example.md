# Yanzi Agent Log

Project: yanz-qa
Exported: 2026-05-10T16:32:22Z
Version: v2.7.0

---

### Capture: 78f83056a420c00df61bcbec84f41e1d

Role: QA-Agent
Timestamp: 2026-05-10T16:22:59.894361Z
Hash: f39734cf814d5ac29b0a8b8450647ed20db3ef36f6276c8f31e5105908bc7e68

**Prompt**
```text
What is the purpose of the yanz-qa project?
```

**Response**
```text
To perform comprehensive QA validation of the Yanzi CLI v2.7.0 release candidate, covering all commands, flags, exports, and documentation accuracy.
```

---

### Capture: 6610ca5d44b96c7037e8940c2984254e

Role: QA-Agent
Timestamp: 2026-05-10T16:23:00.351276Z
Hash: 442cc8b7ea2e8b1c7eadecfc5c013cce061fc9f82421acb25ef772bfa61588f0

Metadata:
  area: validation
  decision_type: test
  profile: qa
  tags: qa,coverage

**Prompt**
```text
Test capture with all optional flags
```

**Response**
```text
This capture tests title, source, profile, and meta flags.
```

---

### Capture: 3a1a137ec731e8505349dbc008cf5293

Role: QA-Agent
Timestamp: 2026-05-10T16:23:05.164828Z
Hash: b161e3936eabb0ebf57c63438210ea81ed42b9fcde22b8c5075b7c0839080b8b

**Prompt**
```text
Test explicit chaining with prev-hash
```

**Response**
```text
This capture is explicitly chained to the previous one.
```

---

### Capture: c0bc5c03055527d8d1333bfa172b6eaf

Role: QA-Agent
Timestamp: 2026-05-10T16:23:12.210599Z
Hash: a999dc746e09ae8bef92a7960b41e9ffa62c9cf74ef51f0b7989d665c538a2e7

Metadata:
  area: cli-coverage

**Prompt**
```text
What are the key CLI commands for Yanzi v2.7.0?

```

**Response**
```text
The key commands are: capture, list, show, verify, chain, delete, restore, project, checkpoint, rehydrate, export, rules, message, context, intent, bootstrap, types, mode, version.

```

---

### Capture: af05d55a68af86218a4e5847ddf793f7

Role: QA-Delete-Test
Timestamp: 2026-05-10T16:23:58.086761Z
Hash: 11f3f6354aab0eb612bde88685c2622f50b74c0e05eaa8e383cf699282ccea2b

**Prompt**
```text
This capture will be deleted
```

**Response**
```text
Delete test response
```

---

## Checkpoint: 9318cbc8f8aa4e945f70fed435be1d227cb584a42af681a2b4e40bd4a658378b

Summary: QA Session Phase 1: Environment and basic command validation
Timestamp: 2026-05-10T16:24:10.097717Z
----------------------

### Capture: d7081fe9639923d27fd4a04099afd195

Role: QA-Agent
Timestamp: 2026-05-10T16:24:19.848108Z
Hash: 4fe758a6b5360923e6a91c243576d3c181bae84f75772ad6e70bb6efc3a10ff9

**Prompt**
```text
Post-checkpoint test: validate rehydration shows this
```

**Response**
```text
This capture should appear in rehydration output as it was created after the checkpoint.
```

---

### Capture: d85c1c6b0c39ca454a6ca9a50edfa56e

Role: human
Timestamp: 2026-05-10T16:25:02.077394Z
Hash: 6ef3e6c6f65de21b49005d8aca84269e065741293f4d995b652fb0304a109ead

Metadata:
  scope: global
  subtype: rules
  type: context

**Prompt**
```text
# QA Rules

## Rule 1: Validate all commands
All CLI commands must be exercised during QA.

## Rule 2: Test error conditions
Error conditions must produce clear, actionable messages.

```

**Response**
```text
Canonical rules
```

---

### Capture: d5581d03d280b8a8ce9ffa843857b382

Role: human
Timestamp: 2026-05-10T16:25:02.089293Z
Hash: a2fba6125d0442cb7d285bcecc33b5566c7733f10a1b9befece47e0d7872aa51

Metadata:
  scope: global
  subtype: rules
  type: context

**Prompt**
```text
Test content

```

**Response**
```text
Canonical rules
```

---

### Capture: 8e323b022458948b04a7bf39a349a314

Role: qa-agent
Timestamp: 2026-05-10T16:27:59.07095Z
Hash: e216a51acab307bf770a1d77c141ce6a6316a5cbadc5ba2eb92e69b7cd4bd106

Metadata:
  channel: handoff
  from: qa-agent
  subtype: message
  to: claude
  type: note

**Prompt**
```text
QA Phase 1 complete. All basic commands validated.
```

**Response**
```text
Message note
```

---

### Capture: 0d437ba33b9f67790e3825c26d1e25a6

Role: operator
Timestamp: 2026-05-10T16:27:59.084424Z
Hash: bc5fc060638ceb886505889c3f01a7092d085c058141b5ea726ae32ac1dfc69a

Metadata:
  channel: execution
  from: operator
  subtype: message
  to: codex
  type: note

**Prompt**
```text
Continue QA validation from phase 2.

```

**Response**
```text
Message note
```

---

### Capture: 5ff523f450f0b06756359c8875eedab8

Role: QA-Agent
Timestamp: 2026-05-10T16:32:22.765708Z
Hash: 2dc8453b7310465c3a2f1bd1333ff03731f45b393f713c5ba64e02365806fa04

**Prompt**
```text
Unicode test: 日本語テスト 中文测试 한국어테스트 Ñoño émoji 🚀✅❌
```

**Response**
```text
Unicode response: αβγδ ← → ↑ ↓ ∞ ≠ ≤ ≥
```

---

### Capture: c9e096ffec624a4aa2668b67d68674d0

Role: QA-Agent
Timestamp: 2026-05-10T16:32:22.77953Z
Hash: 1f1720373df1a9b39beb3aacc6b4bf95c506214b4c4c8d618e5dee5b2312ed01

**Prompt**
```text
Multi-line
prompt
test
across
five
lines
```

**Response**
```text
Multi-line
response
also
across
multiple
lines
```

---

