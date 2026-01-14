# Code Review: Notification Module

**Date:** 2026-01-14 15:37
**Reviewer:** Antigravity (Principal Software Engineer)

## Executive Summary

The `notification` module (Optional Feature) is a scaffold with interfaces defined but logic missing (similar to Moderation). The domain structure is sound, defining the necessary properties for a notification system. However, since it is not implemented, the review is limited to the design contract.

## Critical Issues (Must Fix)

- **ISSUE-1: Missing Implementation**
  - **Location:** All files in `application/` and `adapters/`.
  - **Probability:** High
  - **Description:** Logic is stubbed with `return nil` or `http.StatusNotImplemented` placeholders.
  - **Proposed Fix:** Implement when feature is scheduled.

## Performance & Optimization

- **PERF-1: Polling vs Push**
  - **Description:** The current design (`GET /api/notifications`) implies a polling architecture. If client-side polling is aggressive, this will pound the database.
  - **Optimization:** For a real-time feel, consider Server-Sent Events (SSE) or WebSockets. If polling, ensure `List` endpoints support "Since ID" or "After Timestamp" to fetch only new items (incremental sync) rather than sending the whole list every time.

## Nitpicks & Best Practices

- **TargetID resolution:** Like the `moderation` module, `notification` stores a `TargetID` (int) and `Type` to refer to posts/comments. Resolving these for the UI (e.g., to generate a link to the post) will require polymorphic queries or multiple round trips.

---
