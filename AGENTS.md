# AGENTS.md

Conventions for AI/human agents working in this repo.

## Stack

- Go 1.26 (see `go.mod`)
- `github.com/spf13/cobra` for CLI
- `gopkg.in/yaml.v3` for config parsing
- `text/template` for response rendering
- Single binary, no runtime deps

## Layout

- `main.go` — entry, delegates to `cmd.Execute()`
- `cmd/` — cobra commands (`serve`, `validate`)
- `internal/` — all real code: ConfigLoader, Router, Match, Response, Template, Server modules
- `examples/` — sample YAML configs and response files
- `docs/adr/` — Architecture Decision Records
- `CONTEXT.md` — Domain glossary (source of truth for naming)
- `docs/agents/` — Agent workflow docs (triage labels, issue tracker, domain)

No public Go API. `internal/` is the only Go package.

## Common tasks

```sh
make test        # go test ./...
make lint        # golangci-lint run
make lint-fix    # golangci-lint run --fix
go run . validate examples/config.yaml
go run . serve examples/config.yaml --port 8080
```

## Testing

- Standard `go test`, no third-party assertion lib
- Tests live next to the code (`*_test.go`)
- `internal/build_test.go` covers end-to-end HTTP via `httptest`
- Run a single test: `go test ./internal -run TestBuild_EndToEnd`
- All tests must pass before merge

## Lint

- golangci-lint v2.11.4 — pinned in `.github/workflows/lint.yml` and `.qlty/qlty.toml`
- Config: `.golangci.yml` (v2 schema)
- Pragmatic baseline: `default: standard` + `revive`, `gocritic`, `errorlint`, `nilerr`, `bodyclose`, `noctx`, `gosec`, `gocognit`, `cyclop`, `dupl`, `unparam`, `goprintffuncname`, `makezero`
- Formatters: `gofmt` (default) + `goimports` with `local-prefixes: github.com/hugnun/json-server`
- Bumping golangci-lint: update **both** `.github/workflows/lint.yml` and `.qlty/qlty.toml`

## Domain docs

- Add a term to `CONTEXT.md` (alphabetical inside each section) before naming
  a struct/func/type after it
- ADRs in `docs/adr/` use the `NNNN-kebab-case-title.md` convention
- Triage labels: see `docs/agents/triage-labels.md`
- Issue workflow: see `docs/agents/issue-tracker.md`

## PRs

- Title: `[json-server] <Title>`
- Run `make lint && make test` before pushing
- Squash merge; commit messages use Conventional Commits
