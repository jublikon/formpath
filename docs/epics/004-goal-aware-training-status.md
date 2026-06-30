---
id: epic-004
number: 4
slug: goal-aware-training-status
related_adrs: []
related_changelogs:
  - changelog-008
---

# Epic 004: Goal-Aware Training Status

## Goal

As a user, I can define a dated running or cycling distance goal and see the
recent training context that is relevant to that goal.

## Why

The existing training overview shows what happened, but it does not know what
the user is preparing for. A concrete event goal gives Formpath the first
durable piece of user intent and makes the same activity history more useful.

This epic should help a user answer:

1. What am I training for, and how much time remains?
2. What relevant training have I completed recently?
3. How does that compare with the preceding period?

The result is a factual training status, not a readiness prediction or
coaching recommendation. Formpath does not yet have enough context to claim
that a user is on track for an event.

## Product Shape

The existing training overview gains a focused goal area.

A user without a goal can create one. A user with an active goal can see:

- the event name, sport, target distance, and target date
- an optional target time
- the number of calendar days remaining
- recent sport-specific activity count, distance, moving time, and longest
  activity
- comparison of recent distance and activity count with the preceding period
- the period and data basis behind the displayed status

The goal can be edited or removed. The rest of the existing overview remains
available.

## Scope

### In Scope

- Support one active distance-event goal for the local user
- Support canonical `run` and `ride` goal sports
- Store:
  - an event name
  - goal sport
  - target distance
  - target calendar date
  - an optional target duration
- Create, load, edit, and remove the active goal
- Calculate goal status only from locally stored canonical activities
- Include only activities whose canonical `activity_type` matches the goal
  sport
- Show a recent 28-day period and compare it with the immediately preceding
  28-day period
- Show recent activity count, distance, moving time, and longest activity
- Show absolute changes between the recent and preceding periods
- Show calendar days remaining until the target date
- Refresh the goal status after a successful activity sync
- Keep calculations provider-neutral
- Provide understandable empty, incomplete, loading, validation, and failure
  states
- Preserve the existing training overview, activity-type graphs, and recent
  activity list
- Provide unit coverage for goal validation and status calculations

### Out of Scope

- Capability goals such as being able to run five kilometers comfortably
- Multiple simultaneous goals or historical completed-goal management
- Goals without a target date
- Goal types other than running and cycling distance events
- Training-plan generation or scheduled workouts
- Prescriptive next-workout recommendations
- Claims that the user is ready, on track, behind, or likely to achieve the
  goal
- Completion percentages derived by dividing recent training distance by event
  distance
- Percentage changes between the recent and preceding training periods
- Recovery, readiness, fatigue, or injury-risk calculations
- Apple Health, Polar H10, or additional provider integrations
- Subjective check-ins
- AI-generated coaching or analysis
- Notifications and reminders
- Multi-user authentication or public deployment
- Introducing dbt, Athena, Iceberg, or another analytical platform solely for
  this slice

## Goal Rules

- The goal type is `distance_event`.
- The goal sport is either canonical `run` or canonical `ride`.
- Target distance must be greater than zero.
- Target date is stored as a calendar date rather than a timestamp.
- The browser rejects a newly created target date before its current local
  calendar date.
- The backend validates the target date as a real calendar date but does not
  compare it with the server clock in this slice.
- Target duration is optional and must be greater than zero when present.
- Only one goal can be active for the local user. Creating another active goal
  requires replacing or removing the existing goal.
- Removing a goal does not remove activities or imported provider data.
- The optional target duration is displayed as goal context but is not used to
  infer readiness in this epic.

## Time and Aggregation Rules

- The recent period covers the user's current local calendar day and the
  preceding 27 local calendar days.
- The comparison period is the 28 local calendar days immediately before the
  recent period.
- Activity inclusion is based on `started_at`.
- Only activities matching the goal's canonical sport are included.
- Distance and moving time are summed within each period.
- Activity count counts every included canonical activity.
- Longest activity means the included activity with the greatest
  `distance_meters`.
- Only absolute differences are shown for distance and activity count.
- Missing optional activity values do not exclude an otherwise valid activity.
- Invalid activity timestamps are ignored rather than breaking the goal
  status.
- Days remaining is calculated in the browser from local calendar dates. It is
  not calculated from elapsed 24-hour intervals.
- Formatting uses the user's browser locale, following the existing overview.

## Status Interpretation

The UI may state facts such as:

- "You ran 42 km in the last 28 days."
- "That is 8 km more than in the preceding 28 days."
- "Your longest recent run was 15 km."
- "There are 96 days until your event."

The UI must not turn those facts into unsupported conclusions such as:

- "You are 60% marathon-ready."
- "You are on track."
- "You should run 18 km next."

This boundary keeps the status honest while later recovery, planning, and
recommendation capabilities are still absent.

## Acceptance Criteria

1. A user without an active goal can create a running or cycling distance-event
   goal with a name, distance, date, and optional target time.
2. Invalid distances, malformed calendar dates, and invalid target times are
   rejected with understandable validation messages. The UI also rejects a
   target date before the browser's current local date.
3. Reloading the application preserves and displays the active goal.
4. The user can edit or remove the active goal.
5. The goal status uses only locally stored canonical activities matching the
   selected goal sport.
6. The status shows activity count, total distance, moving time, and longest
   activity for the recent 28-day period.
7. The status shows absolute differences in recent distance and activity count
   compared with the immediately preceding 28-day period.
8. The status shows the target date and calendar days remaining.
9. The status updates after a successful activity sync without a page reload.
10. Missing matching activities produce a useful empty status without
    affecting the rest of the training overview.
11. The interface does not claim readiness, predict goal completion, or
    generate a training plan.
12. Goal and status calculations remain provider-neutral and do not expose
    Strava-specific fields.
13. Frontend tests, backend tests, linting, and the production frontend build
    pass.

## Technical Notes

- Persist the goal as application state in Postgres. It is not provider data
  and does not belong in raw object storage.
- Keep the first goal model explicit rather than introducing a generic rules
  engine or arbitrary goal-definition language.
- Use a date-capable Postgres type for the target calendar date.
- Preserve numeric values in base units at storage and API boundaries:
  meters and seconds.
- Reuse the existing canonical activity model as the status input.
- Keep status calculation separate from React rendering so its rules are
  directly testable.
- Keep browser-local date comparisons in the frontend for this slice. The goal
  API accepts and returns the calendar date without converting it through a
  server timezone.
- Do not introduce a data warehouse or analytical job runner for this data
  volume. Revisit a shared server-side metrics layer when the status must be
  consumed by agents or additional clients.
- The current browser-local period semantics are acceptable while Formpath has
  no user profile or configured timezone. A durable user-timezone decision is
  deferred until those concepts enter the product model.

## Milestones

1. **Goal Model and Persistence** — store and validate one active event goal
2. **Goal API and Form** — create, load, edit, and remove the goal
3. **Goal Status Model** — calculate recent and comparison-period metrics with
   tests
4. **Goal-Aware Overview** — present goal context and factual training status
5. **Application States** — handle missing goals, no matching activities,
   validation failures, loading, and API failures
6. **Verification** — run affected frontend and backend suites and verify the
   responsive UI

## Follow-ups

- Add activity details and deeper workout analysis.
- Introduce Apple Health sleep, HRV, and resting-heart-rate data as the first
  recovery inputs.
- Consider subjective check-ins and personal baselines as later recovery
  inputs.
- Define an explainable recovery state before introducing adaptive
  recommendations.
- Revisit a reusable metrics API when goal status is consumed outside the
  current React application.
- Keep capability goals as a future vision scenario unless they are explicitly
  promoted into the roadmap.
