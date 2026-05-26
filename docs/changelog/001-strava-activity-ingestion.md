---
id: changelog-001
number: 1
slug: strava-activity-ingestion
related_epics:
  - epic-001
related_adrs:
  - ADR-001
  - ADR-002
  - ADR-003
---

# Changelog 001: Strava Activity Ingestion Tests

## Summary

Added test coverage and structural cleanup around the first Strava ingestion slice. The change focuses on the documented OAuth, token refresh, activity fetch, canonical mapping, deduplication, raw payload storage, and sensitive-data handling expectations.

## Context Sources

- Product and epic context came from the repository documentation in `docs/product`, `docs/epics`, and `docs/adr`.
- Implementation context came from the current Go backend under `cmd/server`.
- Any decisions from conversation should be captured here explicitly, because the changelog is intended to become the durable project memory.

## Related Epics

- [Epic 001: Fetch and Store Strava Activities Locally](../epics/001-strava-activity-ingestion.md)

## Related ADRs

- [ADR-001: Go as Backend Language](../adr/001-go-backend-language.md)
- [ADR-002: Strava OAuth2 Flow](../adr/002-strava-first-provider-oauth2.md)
- [ADR-003: Postgres + Token Storage](../adr/003-postgres-local-first-token-storage.md)

## Relevant Changes

- Added OAuth start-handler tests for successful Strava redirects, missing `STRAVA_CLIENT_ID`, OAuth `state` generation, and state cookie creation.
- Added OAuth callback tests for missing authorization codes, missing URL state, missing state cookies, invalid state values, missing credentials, and successful token exchange.
- Made the Strava token exchange call mockable in tests so callback success and token refresh paths can run without calling the real Strava API.
- Persisted successful OAuth token exchange results through the provider token store, including access token, refresh token, expiry, provider user ID, and granted scopes.
- Added token refresh coverage for still-valid stored tokens, expired tokens, rotated refresh tokens, and missing refresh credentials.
- Added unit tests for Strava activity mapping into the canonical `Activity` model.
- Covered optional Strava fields such as heart rate, elevation gain, and calories.
- Added validation tests for required Strava activity fields.
- Added normalization tests for Strava activity types such as run, ride, swim, walk, workout, yoga, unknown, and fallback values.
- Added handler error-path tests for missing tokens, Strava API failures, invalid Strava JSON, raw object storage failure, mapping failure, activity save failure, and activity list failure.
- Added a regression test that activity fetch errors do not expose token values in HTTP responses.
- Added typed handling for non-200 Strava activity responses via `HTTPStatusError`, preserving the upstream Strava status code for handler-level decisions.
- Added explicit handling for Strava `429 Too Many Requests`, returning `429` to the caller instead of collapsing the case into a generic bad gateway response.
- Added opt-in Postgres integration tests for provider token upsert, activity deduplication, activity listing, and migration idempotency.
- Added an opt-in MinIO integration test for raw payload storage metadata.
- Moved persistence startup concerns out of `main.go` into a dedicated database bootstrap path, covering database connection, migrations, token store wiring, and optional MinIO raw object store wiring.
- Moved Strava response DTOs out of `main.go` so the entrypoint stays focused on configuration, persistence setup, route registration, and server startup.
- Split Strava OAuth/token code from Strava athlete and activity handler code:
  - `strava_auth.go` now focuses on OAuth state, auth redirect, callback handling, token exchange, token refresh, and valid-token lookup.
  - `strava_client.go` contains shared Strava API client configuration.
  - `strava_athlete.go` contains the Athlete DTO, athlete fetch, and athlete HTTP handler.
  - `strava_activity_handlers.go` contains activity ingestion HTTP handler orchestration.
- Reorganized tests to mirror the production file responsibilities:
  - `strava_auth_test.go` covers OAuth and token behavior.
  - `strava_activities_test.go` covers Strava activity mapping, normalization, and object-key behavior.
  - `strava_activity_handlers_test.go` covers activity handler orchestration and error paths.
  - `strava_athlete_test.go` covers athlete fetch and athlete handler behavior with local `httptest` servers.
  - `strava_athlete_smoke_test.go` keeps the optional real Strava smoke test separate.
  - `test_fakes_test.go` contains shared test fakes for token, raw object, and activity stores.
- Added separate local-list and sync handlers:
  - `GET /api/activities` returns locally stored activities.
  - `POST /api/activities/sync` triggers request-driven Strava ingestion, raw payload storage, canonical mapping, deduplication, and activity listing.

## Decisions

- Keep normal `go test ./...` infrastructure-free so local and CI runs remain fast and reliable.
- Treat request-triggered syncing via `POST /api/activities/sync` as sufficient for this feature; background jobs, scheduled sync, or automatic periodic fetching are intentionally out of scope for this slice.
- Keep `GET /api/activities` local-only so listing stored activities does not call Strava or consume Strava rate limits.
- Use method-specific route patterns for activity endpoints so local listing is `GET /api/activities` and sync side effects are triggered through `POST /api/activities/sync`.
- Use typed errors only for real upstream Strava HTTP responses. Internal errors such as request construction, network failures, response reading, or JSON decoding remain normal Go errors and are mapped to generic gateway failures by the handler.
- Gate Postgres integration tests behind `FORMPATH_DB_TEST=1`.
- Gate MinIO raw object storage tests behind `FORMPATH_S3_TEST=1`.
- Use the existing Go test stack and avoid introducing a mocking or container orchestration dependency for now.
- Treat the changelog as merge-level project memory, not a generated list of commits.
- Keep `main.go` as a thin composition root and place infrastructure initialization beside the storage/migration code instead of embedding it in the HTTP server entrypoint.

## Verification

```bash
go test ./...
```

The default test suite passed.

## Follow-ups

- Consider introducing a small activity service layer before implementing more endpoint logic. The HTTP handlers should stay focused on request/response handling, while the service owns local listing and Strava sync orchestration.
- Consider defining a narrow `StravaClient` interface for the service, for example `FetchActivities(ctx, accessToken)`, so sync logic can depend on an abstraction instead of directly calling global Strava HTTP functions. This would make local-vs-online behavior and tests clearer without introducing a heavy dependency container.
- Consider moving Strava-specific code into its own Go package later, for example under `internal/strava`, once the app has clearer package boundaries for domain models, storage interfaces, and handler dependencies. Do not do this as a blind file move, because a subdirectory is a separate Go package and would require explicit exported APIs.
- Extend Strava rate-limit handling beyond the current `429 Too Many Requests` response mapping. Useful next steps include forwarding or honoring `Retry-After`, capturing Strava rate-limit headers, and adding explicit tests for retry/rate-limit metadata.
- Make token refresh race-safe before multi-request or multi-instance usage matters. ADR-003 calls out idempotent refresh behavior; the current flow loads, refreshes, and saves without a compare-and-swap update, lock, or other concurrency guard.
- Improve Strava permission and scope handling. The current token persistence records configured scopes, but should eventually validate and store the scopes actually granted by Strava, especially for `activity:read_all`.
- Run the opt-in Postgres tests against a disposable database before merging storage-heavy changes.
- Run the opt-in MinIO test when changing raw object storage behavior.
- Add one changelog file for each future merged feature or meaningful fix.
