---
id: changelog-005
number: 5
slug: activity-type-graphs
related_epics:
  - epic-003
related_adrs: []
---

# Changelog 005: Activity-Type Graphs

## Summary

Extended the training overview with separate four-week graphs for running
distance, cycling distance, and workout moving time while retaining the
existing cross-sport daily moving-time graph.

## Related Epics

- [Epic 003: Activity-Type Graphs](../epics/003-activity-type-graphs.md)

## Related ADRs

- No architectural decision record applies to this frontend slice.

## Relevant Changes

- Added daily running distance, cycling distance, and workout moving-time
  series to the existing 28-day frontend aggregation.
- Added three responsive graphs with metric-specific y-axis labels.
- Retained the existing cross-sport daily moving-time graph.
- Kept zero-value days and zero-only graphs visible.
- Added accessible graph descriptions and text values for every day.
- Preserved the existing overview totals, synchronization behavior, and recent
  activity list.
- Added unit coverage for activity-specific aggregation and kilometer
  formatting.

## Decisions

- Use distance for canonical `run` and `ride` activities and moving time for
  canonical `workout` activities.
- Keep aggregation in the frontend because the graphs use the same fixed period
  and already-loaded canonical activity data as the overview.
- Do not infer graph membership for other activity types.
- Keep all three graphs visible even when one has no matching activities.
- Keep the cross-sport graph as the overall training-volume view and use the
  activity-specific graphs as additional detail.

## Verification

- `npm test` passed with 23 unit tests across three test files.
- `npm run lint` passed.
- `npm run build` passed with TypeScript checking and a Vite production build.
- `go test ./...` passed.
- Visually verified the graph layout at the default desktop viewport and at
  390×844 with synthetic, non-personal activity data that was not retained in
  the product code.
- Verified visible kilometer and duration y-axis labels, responsive stacking,
  no horizontal overflow, and no browser console warnings or errors.

## Follow-ups

- Add keyboard, pointer, and touch inspection for exact visible chart points.
- Consider additional activity types only when their primary volume metric is
  explicitly defined.
