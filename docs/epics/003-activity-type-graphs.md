---
id: epic-003
number: 3
slug: activity-type-graphs
related_adrs: []
related_changelogs:
  - changelog-005
---

# Epic 003: Activity-Type Graphs

## Goal

As a user, I can see how my running, cycling, and workout training developed
over the last four weeks in separate graphs using a metric that fits each
activity type.

## Why

The training overview established a provider-neutral daily moving-time curve.
That combined value is useful for overall volume, but it hides the different
shape and meaning of common activity types. Distance is the most direct daily
volume measure for running and cycling, while moving time remains the useful
measure for workouts without a meaningful distance.

## Scope

### In Scope

- Use canonical activities already returned by `GET /api/activities`
- Show separate 28-day graphs for canonical `run`, `ride`, and `workout`
- Use daily distance in kilometers for running and cycling
- Use daily moving time for workouts
- Include zero-value days in every graph
- Sum multiple matching activities on the same local calendar day
- Show visible y-axis values with the metric unit
- Keep the graphs understandable with text and accessible daily values
- Refresh all graphs after a successful Strava sync
- Preserve the existing cross-sport daily moving-time graph
- Preserve the existing summary metrics and recent activity list
- Keep the layout usable on desktop and mobile

### Out of Scope

- Additional activity-type graphs
- Combining or reclassifying unknown canonical activity types
- Pace, speed, power, heart rate, training load, or intensity analysis
- Interactive date ranges, tooltips, zooming, or chart comparison controls
- New backend aggregate endpoints

## Time and Aggregation Rules

- The graphs use the same today-plus-preceding-27-days period as the existing
  overview.
- Activity inclusion is based on `started_at` in the user's local calendar.
- Only exact canonical types `run`, `ride`, and `workout` contribute to their
  respective graph.
- Running and cycling values sum `distance_meters` per day.
- Workout values sum `moving_seconds` per day.
- A graph remains visible with a zero baseline when no matching activity exists.

## Acceptance Criteria

1. The overview retains the cross-sport daily moving-time graph and adds
   separate running, cycling, and workout graphs.
2. Running and cycling graphs plot daily distance and label the y-axis in km.
3. The workout graph plots daily moving time and labels the y-axis as time.
4. Each graph contains 28 daily values, including zero-value days.
5. Multiple matching activities on one day are summed.
6. Other activity types do not contribute to these three graphs.
7. The graphs update after sync without a page reload.
8. The graph period matches the summary period.
9. The graphs include accessible names, descriptions, and daily value text.
10. Frontend tests, linting, build, and backend tests pass.

## Technical Notes

- Extend the existing frontend overview aggregation instead of introducing a
  backend endpoint.
- Keep chart geometry metric-agnostic and apply distance or duration formatting
  at the component boundary.
- Reuse the canonical `run` and `ride` normalization already performed during
  provider transformation.
