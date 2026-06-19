---
id: epic-002
number: 2
slug: training-overview
related_adrs: []
related_changelogs:
  - changelog-003
---

# Epic 002: Training Overview

## Goal

As a user, I can open Formpath and immediately understand my recent training volume, how it has developed over the last four weeks, and which activities contributed to it.

## Why

Epic 001 established the first end-to-end data flow from Strava into Formpath. The current UI proves that the imported activities are available, but it presents them primarily as raw records.

This epic turns the existing data foundation into the first useful product experience. It should help a user answer three questions without leaving the page:

1. How much have I trained recently?
2. How has my training volume changed over the last four weeks?
3. What were my latest activities?

The overview is part of the real Formpath application. It does not use a separate demo mode, seeded showcase data, or provider-specific presentation logic.

## Product Shape

The first application experience is a focused single-page dashboard.

The page contains:

1. A compact Formpath header and the active period
2. Strava connection and sync status
3. Summary metrics for the last four weeks
4. A daily training-volume curve
5. A list of recent activities

This is an application dashboard, not a public marketing landing page. Additional navigation and routes should only be introduced when Formpath has distinct product areas that need them.

## Scope

### In Scope

- Use activities already stored in Formpath as the only data source for the overview
- Show summary metrics for the last four weeks:
  - number of activities
  - total distance
  - total moving time
  - total elevation gain when available
- Show daily training volume as a line and area chart for the same four-week period
- Use moving time as the primary cross-sport volume metric
- Show recent activities below the overview
- Format dates, distances, durations, elevation, and activity types for humans
- Retain the existing Strava connect and manual sync flow
- Refresh the overview and activity list after a successful sync
- Provide useful loading, disconnected, empty, error, and syncing states
- Provide a responsive layout that works on desktop and mobile
- Keep calculations provider-neutral and based on the canonical activity model

### Out of Scope

- Demo data, seeded showcase activities, or a separate demo edition
- Public access to a user's training data
- Multi-user authentication and authorization
- Additional data providers
- Activity detail pages
- Interactive date-range selection
- Sport-specific dashboards or comparisons
- Training load, readiness, recovery, fitness, or fatigue scores
- AI-generated analysis, coaching, or recommendations
- Goal and training-plan management
- Background or scheduled activity synchronization
- Production deployment

## Time and Aggregation Rules

- The overview covers today and the preceding 27 days.
- The visualization groups those activities into 28 daily values, including zero-value days.
- Activity inclusion is based on `started_at`.
- Summary metrics and daily chart values use the same included activity set.
- Moving time is preferred over elapsed duration for training-volume calculations.
- Missing optional values such as elevation gain do not cause an activity to be excluded.
- Formatting and period labels use the user's browser locale.

These rules intentionally avoid introducing user profile and timezone settings in this epic. Calendar-week and configurable-period semantics can be added when those concepts exist in the product model.

## Acceptance Criteria

1. Opening Formpath shows a single-page training overview rather than only a raw activity list.
2. The overview is calculated exclusively from activities returned by `GET /api/activities`.
3. The page shows activity count, total distance, total moving time, and available elevation gain for the last 28 days.
4. The page shows a 28-day training-volume curve derived from daily moving time.
5. Summary metrics and the visualization update after a successful Strava sync without a page reload.
6. Recent activities show human-readable dates, distances, durations, elevation, and activity labels.
7. A disconnected user is prompted to connect Strava.
8. A connected user with no stored activities is prompted to sync activities.
9. Loading, API failure, and sync failure states are understandable and do not expose sensitive or internal data.
10. The layout remains usable at mobile and desktop widths.
11. The production UI does not contain hard-coded or seeded activity data.
12. Frontend build, linting, and backend tests pass.

## Technical Notes

- Keep the existing React application component-light while the product remains a single page.
- Extract formatting and aggregation into small testable frontend modules rather than embedding calculations in JSX.
- Render the chart as a custom responsive SVG owned by the Formpath React UI.
- Use `d3-shape` with `curveMonotoneX` to generate the line and area paths without introducing a full charting component library.
- The curve represents the actual daily values. It must not be presented as a rolling average or predictive trend.
- Use semantic text alongside the visualization so its period and values are understandable without relying only on graphics.
- Continue using the canonical activity API response. A dedicated aggregate backend endpoint is not required for this data volume and fixed period.
- Avoid exposing provider-specific concepts in overview calculations so future providers can contribute activities without changing the UI model.

## Milestones

1. **Overview Model** — formatting and four-week aggregation functions with tests
2. **Summary Metrics** — responsive metric cards populated from real activities
3. **Daily Volume Curve** — accessible 28-day moving-time visualization using a responsive SVG and `d3-shape`
4. **Recent Activities** — polished, human-readable activity list
5. **Application States** — connected, disconnected, empty, loading, syncing, and error flows
6. **Visual QA** — desktop and mobile verification using real imported data

## Follow-ups

- Introduce application navigation when activity details, goals, integrations, or settings become separate product areas.
- Add an optional rolling seven-day trend line when Formpath supports richer chart interpretation and comparison.
- Decide whether the next product slice should add activity details or a first actionable training insight.
- Add authentication and authorization before making personal Formpath data reachable on a public deployment.
- Revisit server-side aggregation when activity volume, configurable periods, or multi-user workloads make client-side aggregation inefficient.
