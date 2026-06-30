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
- Additional recovery and training-data imports from providers such as Garmin,
  Polar, Oura, WHOOP, and Amazfit
- Subjective check-ins and personal baselines as a possible later recovery
  input
- An explainable recovery state based on personal baselines and training load
- Guided Polar H10 recovery measurements and later live workout streams
- An iOS companion app prototype for HealthKit import, check-ins, and local
  mobile workflows
- Adaptive training recommendations and plans grounded in goals and recovery
- A daily training decision API that returns a typed decision, workout context,
  reasons, confidence, risk flags, and any plan change
- Authentication and a privacy-safe deployment
- Additional health, training, and nutrition data providers

## Future Planning Notes

### Decision Engine Direction

The long-term product should sharpen toward an open, self-hostable,
device-agnostic adaptive training decision engine for endurance athletes.

This is different from building only a training dashboard, only a Strava
alternative, only a recovery wearable clone, or only an AI chatbot over fitness
data. The future decision flow should be:

1. Import provider data into canonical athlete models.
2. Compute training, recovery, goal, and subjective-context metrics.
3. Apply testable decision and plan-adaptation rules.
4. Return a typed decision object with reasons, confidence, risk flags, and
   plan changes.
5. Use language models later for explanation, summarization, feedback handling,
   and interaction rather than as the sole decision-maker.

### Device and Sensor Strategy

Formpath should not plan around building its own 24/7 raw-data wearable in the
near term. The stronger path is to import and normalize data from existing
activity, health, wearable, and provider ecosystems, then create value through
open recovery logic, training adaptation, and explainable decisions.

Consumer devices should be treated according to the access they realistically
provide:

- Apple Health and HealthKit are likely first-class paths for sleep, HRV,
  resting heart rate, and iOS-local health data.
- Garmin, Polar Flow, Oura, WHOOP, Amazfit, and similar providers are future
  candidates for synchronized recovery and activity data where APIs or exports
  make this practical.
- Polar H10 and Polar Verity Sense are useful for live workout streaming,
  guided recovery measurements, and experiments, but not as the main 24/7
  recovery-data strategy.
- Consumer watches and rings should generally be expected to provide processed
  or aggregated health data rather than unrestricted continuous raw streams.
- Research hardware with deeper raw access is a later investigation because it
  is expensive and less aligned with a consumer endurance-coaching product.

### iOS Companion Constraints

An early iOS companion app can be explored before public distribution. A normal
Apple ID is enough for local development on a personal device through Xcode,
but these development builds have short-lived signing and are not a substitute
for distribution. Broader beta distribution through TestFlight and App Store
distribution belong to an Apple Developer Program milestone.

The iOS companion should prioritize HealthKit-backed imports, local check-ins,
and product workflows that fit the platform. CoreBluetooth background behavior
may support constrained sensor experiments, but Formpath should not assume that
iOS will allow arbitrary guaranteed 24/7 Bluetooth streaming.
