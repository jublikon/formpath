# ADR-001: Go as Backend Language

## Status

Accepted

## Context

Formpath needs a backend that collects data from external APIs, normalizes it, stores it, and exposes it through its own interfaces. The developer (solo) wants to learn Go while also building a system that can be used productively.

Alternatives would be Python (familiar, but less suitable for long-running services) or TypeScript (a good ecosystem fit, but not a learning goal).

## Decision

We use Go as the language for the Formpath backend.

## Consequences

- Go enforces explicit error handling and simple structures — this fits the engineering principle of "simple, explicit solutions"
- Strong standard library for HTTP servers, JSON, OAuth2 — few external dependencies are needed
- Compiles to a single binary — simple local setup and later deployment
- The learning curve is real, but addressed through real features instead of exercises
- No generics-heavy design at the beginning — prefer idiomatic Go
