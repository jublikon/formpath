# ADR-003: Local-First Postgres and MinIO Storage

## Status

Accepted

## Context

Formpath needs persistent storage for canonical data, user tokens, and raw provider payloads. The system should run local-first, but be deployable to AWS later.

Token handling has two layers:
- **App credentials** (client_id, client_secret): static, per environment, secret
- **User tokens** (access_token, refresh_token, expires_at): dynamic, constantly rotating, part of user state

Storage alternatives: SQLite (simpler locally, but a weaker AWS path), Postgres (a little more setup, but identical locally and in production).

Raw provider data has a separate lifecycle from canonical Formpath data. It should be kept immutable enough for debugging and future reprocessing, while the database stores queryable canonical records and object metadata. Locally, MinIO gives us the same S3-shaped object storage model that production can use later.

## Decision

1. **Postgres** as the primary database, locally via Docker Compose.
2. **App credentials** (`STRAVA_CLIENT_ID`, `STRAVA_CLIENT_SECRET`) are configured exclusively through environment variables. They never appear in code, the DB, or logs. A `.env` file (gitignored) documents the format.
3. **User tokens** are stored in Postgres. The backend refreshes tokens automatically before they expire and stores the new refresh token.
4. **Raw provider payloads** are stored in MinIO locally, with metadata and canonical references stored in Postgres.
5. **No encryption at rest** in the local phase — Postgres and MinIO run locally, and the effort only becomes worthwhile at deployment. This is a deliberate simplification for local-first.

## Consequences

- Docker Compose becomes the local starting point (`docker compose up` starts Postgres, MinIO, and the backend)
- `.env.example` documents required variables, `.env` is gitignored
- Token table needs: user_id, provider, access_token, refresh_token, expires_at, scopes, updated_at
- Raw object metadata needs: user_id, provider, object type, provider object id, object key, checksum, content type, size, fetched_at
- Canonical activity records should reference raw object keys instead of embedding large raw payloads directly
- Refresh logic must be idempotent (watch for race conditions with parallel requests)
- Local Postgres = Postgres on AWS (RDS) — no storage change is needed for deployment
- Local MinIO = S3-compatible object storage — production can move to S3 without changing the raw storage contract
- Encryption at rest will be added as a separate ADR when deployment is coming up
