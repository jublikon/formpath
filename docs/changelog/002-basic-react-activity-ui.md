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

Added the first browser-based Formpath user interface as a React and TypeScript application. The UI completes a thin local end-to-end read path by loading canonical activities from the existing Go API and presenting them in a responsive activity list.

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
- Added a typed `Activity` representation matching the JSON fields returned by the Go API.
- Added loading, error, empty, and populated states for the activity list.
- Added a real `GET /api/activities` request when the application is mounted.
- Normalized a possible `null` activity response to an empty array in the UI.
- Added a responsive two-column activity layout that becomes single-column on narrow screens.
- Kept the Go backend and persistence implementation unchanged.

## Decisions

- Use React, TypeScript, and Vite as the frontend foundation because the intended product will need interactive dashboards, visualizations, and agentic workflows.
- Keep the frontend as a separate application under `web/` while retaining the Go backend under `cmd/server`.
- Use the Vite proxy during local development so browser requests can use relative `/api` URLs without adding development-only CORS behavior to the Go backend.
- Start with a read-only activity listing instead of introducing routing, global state management, a component library, or an additional data-fetching library.
- Keep API field names in their current `snake_case` form in the first TypeScript model so the frontend representation directly matches the backend JSON contract.
- Treat runtime API validation as deferred work. The current TypeScript type documents the expected response shape but does not validate untrusted JSON at runtime.
- Defer production delivery of the built frontend. The Vite proxy only applies during local development.

## Verification

```bash
cd web
nvm use
npm run build
```

The TypeScript check and Vite production build passed.

The local end-to-end flow was also verified manually with Postgres and MinIO running through Docker Compose, the Go backend on port `8080`, and the Vite development server on port `5173`. The browser displayed activities stored by the existing Strava ingestion flow.

## Follow-ups

- Add a user-triggered activity sync action backed by `POST /api/activities/sync`.
- Add a Strava connection entry point backed by `/auth/strava`.
- Format timestamps, distances, durations, and activity types for human-readable display.
- Decide how the production frontend build will be served, for example by the Go server or separate static hosting.
- Add frontend tests as interaction and data transformation logic grows.
- Remove unused Vite scaffold assets and replace the generic page metadata and frontend README with Formpath-specific content.
