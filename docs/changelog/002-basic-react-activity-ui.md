---
id: changelog-002
number: 2
slug: basic-react-activity-ui
related_epics:
  - epic-001
related_adrs: []
---

# Changelog 002: Basic React Activity UI

## Summary

Added the first browser-based Formpath user interface as a React and TypeScript application. The UI completes a thin local end-to-end flow by letting a local user connect Strava, return to the UI after OAuth, sync activities, and view locally stored canonical activities.

## Context Sources

- The existing Strava ingestion slice provides the canonical activity data and `GET /api/activities` endpoint.
- The product direction calls for an interactive application layer on top of the data platform.
- The initial UI intentionally focuses on displaying real local data before adding broader frontend architecture or more complex workflows.

## Related Epics

- [Epic 001: Fetch and Store Strava Activities Locally](../epics/001-strava-activity-ingestion.md) provides the backend flow consumed by this UI.

## Related ADRs

- No new architectural decision record was introduced for this initial UI slice.

## Relevant Changes

- Added a React 19 and TypeScript frontend under `web/`, scaffolded with Vite.
- Added a repository-level `.nvmrc` selecting Node.js 24 for frontend development.
- Added a Vite development proxy that forwards `/api` requests to the Go backend at `http://localhost:8080`.
- Extended the Vite development proxy to forward `/auth` requests to the Go backend so the React UI can start the existing Strava OAuth flow locally.
- Added a typed `Activity` representation matching the JSON fields returned by the Go API.
- Added a typed Strava integration status representation matching `GET /api/integrations/strava`.
- Added loading, error, empty, and populated states for the activity list.
- Added initial `GET /api/activities` and `GET /api/integrations/strava` requests when the application is mounted.
- Normalized a possible `null` activity response to an empty array in the UI.
- Added a responsive two-column activity layout that becomes single-column on narrow screens.
- Added `GET /api/integrations/strava`, a backend status endpoint that reports whether the local user has a refreshable Strava connection without exposing tokens.
- Added tests for connected, disconnected, and token-store failure responses from the Strava integration status endpoint.
- Added a compact Strava integration bar that shows either a `Connect with Strava` action or a `Sync activities` action based on backend connection status.
- Added a `POST /api/activities/sync` action from the React UI, including syncing state, error handling, and activity-list refresh after a successful sync.
- Added `FRONTEND_URL` backend configuration and changed successful Strava OAuth callbacks to redirect back to the React UI with a non-sensitive `?strava=connected` status parameter.
- Added UI feedback for the successful Strava OAuth return and removed the status parameter from the browser URL after it has been read.

## Decisions

- Use React, TypeScript, and Vite as the frontend foundation because the intended product will need interactive dashboards, visualizations, and agentic workflows.
- Keep the frontend as a separate application under `web/` while retaining the Go backend under `cmd/server`.
- Use the Vite proxy during local development so browser requests can use relative `/api` URLs without adding development-only CORS behavior to the Go backend.
- Route `/auth` through the same Vite development proxy so the UI can link to `/auth/strava` without hardcoding the backend origin in React code.
- Define Strava integration status as the presence of a stored refresh token, not as a currently valid access token. Expired access tokens still count as connected when they can be refreshed by the backend.
- Keep the integration status endpoint local and lightweight. It checks stored provider tokens only and does not call Strava or refresh tokens.
- Never return OAuth tokens from integration status responses. The UI receives only connection state.
- Keep this UI slice single-screen and component-light instead of introducing routing, global state management, a component library, or an additional data-fetching library.
- Keep API field names in their current `snake_case` form in the first TypeScript model so the frontend representation directly matches the backend JSON contract.
- Treat runtime API validation as deferred work. The current TypeScript type documents the expected response shape but does not validate untrusted JSON at runtime.
- Use a query parameter only as a non-sensitive UI signal after OAuth. Tokens and detailed OAuth response data stay server-side.
- Defer production delivery of the built frontend. The Vite proxy only applies during local development.

## Verification

```bash
cd web
nvm use
npm run build
```

The TypeScript check and Vite production build passed.

```bash
go test ./cmd/server
```

The Go backend test suite passed, including coverage for `GET /api/integrations/strava`.

The local end-to-end flow was also verified manually with Postgres and MinIO running through Docker Compose, the Go backend on port `8080`, and the Vite development server on port `5173`. The browser displayed activities stored by the existing Strava ingestion flow, and the UI now has the controls needed to connect Strava and trigger a sync locally.

## Follow-ups

- Format timestamps, distances, durations, and activity types for human-readable display.
- Decide how the production frontend build will be served, for example by the Go server or separate static hosting.
- Add frontend tests as interaction and data transformation logic grows.
- Remove unused Vite scaffold assets and replace the generic page metadata and frontend README with Formpath-specific content.
