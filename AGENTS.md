# Formpath Agent Guide

## Project overview

Formpath is a local-first training application for ambitious recreational athletes. The current product imports Strava activities, stores canonical activity data in Postgres and raw provider payloads in MinIO, and presents a React training overview. The backend is written in Go; the frontend uses React, TypeScript, and Vite.

## Repository layout

- `cmd/server/`: Go HTTP server, Strava integration, storage, migrations, and backend tests.
- `migrations/`: Postgres schema migrations applied by the backend.
- `web/`: React/TypeScript frontend, Vite configuration, ESLint configuration, and Vitest tests.
- `docs/epics/`: product slices and acceptance criteria.
- `docs/adr/`: accepted architectural decisions.
- `docs/changelog/`: durable records of merged changes and their verification.
- `docker-compose.yml`: local Postgres, MinIO, and backend stack.

## Development commands

- `docker compose up --build`: start the full local backend stack. It requires a local `.env` based on `.env.example`.
- `go run ./cmd/server`: run the backend directly with localhost-oriented environment values.
- `cd web && npm ci`: install the locked frontend dependencies.
- `cd web && npm run dev`: start the Vite development server; `/api` and `/auth` are proxied to `http://localhost:8080`.
- `cd web && npm run build`: type-check and create the production frontend build.
- `cd web && npm run lint`: run ESLint.

## Test commands

- `go test ./...`: run the normal, infrastructure-free Go test suite.
- `cd web && npm test`: run frontend unit tests once with Vitest.
- Postgres tests require a running test database and `FORMPATH_DB_TEST=1`.
- MinIO tests require a running test service and `FORMPATH_S3_TEST=1`.
- The real Strava smoke test is opt-in with `STRAVA_SMOKE_TEST=1` and uses local credentials; do not run it unless explicitly needed.

## Architecture principles

- Prefer simple, explicit, idiomatic Go and small React/TypeScript modules over abstraction-heavy designs.
- Build thin vertical slices that work through the real product flow.
- Keep provider-specific adapters separate from canonical activity data and provider-neutral UI calculations.
- Keep `main.go` a thin composition root and HTTP handlers focused on transport concerns.
- Store canonical, queryable records in Postgres and raw provider payloads in MinIO/S3-compatible storage.
- Keep ingestion idempotent and deduplication rules explicit.
- Optimize for local development and debuggability without introducing premature cloud-specific infrastructure.
- Do not add major dependencies, infrastructure, or schema changes without clear justification.

## Testing expectations

- Run the smallest relevant tests while iterating, then the broader affected suites before completion.
- Add deterministic tests for changed logic, especially normalization, aggregation, date/time behavior, persistence boundaries, deduplication, and error paths.
- Keep `go test ./...` independent of external services; gate infrastructure and real-provider tests behind their existing environment flags.
- For UI changes, validate loading, disconnected, empty, syncing, success, and failure states as applicable, including responsive and accessible behavior.
- Never use real personal activity data, tokens, or credentials in fixtures or snapshots.

## Documentation expectations

- Update the README when local setup or standard commands change.
- Update an epic when scope or acceptance criteria change.
- Add or update an ADR for durable architectural decisions.
- Add a changelog entry for a merged feature or meaningful fix, including decisions, verification, and follow-ups.
- Keep documentation self-contained; do not rely on private conversations for project context.

## Security and secrets

- Never edit or commit `.env`, credentials, tokens, personal activity data, or local-only configuration.
- Use `.env.example` only to document variable names and safe local defaults.
- Do not expose OAuth tokens, client secrets, raw provider payloads, or internal errors in logs, responses, fixtures, screenshots, or documentation.
- Treat health, training, location, and user-adjacent data as sensitive; use synthetic or anonymized examples.

## Definition of done

- The requested behavior is implemented with a focused, reviewable diff.
- Relevant Go and/or frontend tests pass; lint and build checks pass for frontend changes.
- External-service tests are run when the touched behavior requires them, or the omission is stated.
- User-visible flows and important error states are verified when applicable.
- Documentation is updated when behavior, setup, architecture, or project history changes.
- No secrets, generated build output, unrelated formatting, or unrelated refactors are included.
- Changed files, validation performed, assumptions, risks, and follow-ups are summarized.

## Agent working style

- Explore relevant code, tests, and documentation before editing.
- Make the smallest safe assumption when details are ambiguous and state it.
- Prefer small, reviewable diffs and avoid broad cleanup unrelated to the task.
- Preserve existing conventions and user changes.
- Validate incrementally and report commands that were actually run.

## Traceability

For non-trivial changes, briefly explain the next step and its purpose, keep diffs small, name changed files, and summarize validation performed. Do not expose private chain-of-thought; provide concise, human-readable plans, decisions, and validation results.
