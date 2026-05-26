# ADR-002: Strava as the First Provider with OAuth2 Authorization Code Flow

## Status

Accepted

## Context

Formpath should collect data from multiple sources. We need a first provider to build the entire ingestion path end to end: OAuth integration, API fetch, normalization, and storage.

Strava is the obvious choice:
- Well-documented REST API
- Standard OAuth2 Authorization Code Grant
- Provides structured activity data (running, cycling, etc.)
- `localhost` is allowed as a callback domain for local development
- Rate limits are clearly defined (200/15min, 2000/day)

## Decision

Strava is the first provider connector. We implement the full OAuth2 Authorization Code Flow:

1. The user is redirected to Strava (`/oauth/authorize`)
2. Strava redirects back with an authorization code
3. The backend exchanges the code for an access token + refresh token
4. The backend stores tokens and refreshes them automatically

Initial requested scopes: `activity:read_all` (all activities including private ones).

During the local-first phase, OAuth `state` is stored in a short-lived `HttpOnly` cookie and compared in the callback with the `state` returned by Strava. This avoids a server-side session store as long as user/session persistence does not exist yet.

For production or multi-instance deployment, `state` is moved to a shared server-side store, such as Redis/ElastiCache or Postgres, with TTL and one-time use.

## Consequences

- The OAuth2 flow is the most complex part of the first slice — but it is reusable for other providers
- Rate limits (200/15min) require deliberate pacing for historical imports
- Token rotation (access token every 6h, refresh token rotates on every refresh) must run reliably in the backend
- Local-first `state` handling via cookie is simple and stateless, but must be moved to a shared store with TTL and one-time use before production
- The provider adapter layer must be cleanly separated from the domain model from the beginning
- Strava-specific fields must not leak into the canonical model
