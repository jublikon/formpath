# Formpath Roadmap

## Current

### Epic 004: Goal-Aware Training Status

Let a user define one dated running or cycling distance goal and understand
their recent, sport-specific training context relative to that goal without
claiming readiness or generating a training plan.

## Completed

### Epic 003: Activity-Type Graphs

Show separate four-week graphs for running distance, cycling distance, and
workout moving time using the existing canonical activity data.

### Epic 002: Training Overview

Turn imported canonical activities into the first useful Formpath product
experience: a single-page dashboard with four-week summary metrics, a daily
training-volume curve, and a human-readable recent activity list.

### Epic 001: Strava Activity Ingestion

Connect Strava, import activities through a raw-first ELT pipeline, preserve provider data as the source of truth, derive the canonical activity model, store it locally, and expose it through the API and initial React UI.

## Next Candidates

- Activity details and deeper workout analysis
- Apple Health recovery signals, beginning with sleep, HRV, and resting heart
  rate
- Subjective check-ins and personal baselines as a possible later recovery
  input
- An explainable recovery state based on personal baselines and training load
- Guided Polar H10 recovery measurements and later live workout streams
- Adaptive training recommendations and plans grounded in goals and recovery
- Authentication and a privacy-safe deployment
- Additional health, training, and nutrition data providers
