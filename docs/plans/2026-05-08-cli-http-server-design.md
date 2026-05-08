# CLI HTTP Server from YAML Config

## Overview

CLI tool that runs an HTTP server based on a YAML configuration file. Supports static JSON files, inline JSON, and dynamic responses with delays and custom status codes.

## Configuration Schema

```yaml
server:
  port: 8080          # default: 8080
  host: "localhost"  # default: ""
  cors:
    enabled: true
    origins: ["*"]
    headers: ["Content-Type"]
  logging:
    enabled: true
    format: "text"  # or "json"
  defaultHeaders:
    X-Custom: "value"

paths:
  # Simple: auto-detect
  - name: users
    method: GET
    response: data/users.json        # file path → load & serve

  - name: users
    method: GET
    response: '{"id": 1, "name": "Bob"}'  # inline JSON

  # Advanced: explicit override
  - name: users/{id}
    method: GET
    response:
      body: '{"id": "{{params.id}}"}'
      status: 200
      delay: 100ms
      headers:
        Content-Type: "application/json"

  # Query matching
  - name: search
    method: GET
    query:
      q: "type:user"  # only match if ?q=type:user
    response:
      body: '[]'

  # Body matching
  - name: login
    method: POST
    body:
      contains: "username"  # match if body contains
    response:
      body: '{"token": "abc123"}'
```

## CLI Interface

```bash
json-server serve config.yaml      # start server
json-server serve config.yaml --port 3000  # override port
json-server validate config.yaml  # validate config only
json-server version              # show version
```

## Routing Engine

- **Path matching**: Exact match for `/users`, wildcard for `/users/{id}` (captures as `{{params.id}}`)
- **Query matching**: Optional `query` map — all specified params must match
- **Body matching**: Optional `body.contains` — string match in request body
- **Priority**: Most specific match wins (exact > params > query)

## Response Templating

- Use `{{params.id}}` for path params
- Use `{{query.name}}` for query params
- Use `{{body.field}}` for body fields (JSON or form)
- Support Go templates for transformations

## Error Handling

- Invalid YAML → clear error with line number
- Missing response file → 404 with helpful message
- Template error → 500 with error details
- Port in use → clear error message

## Testing Strategy

- Unit tests for config parsing
- Unit tests for route matching logic
- Integration tests with actual HTTP requests