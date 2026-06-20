@AGENTS.md

# Claude Code Notes

- Use plan mode for larger changes.
- Prefer small, reviewable diffs.
- Use subagents or focused review passes for broad codebase exploration.
- Do not edit secrets or local-only configuration.
- Do not use `claude`, `codex`, or another coding-agent name in branch names unless explicitly requested by the user.
- Keep commits, pull requests, changelogs, and other project history free of coding-agent attribution.
