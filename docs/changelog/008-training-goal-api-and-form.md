---
id: changelog-008
number: 8
slug: training-goal-api-and-form
related_epics:
  - epic-004
related_adrs: []
---

# Changelog 008: Training Goal API and Form

## Summary

Added the first goal-aware product surface: a local user can load, create,
edit, and remove one active running or cycling distance-event goal. The change
keeps the experience factual and does not introduce readiness, strain,
recommendation, or training-plan claims.

## Related Epics

- [Epic 004: Goal-Aware Training Status](../epics/004-goal-aware-training-status.md)

## Related ADRs

- No architectural decision record currently applies to this slice.

## Relevant Changes

- Added `GET /api/training-goal`, `PUT /api/training-goal`, and
  `DELETE /api/training-goal` route registration.
- Added training-goal HTTP handlers for loading, saving, and removing the
  local user's active goal.
- Kept server-owned goal fields under backend control: the handler sets the
  configured app user ID and `distance_event` goal type.
- Added handler coverage for stored goals, missing goals, invalid JSON,
  validation failures, store failures, and idempotent deletion.
- Added frontend training-goal API types and form helpers for calendar-day
  calculations, base-unit conversion, target-duration parsing, and browser-local
  validation.
- Added a focused training-goal panel to the existing dashboard.
- Supported create, edit, cancel, save, and remove flows without adding routes,
  navigation, or a separate goal page.
- Displayed factual active-goal context: sport, distance, target date, calendar
  days remaining, and optional target time.
- Kept target-date validation in the browser for this slice, rejecting dates
  before the current local calendar date.
- Separated missing-goal empty state from goal-load failures so a temporary API
  failure does not open a create-or-replace form.
- Added a retry action for goal-load failures.
- Added confirmation before removing the active goal.
- Bounded optional target duration to the Postgres integer range before save.
- Preserved the existing training overview, activity graphs, Strava integration
  panel, and recent activity list.

## Decisions

- Keep the first goal UI inline on the existing dashboard because goals are not
  yet a separate product area.
- Use `PUT /api/training-goal` for create-or-replace behavior, matching the
  one-active-goal constraint in Epic 004.
- Treat a missing goal as a normal empty state in the frontend rather than a
  page-level loading failure.
- Treat a goal-load failure as an unavailable state with retry, not as an empty
  goal state, because `PUT /api/training-goal` replaces the one active goal.
- Confirm deletion before removing the active goal because the first goal UI has
  no undo or goal history.
- Keep goal status calculations out of this slice. Recent-vs-previous-period
  goal status is the next milestone after the create/edit/remove flow.
- Keep CSS in the existing app stylesheet for this change. Splitting styles
  should happen when the frontend structure itself is ready for a small
  organization pass.

## Verification

- `go test ./cmd/server` passed.
- `go test ./...` passed.
- `source ~/.nvm/nvm.sh && nvm use && npm test` passed.
- `source ~/.nvm/nvm.sh && nvm use && npm run lint` passed.
- `source ~/.nvm/nvm.sh && nvm use && npm run build` passed.
- `git diff --check` passed.

## Follow-ups

- Build the goal status model for recent 28-day sport-specific training context
  and comparison with the immediately preceding 28-day period.
- Present factual goal status in the overview without readiness, strain,
  plan-generation, or recommendation claims.
- After goal status is added, consider moving API loading and mutation state
  out of `App.tsx` into small API/state hooks. `App.tsx` is becoming the
  orchestrator for activities, Strava integration, and goals; that remains
  acceptable for this PR but should not keep growing indefinitely.
- After the goal UI and status UI settle, consider splitting `App.css` into
  smaller style files by concern or component area. A single stylesheet remains
  acceptable for this PR, but goal, chart, and activity-list styles will become
  harder to scan as the UI grows.
