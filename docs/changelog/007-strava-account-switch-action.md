---
id: changelog-007
number: 7
slug: strava-account-switch-action
related_epics:
  - epic-001
related_adrs: []
---

# Changelog 007: Strava Account Switch Action

## Summary

Kept the Strava OAuth action available after a user is already connected, so
they can start the authorization flow again and connect a different Strava
account without losing access to the manual activity sync action.

## Related Epics

- [Epic 001: Fetch and Store Strava Activities Locally](../epics/001-strava-activity-ingestion.md)

## Related ADRs

- No architectural decision record applies to this frontend interaction.

## Relevant Changes

- Changed the connected Strava integration panel from a single sync button to
  separate `Change Account` and `Sync activities` actions.
- Kept `/auth/strava` as the OAuth entry point for both initial connection and
  account changes.
- Added responsive styling for the two-action integration control.
- Added component coverage for connected, syncing, and disconnected Strava
  panel states.

## Decisions

- Use `Change Account` instead of `Re-authenticate` because the action
  communicates the user-visible outcome, works across future providers, and
  avoids exposing the technical OAuth step.
- Keep activity synchronization as a separate action so reconnecting an
  account and importing activities remain explicit user choices.

## Verification

- `npm test` passed with 35 tests across five test files.
- `npm run lint` passed.
- `npm run build` passed with TypeScript checking and a Vite production build.

## Follow-ups

- Decide whether a future multi-provider integration surface should use a more
  general label such as `Change data source` once providers beyond Strava are
  available.
- Define explicit disconnect or logout behavior for provider integrations,
  including whether it only removes stored tokens or also offers local data
  deletion.
- Decide how account changes should handle already imported local activities
  and raw provider objects when the connected provider user changes.
- Consider storing or scoping canonical activities by provider account identity
  so data from two different accounts cannot be mixed unintentionally.
