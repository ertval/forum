# Code Review: Moderation Module

**Date:** 2026-01-14 15:34
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `moderation` module is currently a scaffold with defined interfaces and domain entities but lacks concrete implementation. Most methods are placeholders returning `nil` or "Not implemented". This is expected for an optional feature marked as `[OPTIONAL FEATURE: forum-moderation]`. The structure follows the project's architectural patterns well.

## Critical Issues (Must Fix)

- **ISSUE-1: Missing Implementation**
  - **Location:** All files in `application/` and `adapters/`.
  - **Probability:** High (Guaranteed failure)
  - **Description:** Returns `nil` or panic-like behavior (empty returns) for all operations. This code is currently dead/stubbed.
  - **Proposed Fix:** Implement the logic when the feature is prioritized.

## Performance & Optimization

- **PERF-1: N+1 Queries in ListReports (Potential)**
  - **Description:** When `ListReports` is implemented, it will likely need to enrich the `Report` entities with `Reporter` (user) and `Target` (post/comment) details for the UI. Doing this in a loop (fetching user per report, fetching post per report) will result in N+1 queries.
  - **Optimized Code:** Use `JOIN`s in the SQLite repository or batch fetching not to kill performance.

## Nitpicks & Best Practices

- **TargetID Type:** The `TargetID` is an `int` (internal ID). To support different target types (post vs comment), the `TargetType` string discriminator is used. Ensure that when resolving `targetPublicID` (UUID) to `TargetID` (int) in `CreateReport`, the correct table (posts or comments) is queried based on `TargetType`. This "polymorphic association" can be fragile in SQL without strict integrity checks.

---
