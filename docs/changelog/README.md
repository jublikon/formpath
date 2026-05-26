# Changelog

This folder contains one Markdown file per merged change.

The goal is not to mirror every commit. Each entry should capture the decisions, relevant behavior changes, trade-offs, and follow-up work that matter when reading the project history later.

Entries should be self-contained. Do not rely on private chat history or memory outside the repository. If a decision came from a conversation, summarize the decision and the reasoning in the changelog entry so future readers can understand it from the repo alone.

## File Naming

Use this format:

```text
NNN-short-change-name.md
```

Example:

```text
001-strava-activity-ingestion.md
```

The number and slug should match the primary epic when the changelog records work for a single epic:

```text
docs/epics/001-strava-activity-ingestion.md
docs/changelog/001-strava-activity-ingestion.md
```

The date belongs in the changelog content only when it is relevant to the change history. It is not part of the filename.

## Entry Structure

```markdown
---
id: changelog-001
number: 1
slug: strava-activity-ingestion
related_epics:
  - epic-001
related_adrs:
  - ADR-001
  - ADR-002
---

# Changelog 001: Change Title

## Summary

Short description of what changed and why.

## Related Epics

- Links to the epic or epics this changelog belongs to

## Related ADRs

- ADRs that materially informed this change

## Relevant Changes

- User-visible or system-visible changes
- Important test coverage or behavior changes
- Storage/API/configuration changes

## Decisions

- Decisions made during the change
- Trade-offs and intentionally deferred work

## Verification

- Tests, manual checks, or commands run

## Follow-ups

- Open work discovered during the change
```
