## Editing Palettes

- Do not change any palette file or palette keys without explicitly aligning all dependent masks/sprites and confirming with the user.
- If a palette error occurs, fix it by adjusting the palette minimally or by updating the masks—never by ad‑hoc palette edits alone.

## Selecting and Using Colors

- Avoid directly creating colors using `grid.Style`. Prefer loading colors from a palette. Confirm with the user before directly creating colors.

## Proposal Authoring

- Proposals should include an appendix with sample code.
- API documentation must be exhaustive (list every type, function, and method in scope).
- Use this general format unless asked otherwise:
  - Title
  - Summary
  - Why This Is Useful
  - Feature Overview
  - Fit With Existing Architecture
  - Proposed API (exhaustive)
  - Implementation Plan
  - Alternatives Considered
  - Risks / Tradeoffs
  - Testing Plan
  - Appendix: Sample Code
- Put new proposals in ./docs/proposals