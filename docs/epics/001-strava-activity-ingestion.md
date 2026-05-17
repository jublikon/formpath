# Epic 001: Fetch and Store Strava Activities Locally

## Goal

As a user, I can connect my Strava account, fetch my activities into Formpath by calling the backend activity endpoint, store canonical records locally in Postgres, and store raw provider payloads locally in MinIO.

## Why

This is the first vertical slice through the entire system: OAuth integration, API fetch, raw payload storage, normalization, and persistence. After this, a working backend exists that all further features can build on.

## Scope

### In Scope

- Go backend with HTTP server
- Strava OAuth2 flow (authorize → callback → token exchange)
- Token persistence in Postgres (access_token, refresh_token, expires_at)
- Automatic token refresh when expired
- Fetching the activity list from Strava (`GET /athlete/activities`)
- Request-triggered ingestion via `GET /api/activities`; calling this endpoint is sufficient for this feature
- Storage as canonical activity records in Postgres
- Storage of raw Strava payloads in MinIO, with object metadata in Postgres
- Docker Compose setup (Postgres + MinIO + backend)
- `.env.example` with required variables

### Out of Scope

- UI / Frontend
- Historical bulk import (pagination over all old activities)
- Background jobs, scheduled sync, or automatic periodic activity fetching
- Additional providers besides Strava
- Detailed data for individual activities (streams, laps, splits)
- Encryption at rest
- Deployment / AWS

## Acceptance Criteria

1. `docker compose up` starts Postgres, MinIO, and the backend without manual steps
2. Calling `http://localhost:8080/auth/strava` correctly redirects to Strava
3. After the OAuth callback, tokens are stored in Postgres
4. A `GET /api/activities` endpoint fetches activities from Strava on request, stores them, and returns the stored activities as JSON
5. When the token has expired, the backend refreshes it automatically and stores the new token
6. Activities are deduplicated (calling `GET /api/activities` again does not create duplicates)
7. Strava-specific fields are mapped to a canonical activity model
8. Sensitive data (tokens, client_secret) does not appear in logs

## Canonical Activity Model (Draft)

| Field | Type | Source |
|------|-----|--------|
| id | UUID | generated |
| provider | string | "strava" |
| provider_id | string | Strava Activity ID |
| user_id | UUID | internal user |
| activity_type | string | normalized (run, ride, swim, ...) |
| name | string | Strava Activity Name |
| started_at | timestamp (UTC) | start_date |
| duration_seconds | int | elapsed_time |
| moving_seconds | int | moving_time |
| distance_meters | float | distance |
| elevation_gain_meters | float | total_elevation_gain |
| average_heartrate_bpm | float | average_heartrate (optional) |
| max_heartrate_bpm | float | max_heartrate (optional) |
| calories_kcal | float | calories (optional) |
| raw_object_key | text | MinIO/S3 object key for the complete Strava response |
| created_at | timestamp (UTC) | ingestion time |
| updated_at | timestamp (UTC) | last update |

## Technical Notes

- Respect rate limits: 200 requests/15min, 2000/day
- Strava returns a maximum of 200 activities per page — one page is enough for this slice
- This slice does not need a background worker or scheduled sync; request-triggered fetching is the intended behavior
- Raw Strava responses are stored as MinIO objects for later reprocessing
- Postgres stores raw object metadata and canonical references
- Deduplicate via a `(provider, provider_id)` unique constraint

## Milestones

1. **Project Structure + Docker Compose** — Go project, Postgres, MinIO, `docker compose up` works
2. **OAuth2 Flow** — authorize, callback, token in DB
3. **Token Refresh** — automatic when the token has expired
4. **Activity Fetch** — fetch Strava API on request, map canonically, store in Postgres
5. **API Endpoint** — `GET /api/activities` triggers the fetch and returns stored activities as JSON

## Relevant ADRs

- [ADR-001: Go as Backend Language](../adr/001-go-backend-language.md)
- [ADR-002: Strava OAuth2 Flow](../adr/002-strava-first-provider-oauth2.md)
- [ADR-003: Postgres + Token Storage](../adr/003-postgres-local-first-token-storage.md)
