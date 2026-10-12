# orgfile

Org-mode file primitives shared by the org tools: parsing (with a configurable
TODO keyword set), pretty-printing via go-org's `OrgWriter`, subtree
re-leveling, and atomic file replacement.

Extracted from `org-archiver/internal/orgfile` so `org-archiver`, `org-fmt`,
`org-capture` and `org/web` share one implementation.

| | |
|---|---|
| `Parse` | go-org defaults — recognizes only `TODO` and `DONE` |
| `ParseWithTODO` | registers a keyword spec, e.g. `"TODO NEXT \| DONE"` |
| `Render` | `OrgWriter` with tags aligned to column 80 |
| `ShiftHeadline` | re-level a headline and its descendants |
| `WriteAtomic` | temp file in the same dir, then rename; preserves mode |

`Render` normalizes keyword and drawer casing (`#+title:` to `#+TITLE:`), so a
file is a fixed point only once it has been rendered at least once.
