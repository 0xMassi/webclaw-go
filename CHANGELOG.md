# Changelog

## Unreleased

- **Fix:** `ScrapeResponse.ActionsPerformed` is now `[]ActionResult` (`Action`, `Success`, `Error`). It was an `int`, so decoding any scrape response that used browser actions failed.
- **Breaking:** Lead enrichment retired. `Lead`, `LeadBatch`, `GetLeadBatch`, `WaitForLeadBatch` and their `Lead*` request/response types are removed. Use `Extract` with your own schema for structured page extraction.
