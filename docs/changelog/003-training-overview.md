---
id: changelog-003
number: 3
slug: training-overview
related_epics:
  - epic-002
related_adrs: []
---

# Changelog 003: Training Overview

## Summary

Turned locally stored canonical activities into the first useful Formpath
training overview. The React application now presents four-week summary
metrics, a daily training-volume curve, recent activities, and clear
connection, loading, empty, syncing, and error states.

## Related Epics

- [Epic 002: Training Overview](../epics/002-training-overview.md)

## Related ADRs

- No architectural decision record currently applies to this slice.

## Relevant Changes

- Added a fixed 28-day overview model that groups activities into local
  calendar days, includes zero-value days, and calculates activity count,
  distance, moving time, and available elevation gain.
- Added unit-tested formatting for dates, periods, distances, durations,
  elevation, and canonical activity types.
- Standardized numeric output on a decimal point while leaving date formatting
  locale-aware.
- Added responsive summary metrics calculated from real canonical activities.
- Added a custom responsive SVG line and area chart for daily moving time.
- Added `d3-shape` and used `curveMonotoneX` to generate the chart paths.
- Added accessible chart naming, a text description, and a visually hidden
  list of all 28 daily values.
- Added a Strava integration panel with connected, disconnected, and syncing
  behavior.
- Added understandable loading, initial-load error, sync-error, and empty
  states without exposing technical response details.
- Added a recent-activities section that initially shows ten activities and
  can expand to the complete locally loaded list.
- Added human-readable activity dates, distances, moving times, elevation,
  and activity labels.
- Added responsive layouts for desktop, tablet, mobile, and narrow mobile
  widths without horizontal overflow.
- Added Vitest and unit coverage for formatting, overview aggregation, and
  chart geometry.
- Updated documentation conventions so epic, changelog, and ADR numbering are
  independent and relationships are expressed through frontmatter IDs.

## Decisions

- Build the initial product experience as a focused single-page application
  dashboard.
- Calculate the fixed 28-day overview in the frontend from canonical
  activities returned by `GET /api/activities`.
- Render the daily training-volume chart as a custom responsive SVG.
- Use `d3-shape` with `curveMonotoneX` to generate the line and area paths
  without adopting a full chart component library.
- Keep the displayed curve tied to actual daily moving-time values rather than
  presenting it as a statistical trend.
- Keep overview aggregation in the frontend for this fixed period and current
  activity volume; no aggregate backend endpoint is needed yet.
- Show the complete activity list from already loaded API data rather than
  issuing another request when the user expands recent activities.
- Preserve existing dashboard data during sync attempts and sync failures.
- Render visible chart date labels as normal HTML so their text remains
  readable when the SVG is resized on mobile.

## Verification

- `npm test` passed with 21 unit tests across three test files.
- `npm run lint` passed.
- `npm run build` passed with TypeScript checking and a Vite production build.
- `go test ./...` passed.
- `git diff --check` passed.
- Manually verified the application with real locally imported activities at
  1280×800, 768×1024, 700×900, 390×844, and 320×800 viewport sizes.
- Verified that the activity-list expansion changes `aria-expanded`, controls
  the expected list, and renders the complete loaded activity set.
- Verified no horizontal overflow, duplicate IDs, browser console warnings, or
  browser console errors in the tested layouts.
- Verified text and button color combinations meet WCAG AA contrast.

## Follow-ups

- Consider adding a rolling seven-day trend line as a separate analytical
  layer after the initial daily chart exists. Keep it visually and
  semantically distinct from the actual daily values.
- Add focused React component tests with Testing Library when component
  behavior grows beyond the current unit-test coverage. Useful first cases are
  metric label/value rendering, accessible chart naming and SVG output, and
  connected, empty, loading, syncing, and error application states.
- Add interactive chart inspection so a user can hover over the training
  curve and see the exact date and daily moving-time value at that position.
  Provide the same information through keyboard focus and an appropriate
  touch interaction rather than making the feature hover-only.
