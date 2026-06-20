---
id: changelog-004
number: 4
slug: raw-first-elt-ingestion
related_epics:
  - epic-001
related_adrs:
  - ADR-003
  - ADR-004
---

# Changelog 004: Raw-First ELT Ingestion

## Summary

Changed Strava activity synchronization from a partially coupled ingestion
flow to an explicit raw-first ELT pipeline. Successful provider responses are
now stored unchanged and read back from raw storage before any JSON decoding,
validation, or canonical normalization occurs.

## Related Epics

- [Epic 001: Fetch and Store Strava Activities Locally](../epics/001-strava-activity-ingestion.md)

## Related ADRs

- [ADR-003: Local-First Postgres and MinIO Storage](../adr/003-postgres-local-first-token-storage.md)
- [ADR-004: Raw-First ELT for Provider Ingestion](../adr/004-raw-first-elt-ingestion.md)

## Relevant Changes

- Split Strava activity extraction from transformation. The provider client now
  returns the successful response as opaque bytes without decoding it.
- Extended the raw object store with object retrieval and implemented retrieval
  through MinIO.
- Made the persisted raw object the input to Strava JSON decoding, validation,
  activity-type normalization, and canonical activity mapping.
- Preserve successful but invalid JSON responses in raw storage before
  returning a transformation error.
- Stop canonical persistence when the newly written raw object cannot be read
  back.
- Require complete S3 configuration at backend startup whenever Postgres
  persistence is configured.
- Preserve a newly uploaded raw object when its metadata cannot be recorded in
  Postgres, prioritizing source-data retention over immediate catalog
  consistency.
- Verify loaded raw bytes against their catalogued SHA-256 checksum before
  transformation.
- Added tests proving that transformation uses the stored payload rather than
  the original in-memory HTTP response.
- Extended the opt-in MinIO integration test to verify that stored raw bytes
  can be retrieved unchanged.

## Decisions

- Use a synchronous ELT sequence within the existing
  `POST /api/activities/sync` request so the external API and current product
  behavior remain unchanged.
- Treat raw provider objects as the durable source of truth and canonical
  activities as derived records.
- Do not delete successfully extracted raw data to compensate for a Postgres
  metadata failure. Such objects remain uncatalogued until a later
  reconciliation workflow.
- Keep provider-specific simplification, including cycling variants mapped to
  `ride`, in the transform step rather than the extract or load steps.
- Do not introduce background jobs, transformation versions, or a bulk
  reprocessing endpoint in this change.

## Verification

- `go test ./cmd/server` passed.
- `go test ./...` passed.
- The opt-in MinIO/Postgres raw object integration test passed against the
  local Compose stack, including unchanged object retrieval.
- The integration suite covers raw-object preservation after metadata failure
  and checksum mismatch rejection.
- `git diff --check` passed.

## Follow-ups

- Add a reprocessing workflow that rebuilds canonical records from selected raw
  objects without contacting the provider.
- Add transformation version metadata when multiple mapping versions need to
  coexist or be audited.
- Introduce a `pending`/`ready` ingestion-state model and reconciliation for
  uploaded raw objects whose metadata was not finalized.
