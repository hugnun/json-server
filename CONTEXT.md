# json-server — Domain Glossary

Single source of truth for the names of things in this codebase. ADRs in
`docs/adr/` record the *why* of design decisions; this file records the
*nouns* the codebase reasons about.

## Concepts

### Route
A declared endpoint in the YAML config: a `Path` plus the response rules
attached to it. Owned by the **Router module**; identified by `(name,
method)`.

### Path
The `Path` struct in `internal/config.go`. Carries the URL pattern, HTTP
method, response source, query match map, and optional body-match rule.
Raw YAML form; never seen by the Router.

### ResolvedPath
The normalised form of a `Path` produced by the **ConfigLoader** at parse
time. The `Path.Response` vs `Path.ResponseAdvanced` ambiguity is
resolved here; body-source heuristic (file vs inline JSON) is decided
here; priority class (exact / param / query / body) is computed here
from the route's structural shape. The Router and Response modules
only ever see `ResolvedPath`.

### Priority class
A structural classification of a `ResolvedPath` used by the Router to
order match attempts. Computed by the **ConfigLoader**, not declared.
Order — most specific first:

1. **Exact** — fixed path, no params, no query, no body rule.
2. **Param** — path contains `{name}` placeholders; no query, no body.
3. **Query** — path may or may not contain params; has query match map; no body.
4. **Body** — has a body rule. Tried last so non-body routes short-circuit.

Tiebreaker within a class: declaration order.

### Response
The body, status, headers, and delay that the server sends back for a
matched Route. Built by the **Response module** from either a static
source (file or inline JSON) or a `ResponseAdvanced` rule.

### Match
The decision that a request satisfies a Route's name + method + query +
body rule. Owned by the **Match module**; produces a `TemplateData`
(params, query, body) and a `MatchResult` (`Matched` | `NoMatch` |
`BodyInvalid`). **Match owns the body read** — reads once, returns the
parsed body in `TemplateData.Body` for the Response module to reuse.

### MatchResult
Tri-state outcome of a Match call. `BodyInvalid` maps to HTTP 400 at
the Router.

### Template
The Go-template rendering of a Response body, given a `TemplateData`.
Lives in the **Template module** (`RenderResponse`).

### Middleware
A function that wraps an `http.Handler` to add cross-cutting behaviour
(CORS, logging, default headers). Composed by the **Server module**.

### Server
The HTTP listener. Composed of Middleware wrapping the Router. Built by
the **Server module** (`Run`, future `Build`).

### Config
The parsed YAML document. Loaded by the **ConfigLoader module** from a
file path. Single entry point for both `serve` and `validate` commands.

### ConfigLoader
The module that turns a filesystem path into a `Config`. Single seam
justified by two adapters (CLI serve, CLI validate).

## Module map

| Module | Owns | Interface |
|---|---|---|
| ConfigLoader | path → `Config` (with `[]ResolvedPath`) | `LoadConfig(path string) (Config, error)` |
| Router | request → matched `ResolvedPath` + handler | `New() *Router`; `Add(ResolvedPath, HandlerFunc) error`; `ServeHTTP` — 4 priority buckets, body last |
| Match | request + `ResolvedPath` → `TemplateData` + `MatchResult` | `Match(*http.Request, ResolvedPath) (TemplateData, MatchResult)` |
| Response | request + `ResolvedPath` + data → status, headers, body, deadline | `Resolve(*http.Request, ResolvedPath, TemplateData) (int, http.Header, []byte, time.Time, error)` |
| Template | template body + data → rendered body | `RenderResponse(body string, data TemplateData) (string, error)` |
| Middleware | wrap a handler with cross-cutting concerns | `func(http.Handler) http.Handler` |
| Server | compose + listen | `Build(cfg) (http.Handler, error)`, `Run(cmd, args) error` |
