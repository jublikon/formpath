# Formpath Vision

## Vision

Formpath exists to help ambitious recreational athletes turn fragmented health and training data into better decisions, better habits, and better progress toward real-world sports goals.

Our vision is to build an open, cross-platform, agentic health and performance platform that helps people prepare for goals such as a marathon on a specific date, while fitting training into the realities of everyday life.

Instead of forcing users into a static plan or a closed ecosystem, Formpath brings together data from multiple sources, analyzes it in one place, and enables adaptive applications on top — including agentic experiences that can explain progress, surface insights, and dynamically generate or adjust training plans based on the user’s current state.

---

## The Problem

Today, ambitious hobby athletes often have data spread across multiple apps, devices, and ecosystems.

- Apple Health stores a lot of data, but offers limited analysis and guidance.
- Closed products such as Whoop can be powerful, but are expensive and tied to one ecosystem.
- Combinations like Strava Premium and Runna Premium can also become expensive, while still leaving data fragmented across tools.
- Many products either focus on raw tracking, static dashboards, or rigid plans, rather than adaptive support for real goals in real life.

As a result, users often have the data, but not a unified system that helps them act on it.

---

## Who Formpath Is For

Formpath is built for ambitious recreational athletes.

These users care about performance, health, and consistency. They may train for concrete goals such as a marathon, a cycling event, or general endurance improvement, but they also need to balance sport with work, recovery, sleep, stress, and daily life.

They do not just want more data.
They want better decisions.

---

## What Formpath Will Enable

Formpath will provide a flexible platform that collects, unifies, and analyzes health and sport data from different sources.

On top of that platform, applications can be built that help users:

- track progress toward concrete athletic goals
- understand trends across workouts, recovery, sleep, and wellbeing
- adapt training plans on demand
- interact with their data in a more intelligent way
- generate tailored outputs such as plans, analyses, and visualizations

Examples include:

- tracking maximum heart rate across workouts over time
- measuring recovery speed after hard sessions
- analyzing sleep quality and sleep changes after workouts
- tracking focus, wellbeing, and subjective readiness
- measuring progression in athletic performance over time
- reducing the load of the next workout after poor sleep or insufficient recovery
- adjusting plans dynamically instead of following a rigid static schedule

### Future Scenario: Adaptive Training Decision Engine

A long-term product direction is for Formpath to become an open-source,
device-agnostic adaptive training decision engine for endurance athletes.

The important distinction is engine-first rather than dashboard-first or
LLM-first. Formpath should not merely visualize imported data or ask a language
model to improvise coaching advice. The durable core should be a testable
decision system that combines:

- canonical activity history
- training load and recent workout context
- concrete goals and available training time
- recovery signals such as sleep, HRV, and resting heart rate
- subjective check-ins such as fatigue, soreness, stress, and motivation

The eventual daily question is: "What should I do today?" Example outputs could
include an easy endurance session, a hard workout, rest, mobility, a plan
change, confidence, risk flags, and a clear explanation of the data basis behind
the decision.

This future direction should remain explainable by design. Deterministic rules,
metrics, baselines, and validation should own the decision object. A language
model can later help summarize, explain, translate, or collect user feedback,
but it should not be the only place where training decisions are made.

### Future Scenario: Device and Data Strategy

Formpath should prefer a device-agnostic import strategy over trying to become a
24/7 raw-sensor wearable. The platform can create more leverage by normalizing
activity, health, recovery, and subjective data from sources such as Strava,
Apple Health, Garmin, Polar, Oura, WHOOP, Amazfit, and future providers.

For 24/7 recovery signals, the expected path is imported or synchronized health
data such as sleep, HRV, resting heart rate, and daily activity summaries rather
than a guaranteed continuous raw stream from consumer wearables. Consumer
watches, rings, bands, and vendor clouds should be treated primarily as data
providers whose access patterns and raw-data availability vary by platform.

An iOS companion app is a plausible future product surface for low-friction
HealthKit-backed imports, local check-ins, notifications, and lightweight
capture workflows. The intended experience is that a user grants permissions
once and Formpath keeps relevant local health and training context fresh in the
background where the platform allows it, without requiring frequent manual
sync actions.

That background-sync goal is separate from unrestricted background Bluetooth
streaming. Background BLE experiments may be valuable for session-based sensors,
but the product plan should not depend on arbitrary always-on 24/7 raw BLE
streams from consumer devices.

Polar H10 or Polar Verity Sense integrations are useful candidates for guided
live workout or recovery-measurement experiments. They should be treated as
session-based research inputs, not as the main path to a consumer-grade 24/7
recovery system. Dedicated research hardware may offer deeper raw access, but
its cost and audience make it a later investigation rather than the core
consumer direction.

### Future Scenario: Capability Goals

As a possible future product direction, Formpath could support ongoing
capability goals in addition to event goals with a fixed date. An example is
building and maintaining enough running fitness to complete five kilometers
comfortably, rather than preparing for a particular race.

Such a goal would need an explicit, measurable definition of terms such as
"comfortably", potentially combining completed distance with subjective effort
and continuity. This is a future scenario, not a current roadmap commitment or
planned development slice.

---

## Product Direction

Formpath is not just an app.
It is a platform plus an application layer.

The platform provides a central and consistent foundation for collecting and evaluating health and sport data.

On top of that, Formpath can support agentic applications that let users:

- explore and interact with their data
- ask questions about their progress
- receive adaptive guidance
- generate training plans and custom visualizations
- use intelligent workflows tailored to their goals and current condition

This creates a product experience that is interactive, adaptive, and personalized rather than static.

---

## Differentiation

Formpath is differentiated by being:

- **cross-platform** rather than tied to a single manufacturer
- **open** rather than locked into one device ecosystem
- **data-unified** rather than fragmented across apps
- **decision-engine-first** rather than only a dashboard or generic chatbot
- **explainable** rather than opaque about why a recommendation changed
- **interactive** rather than passive
- **adaptive** rather than static
- **agentic** rather than limited to fixed dashboards or predefined flows

The goal is not just to show data, but to make data useful.

---

## What Formpath Is Not

Formpath is not:

- a pure dashboarding app
- a monolithic closed ecosystem
- a rigid PDF training plan
- just another generic fitness plan app
- a direct attempt to clone a proprietary wearable
- a product that depends on unrestricted 24/7 raw sensor streaming
- a product that only stores data without helping users act on it

---

## What Success Looks Like

Formpath is successful when users can:

- import data flexibly from different sources
- view and understand their health and training data in one place
- track progress toward health and sports goals easily
- adapt plans and decisions based on real-world changes such as poor sleep, fatigue, or recovery patterns
- use one coherent system instead of stitching together multiple expensive tools

Over time, success means Formpath becomes the trusted foundation on which adaptive sports and health applications can be built.

---

## Long-Term Vision

In the long term, Formpath should become the open foundation for adaptive performance and health applications.

It should make it possible to build software that does not just record what happened, but helps users decide what to do next.

The long-term ambition is to create a platform where data collection, analysis, interaction, and action come together in one system — enabling a new generation of agentic applications for ambitious everyday athletes.
