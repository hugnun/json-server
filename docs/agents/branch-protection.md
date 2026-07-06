# Branch protection: `main`

Recommended rules for the `main` branch. Apply via repo **Settings → Branches → Add rule**.

## Required status checks

All three jobs from `.github/workflows/lint.yml` must pass before merge:

- `golangci-lint`
- `test`
- `govulncheck`

Mark **"Require branches to be up to date before merging"** so PRs re-run checks after a push.

## Other rules

- **Dismiss stale pull request approvals on new commits** — keeps reviews honest.
- **Require linear history** — aligns with the squash-merge policy in `AGENTS.md`.
- **Do not allow force pushes** — `main` is append-only.
- **Do not allow deletions** — `main` is permanent.
- **Allow specified actors to bypass required pull requests** — leave empty. Dependabot PRs are not bypassed; the auto-merge workflow in `.github/workflows/dependabot-auto-merge.yml` gates them on `semver-patch` updates only, after CI is green.

## Why no `CODEOWNERS`

This is a single-maintainer repo. `CODEOWNERS` adds review friction with no enforcement value until a second reviewer appears. Re-evaluate when the first external PR is merged.
