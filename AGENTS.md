## Agent skills

### Issue tracker

Issues live as local markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical roles use their default names: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.

### Versioning

`cmd/mybible/main.go` holds the only version string, semver `major.minor.patch`. Bump the patch for fixes, the minor for features; while below 1.0 a minor bump may break things. Tag the bump `v<version>` so the tag and the constant stay in step. Tags stay local until we publish releases.
