# ADR-0001 — Priority classification for route matching

- **Status:** Accepted
- **Date:** 2026-07-05
- **Context:** C1 of the architecture review (deepening the response pipeline).
- **Supersedes:** The priority note in `docs/plans/2026-05-08-cli-http-server-design.md` ("Priority: Most specific match wins (exact > params > query)"), which left body-rule ordering and tiebreakers undefined.

## Context

The Router needed to choose between multiple matching routes for one request. The original design specified "most specific match wins" but the implementation does a linear scan and picks the first append. With body-rule routes, the body is only readable inside the handler — putting a body-rule route first means a body read happens for every request that could otherwise have matched a non-body route.

## Decision

Priority is **structural**, computed by the **ConfigLoader** from the `ResolvedPath`'s shape. It is not user-declared in YAML.

Four classes, evaluated in this order — most specific first:

1. **Exact** — fixed path, no `{param}` placeholders, no query match map, no body rule.
2. **Param** — path contains one or more `{name}` placeholders; no query, no body.
3. **Query** — has a query match map; no body rule. (May also contain params.)
4. **Body** — has a body rule. Evaluated last so non-body routes short-circuit.

**Tiebreaker within a class:** declaration order. First declared route of a given class wins when multiple match.

The Router holds four internal buckets (one per class), populated in declaration order at `Add()` time. `ServeHTTP` walks the buckets in class order and returns the first match.

## Consequences

- The ConfigLoader becomes the single place that knows the priority rules. Adding a class (e.g. a future "host" class for vhosts) means editing one classifier and one bucket.
- Body reads happen only for routes that nothing cheaper matched. Cheaper routes short-circuit. Body cost is bounded.
- Users cannot express ambiguous priorities in YAML. If two routes of the same class both match, the first wins — no startup error. This matches the existing append-order behaviour and avoids breaking configs.
- The structural rule means a body-rule route will never out-prioritise a non-body route. If a user wants the body rule to win, they must remove the non-body route or accept the structural order. This is documented as a known constraint.
- The Router's interface stays small: `Add(ResolvedPath, HandlerFunc) error` plus `ServeHTTP`. No priority field is exposed in the public interface.
- The priority classifier is pure and unit-testable without an HTTP server — sits in `ConfigLoader` tests.
