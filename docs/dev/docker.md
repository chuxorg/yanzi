# Local Development with Docker

## Why

yanzi-ui development requires a running yanzi API. The containerized
dev instance isolates the development build from your locally
installed yanzi binary.

The dev container is exposed on host port 8742 to avoid conflicts
with other services commonly using 8080.

## Start the dev instance

```bash
docker compose up --build
```

yanzi API will be available at http://127.0.0.1:8742

## Stop the dev instance

```bash
docker compose down
```

## Reset the dev corpus

The dev corpus lives in `./yanzi-dev-data/`. To reset:

```bash
docker compose down
rm -rf ./yanzi-dev-data/*
docker compose up --build
```

## yanzi-ui configuration

yanzi-ui `.env.development` points to `http://127.0.0.1:8742`.
No additional configuration needed.

## Verify the connection

With the container running:

```bash
curl http://127.0.0.1:8742/v0/health
```

Expected: health response from yanzi API.

## Port reference

| Host port | Container port | Purpose |
|---|---|---|
| 8742 | 8080 | yanzi dev container |

8080 remains the default for `yanzi serve` without Docker.

## Postgres storage provider

Postgres provider support is coming in CAP-003 Phase 2. When available,
the dev container will support `YANZI_STORAGE_PROVIDER=postgres` to run
against a Postgres backend instead of the default SQLite file.

See [docs/reference/storage-providers.md](../reference/storage-providers.md)
for current provider configuration options.
