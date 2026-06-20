# Formpath Agent Guide

## Project overview

Do not maintain a separate product description here. Use these sources of truth:

- [Product vision](docs/product/vision.md) for the long-term purpose and intended users.
- [Product roadmap](docs/product/roadmap.md) for current, completed, and candidate work.
- [Epics](docs/epics/) for feature scope, acceptance criteria, and explicit exclusions.
- [Changelog](docs/changelog/) for what has actually been implemented and verified.
- [README](README.md) for the documentation index and local setup.

Before making product or architecture assumptions, read the relevant documents above and verify them against the current code.

## Repository layout

- [`cmd/server/`](cmd/server/) contains the backend entry point, implementation, and colocated tests.
- [`migrations/`](migrations/) contains database migrations.
- [`web/`](web/) contains the frontend; use [`web/package.json`](web/package.json) as the command source of truth.
- [`docs/product/`](docs/product/) contains vision and roadmap documents.
- [`docs/epics/`](docs/epics/) contains scoped product work and acceptance criteria.
- [`docs/adr/`](docs/adr/) contains accepted architectural decisions.
- [`docs/changelog/`](docs/changelog/) contains durable records of delivered changes; follow its [README](docs/changelog/README.md).
- [`docker-compose.yml`](docker-compose.yml), [`Dockerfile`](Dockerfile), and [`.env.example`](.env.example) define the visible local runtime setup.

## Development commands

Use the commands documented or configured in the repository; do not invent a parallel command layer.

- Follow [README local development](README.md#local-development) for environment setup and the full local stack.
- Use `docker compose up --build` for the Compose setup defined in [`docker-compose.yml`](docker-compose.yml).
- Use `go run ./cmd/server` for direct backend execution as described in the [README](README.md#local-development).
- In `web/`, use `npm ci`, `npm run dev`, `npm run build`, and `npm run lint`; verify scripts in [`web/package.json`](web/package.json).
- Check [`web/vite.config.ts`](web/vite.config.ts) before making assumptions about development proxy behavior.

## Test commands

- Run `go test ./...` for the normal Go suite. Its infrastructure-free intent is recorded in [Changelog 001](docs/changelog/001-strava-activity-ingestion.md#decisions).
- Run `cd web && npm test` for frontend tests; use [`web/package.json`](web/package.json) for the canonical script.
- Read [`cmd/server/storage_integration_test.go`](cmd/server/storage_integration_test.go) before running opt-in Postgres or MinIO tests.
- Read [`cmd/server/strava_athlete_smoke_test.go`](cmd/server/strava_athlete_smoke_test.go) before running the real-provider smoke test.
- Do not enable tests that use infrastructure, credentials, or external providers unless the changed surface requires them.

## Architecture principles

Architecture is documented in accepted decisions and feature documents. Read the relevant source before changing a boundary:

- [ADR-001](docs/adr/001-go-backend-language.md): backend language decision.
- [ADR-002](docs/adr/002-strava-first-provider-oauth2.md): first provider, OAuth flow, and provider-boundary consequences.
- [ADR-003](docs/adr/003-postgres-local-first-token-storage.md): persistence, raw storage, secrets, and local-first decisions.
- [Epic 001](docs/epics/001-strava-activity-ingestion.md): ingestion slice and canonical activity expectations.
- [Epic 002](docs/epics/002-training-overview.md): training-overview calculations, UI boundaries, and exclusions.
- [Changelog entries](docs/changelog/) for later implementation decisions that refine the epics and ADRs.

Do not restate those decisions here. Preserve them unless the task explicitly changes them; record durable changes in the appropriate ADR, epic, and changelog.

## Testing expectations

- Run the smallest relevant tests while iterating, then the broader affected suites before completion.
- Derive required scenarios from the acceptance criteria in the relevant [epic](docs/epics/) and from nearby existing tests.
- Follow the established backend test organization described in [Changelog 001](docs/changelog/001-strava-activity-ingestion.md#relevant-changes).
- Follow the frontend verification precedent in [Changelog 003](docs/changelog/003-training-overview.md#verification) when the touched UI surface makes it relevant.
- Keep the default test suite independent of external services; preserve the opt-in guards visible in the integration and smoke tests.
- Never use real personal activity data, tokens, or credentials in fixtures or snapshots.

## Documentation expectations

- Use [README](README.md) for local setup and the top-level documentation index.
- Use [product documents](docs/product/) for vision and roadmap changes.
- Use [epics](docs/epics/) for feature scope, acceptance criteria, and exclusions.
- Use [ADRs](docs/adr/) for durable architectural decisions.
- Follow the [changelog guide](docs/changelog/README.md) for delivered features and meaningful fixes.
- Link to the canonical document instead of duplicating descriptions in agent instructions.
- Keep repository documentation self-contained; do not rely on private conversations for project context.

## Security and secrets

- Never edit or commit `.env`, credentials, tokens, personal activity data, or local-only configuration.
- Use [`.env.example`](.env.example) only as the tracked environment-variable template; confirm ignore behavior in [`.gitignore`](.gitignore).
- Follow the credential, token, and raw-storage decisions in [ADR-003](docs/adr/003-postgres-local-first-token-storage.md).
- Follow the non-exposure acceptance criteria in [Epic 001](docs/epics/001-strava-activity-ingestion.md#acceptance-criteria).
- Use synthetic or anonymized examples in tests, screenshots, and documentation.

## Definition of done

- The requested behavior is implemented with a focused, reviewable diff.
- Relevant commands from [Test commands](#test-commands) pass.
- Acceptance criteria in the relevant [epic](docs/epics/) are satisfied or explicitly identified as out of scope.
- External-service tests are run when the touched behavior requires them, or the omission is stated.
- Documentation is updated in its canonical location when behavior, setup, scope, architecture, or project history changes.
- No secrets, generated build output, unrelated formatting, or unrelated refactors are included.
- Changed files, validation performed, assumptions, risks, and follow-ups are summarized.

## Agent working style

- Start with [README](README.md), then read the relevant product document, epic, ADR, changelog, code, and tests before editing.
- Make the smallest safe assumption when details are ambiguous and state it.
- Prefer small, reviewable diffs and avoid broad cleanup unrelated to the task.
- Preserve existing conventions and user changes.
- Validate incrementally and report commands that were actually run.
- Never use `codex`, `claude`, or another coding-agent name in a branch name unless the user explicitly requests it.
- Keep project history tool-neutral. Commit messages, pull request titles and descriptions, changelogs, and similar records must describe the change without stating or implying that a coding agent created it.

## Traceability

For non-trivial changes, briefly explain the next step and its purpose, keep diffs small, name changed files, and summarize validation performed. Do not expose private chain-of-thought; provide concise, human-readable plans, decisions, and validation results.
