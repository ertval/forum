# Logger Refactor Plan

Goal: Reduce terminal log verbosity, make human-readable logs shorter and configurable.

Summary of changes made:
- Added a `Config` type to `internal/platform/logger` to control human output.
- Introduced `NewWithConfig` constructor and kept `New` as a convenience.
- For human (terminal) output, timestamps are shortened to seconds by default.
- The `user_agent` field is omitted by default from human output.
- Lines are truncated to `MaxLineWidth` (default 80) with an ellipsis.
 - The `user_agent` field is omitted by default from human output.
 - Human output is restricted by default to essential HTTP info: `url`, `response`, `status`, `error`, `errors`.
 - Lines are truncated to `MaxLineWidth` (default 120) with an ellipsis.

Plan (steps and status):

1. Add `Config` and `TimePrecision` types — Completed
2. Provide `NewWithConfig` and keep `New` as wrapper — Completed
3. Filter out sensitive/verbose fields for human output — Completed
4. Shorten time format to seconds for human output — Completed
5. Add truncation helper to limit human line width — Completed
6. Run tests for logger package and full suite — Ran; logger tests passed, full suite showed unrelated template failures — Incomplete
7. Investigate unrelated unit test failures (templates) — Pending

Notes / rationale:
- JSON output remains unchanged to preserve programmatic consumers.
- Config only affects human (terminal) formatting so behavior is backward compatible.
- Defaults were chosen to be conservative: shorter timestamps, omit `user_agent`, 80-char width.

Next steps I recommend:
- Do you want me to (A) proceed to fix the failing unit tests (they appear to be template baseline mismatches), or (B) keep changes limited to the logger and open a PR with the current, tested logger changes?

If you want (A), I'll investigate the failing tests and propose minimal fixes or updated baselines.
