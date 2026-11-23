# Changelog

## Unreleased

- `APIError` now has `RequestID string`, read from the `X-Request-Id` response header or the error body's `request_id`. When set, `Error()` ends with `(request id: <id>)`. This changes the error text: code that compares error strings exactly, or snapshot tests of them, needs updating.
- **Fix:** `ScrapeResponse.ActionsPerformed` is now `[]ActionResult` (`Action`, `Success`, `Error`). It was an `int`, so decoding any scrape response that used browser actions failed.
- **Breaking:** Lead enrichment retired. `Lead`, `LeadBatch`, `GetLeadBatch`, `WaitForLeadBatch` and their `Lead*` request/response types are removed. Use `Extract` with your own schema for structured page extraction.
