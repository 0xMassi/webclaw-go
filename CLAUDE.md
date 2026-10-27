# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`github.com/0xMassi/webclaw-go`: Go client for the hosted webclaw REST API (`https://api.webclaw.io`). One flat package, `webclaw`, stdlib only. Keep it that way: no external dependencies. The workspace-level `../CLAUDE.md` covers the API, the sister SDKs, and the dev → staging → main branch flow.

## Commands

```bash
go vet ./...
go test -race ./...                      # what CI runs (.github/workflows/ci.yml), plus go build
go test -run TestLeadBatch_ ./...        # single test or test group, by name regex
```

The `toolchain go1.25.10` line in `go.mod` is a security floor (commit 89fc561), so leave it in place. The README badge still says "Go 1.21+", but `go.mod` actually requires `go 1.23`.

## Architecture

- `webclaw.go` holds `Client`, the functional `Option`s, and `do()`, the single HTTP path. `do()` marshals the body, sets the Bearer token, and turns any non-2xx response into `*APIError`. It takes the message from JSON `message`, then `error`, then the raw body capped at 512 bytes, then the status text. Every endpoint method is a thin wrapper: `var resp T; c.do(ctx, method, path, req, &resp)`.
- `errors.go` holds `APIError` and the `IsAuthError`, `IsRateLimited`, `IsNotFound`, and `IsForbidden` helpers, which use `errors.As`. Paid-only endpoints (X monitors, X audience) return 403.
- There is one file per endpoint family. `types.go` holds the older shared request and response types. Newer endpoints (`research.go`, `lead*.go`, `x_*.go`, `search.go`, `endpoints.go`) define their types in their own file, so follow that for new work.
- Async jobs (crawl, research, lead batch) each have three methods: Start, `Get…`, and `WaitFor…`. The `WaitFor…` loops are deliberately copy-pasted, one per job type, each with its own `…PollOptions`. The shared behavior is: 2s default interval, an optional `Timeout` wraps ctx, and the loop stops on `completed` or `failed`.
- Payloads whose shape varies are `json.RawMessage` (metadata, extract `Data`, vertical `Data`), and callers decode them.
- For request fields where the server default is `true`, or where an explicit `false` or zero must be sent, use pointer types (`*bool`, `*int`) with `omitempty`. See `XMonitorCreateRequest` and `XMonitorUpdateRequest`.
- List endpoints take `(limit, offset int)` and build the query with `url.Values`.

## Conventions

- Tests live in `webclaw_test.go`. Each one uses `newTestServer` (an httptest server plus a client pointed at it) and `assertAuth`, checks the method, path, and decoded request body, and encodes a canned response. Add a success test and at least one error-path test for each new endpoint.
- Retire a field by keeping it and adding a `// Deprecated:` doc comment (see `ResearchRequest.Deep`). Don't delete it, because that would break callers.
- A new or changed endpoint also needs:
  - its README section,
  - a row in the README "All Methods" table,
  - a matching change in `sdk-js/` and `sdk-python/`, because all three SDKs keep identical endpoint coverage.
