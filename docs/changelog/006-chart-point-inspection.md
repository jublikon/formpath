---
id: changelog-006
number: 6
slug: chart-point-inspection
related_epics:
  - epic-003
related_adrs: []
---

# Changelog 006: Chart Point Inspection

## Summary

Added exact daily-value inspection to all training charts so users can identify
the date and value behind a visible point with a mouse, touch input, or
keyboard navigation.

## Related Epics

- [Epic 003: Activity-Type Graphs](../epics/003-activity-type-graphs.md)

## Related ADRs

- No architectural decision record applies to this frontend interaction.

## Relevant Changes

- Added a shared chart inspector to the cross-sport moving-time chart and the
  running, cycling, and workout charts.
- Added an anchored tooltip, vertical guide, and point marker for the selected
  daily value.
- Added pointer hover, touch drag, and left/right keyboard navigation.
- Added a visible keyboard focus state and a polite live-region announcement
  for the selected date and value.
- Kept the existing hidden list of all 28 values as a non-interactive
  accessible representation.
- Added unit coverage for pointer-coordinate mapping, nearest-point selection,
  keyboard index movement, edge clamping, and empty charts.

## Decisions

- Reuse the existing chart geometry and formatted daily values rather than
  introducing a new chart library or backend endpoint.
- Render the inspector as an HTML overlay while leaving chart drawing in SVG.
  This prevents tooltip text from stretching when responsive CSS changes the
  SVG aspect ratio.
- Allow normal vertical touch scrolling with `touch-action: pan-y` while
  retaining horizontal chart inspection.
- Avoid `role="application"` and expose the keyboard interaction as a focused
  group with live value announcements.
- Keep exact values available for zero-value days as well as training days.

## Verification

- `npm test` passed with 32 unit tests across four test files.
- `npm run lint` passed.
- `npm run build` passed with TypeScript checking and a Vite production build.
- `go test ./...` passed.
- `git diff --check` passed.
- Browser verification covered pointer hover, keyboard navigation, tooltip edge
  placement, responsive layout at 390×844, vertical page scrolling, and browser
  console output.

## Follow-ups

- Add component-level interaction tests if the frontend adopts a DOM test
  environment such as Testing Library with jsdom.
