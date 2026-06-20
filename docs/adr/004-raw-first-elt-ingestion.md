# ADR-004: Raw-First ELT for Provider Ingestion

## Status

Accepted

## Context

Formpath imports provider data into a provider-neutral canonical model. Provider
responses contain more detail than the canonical model needs immediately. For
example, Strava distinguishes cycling variants such as `GravelRide`,
`MountainBikeRide`, and `VirtualRide`, while the current canonical activity
model groups them as `ride`.

Transforming a provider response before it has been durably stored couples data
acquisition to the current canonical mapping. A decoding error, a newly added
provider field, or a future mapping change should not require fetching the
provider data again or cause the original response to be lost.

## Decision

Provider ingestion follows a raw-first ELT sequence:

1. **Extract** the successful provider response as opaque bytes.
2. **Load** those bytes unchanged into the raw object store and record their
   metadata.
3. **Transform** only the persisted raw object into provider DTOs and canonical
   records.

The persisted raw object is the source of truth for canonical transformation.
The synchronous Strava sync therefore reads the object back from raw storage
before decoding, validating, normalizing, and writing activities to Postgres.

Provider-specific normalization remains part of the transform step. Canonical
records may intentionally reduce provider detail, but the complete provider
payload remains available in raw storage and is referenced through
`raw_object_key`.

Only successful provider HTTP responses enter the raw data layer. Non-success
responses remain request failures and are not treated as provider data.

## Consequences

- A successful raw load is required before canonical activities can be
  transformed or updated.
- Invalid or currently unsupported provider payloads remain stored even when
  transformation fails.
- Canonical tables contain derived data and can be rebuilt from raw objects
  when a reprocessing workflow is introduced.
- The raw object store must support both writing and reading objects.
- The current sync remains synchronous: extract, load, transform, and canonical
  persistence still complete within one request.
- Background transformation, transformation versioning, and bulk reprocessing
  are deferred until product or operational needs justify them.
