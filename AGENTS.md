# Repository Guidelines

## Project Structure & Module Organization

This repository is the `github.com/0xMassi/webclaw-go` SDK: one flat `webclaw` package using only Go’s standard library. `webclaw.go` owns client configuration and the shared HTTP request path; `errors.go` defines API errors. Endpoint families live in files such as `scrape.go`, `search.go`, and `research.go`. Shared models are in `types.go`; newer endpoint-specific models live beside their methods. Tests are in `webclaw_test.go`, usage examples in `README.md`, and CI and banner assets in `.github/`.

## Build, Test, and Development Commands

Use the toolchain declared in `go.mod`: language version Go 1.23, toolchain Go 1.25.10. Preserve the security toolchain floor; CI uses Go 1.25.x.

- `go build ./...` — compile all packages.
- `go vet ./...` — run static checks.
- `go test -race ./...` — run tests with race detection, matching CI.
- `go test -run TestScrape_ ./...` — run a focused test group.
- `gofmt -w <changed-files.go>` — format edited Go files.

This is a library; no local application server is required.

## Coding Style & Naming Conventions

Use standard Go formatting with tabs, exported `PascalCase` names, and unexported `camelCase` names. Document exported symbols. Keep endpoint methods thin and route requests through `Client.do`. Accept `context.Context` for network operations and reuse existing options and error helpers. Preserve JSON field names and use pointers when omitted values must differ from explicit false or zero. Keep dependencies limited to the standard library. Retain deprecated public fields with `// Deprecated:` comments instead of breaking callers.

## Testing Guidelines

Use `testing` and `net/http/httptest`, reusing `newTestServer` and `assertAuth`. Name tests `Test<Operation>_<Scenario>`. For new endpoints, cover success and an error path; verify authentication, method, path, request encoding, and response decoding. Test polling cancellation where relevant. Tests should use local mock servers without real credentials. No numeric coverage threshold is configured.

## Commit & Pull Request Guidelines

History mixes imperative summaries with prefixes such as `feat:`, `fix:`, and `docs(research):`; follow the scoped prefix style when useful. Keep commits focused. PRs should explain the behavior change, link relevant issues, and report build, vet, and race-test results. Update README examples and the All Methods table for endpoint changes, and coordinate endpoint parity with the JavaScript and Python SDKs. Never commit API keys.
