# Changelog

This folder contains one Markdown file per merged change.

The goal is not to mirror every commit. Each entry should capture the decisions, relevant behavior changes, trade-offs, and follow-up work that matter when reading the project history later.

Entries should be self-contained. Do not rely on private chat history or memory outside the repository. If a decision came from a conversation, summarize the decision and the reasoning in the changelog entry so future readers can understand it from the repo alone.

## File Naming

Use this format:

```text
YYYY-MM-DD-short-change-name.md
```

Example:

```text
2026-05-17-strava-activity-ingestion.md
```

## Entry Structure

```markdown
# YYYY-MM-DD: Change Title

## Summary

Short description of what changed and why.

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
