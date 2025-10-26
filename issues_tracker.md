# 🚀 Phased Implementation Roadmap

This file organizes tickets into **phases** with **priority** and **dependencies** for proper implementation order. Anyone can pick tickets, but follow the phase dependencies to avoid blocking.

## 📅 Phase 0 — Project Setup (**High Priority**)
Initial project structure and configuration. Foundation for all development.

1. [x] 🔴 **Setup-1**: Create project structure — Set up the basic directory layout for the Go forum application.
2. [x] 🔴 **Setup-2**: Create .gitignore file — Configure Git to ignore build artifacts, dependencies, and sensitive files.
3. [x] 🔴 **Setup-3**: Create LICENSE file — Add appropriate open-source license for the project.
4. [x] 🔴 **Setup-4**: Create README.md with project documentation — Write comprehensive documentation including setup, usage, and API details.
5. [x] 🔴 **Setup-5**: Create todo.md for tracking progress — Establish a checklist for implementation phases and tasks.
6. [x] 🔴 **Setup-6**: Initialize Go module — Set up Go module with proper dependencies and versioning.
7. [x] 🔴 **Setup-7**: Create SQL schema file — Design and create the database schema for users, posts, comments, etc.
8. [x] 🔴 **Setup-8**: Create all directory structure — Build out the complete folder hierarchy for handlers, models, middleware, etc.
9. [x] 🔴 **Setup-9**: Create all empty files with purpose comments — Initialize all source files with TODO comments and basic structure.
10. [x] 🔴 **Setup-10**: Create server package — Implement server initialization, router, and graceful shutdown handling to keep main.go minimal.

## 📅 Phase 1 — Database Layer (**High Priority**)
Focus on database setup and connection. Start here for a solid foundation.

11. [ ] 🔴 **DB-1**: Write tests for database connection — Unit test for database initialization. *Dep: Setup*.
12. [ ] 🔴 **DB-2**: Implement database connection and initialization — Implement DB connection, set package DB, configure PRAGMAs. *Dep: Setup*.
13. [ ] 🔴 **DB-3**: Create database migration logic — Implement migration runner to execute schema.sql safely. *Dep: Setup*.

## 📅 Phase 2 — Models (**High Priority, Parallel after DB**)
Implement all data models once DB is ready. These can be worked on in parallel.

14. [ ] 🔴 **User-1**: Write tests and Implement user model, CRUD operations — Unit tests for user creation, password hashing, lookups. Implement CreateUser, VerifyPassword, GetUserByEmail/ID with bcrypt. *Dep: DB-2*.
15. [ ] 🔴 **Post-1**: Write tests and Implement post model — Unit tests for post CRUD. PostCreate, PostList, PostGetByID. *Dep: DB-2*.
16. [ ] 🔴 **Comment-1**: Write tests and Implement comment model — Unit tests for comment CRUD. Comment model. *Dep: DB-2*.
17. [ ] 🟡 **Category-1**: Write tests and Implement category model & mappings — Unit tests for category CRUD. Create/list categories and map to posts. *Dep: DB-2*.
18. [ ] 🔴 **Session-1**: Write tests and Implement session model — Unit tests for session CRUD with UUID tokens and expiry. Session CRUD with UUID tokens and expiry. *Dep: DB-2*.
19. [ ] 🔴 **Reaction-1**: Write tests and Implement reaction model — Unit tests for reaction CRUD. Reaction model. *Dep: DB-2*.

## 📅 Phase 3 — Authentication System (**High Priority, Parallel after Models**)
Build authentication features once models are ready.

20. [ ] 🔴 **Auth-1**: Write tests and Implement password encryption with bcrypt — Tests for bcrypt password encryption. Password hashing/verification.
21. [ ] 🔴 **Auth-2**: Write tests and Implement user registration handler — Unit tests for registration handler. Handler using user model and bcrypt. *Dep: User-1*.
22. [ ] 🔴 **Auth-3**: Write tests and Implement user login handler — Unit tests for login handler. Handler using user/session models. *Dep: User-1, Session-1*.
23. [ ] 🔴 **Auth-4**: Write tests and Implement logout handler — Unit tests for logout handler. Logout handler. *Dep: Session-1*.

## 📅 Phase 4 — Middleware & Error Handling (**High Priority, Parallel after Auth/Models**)
Implement middleware once auth and sessions are ready.

24. [ ] 🔴 **Middleware-1**: Write tests and Implement session middleware — Tests for session middleware. Read cookie, validate session, attach User to context. *Dep: Session-1*.
25. [ ] 🔴 **Middleware-2**: Write tests and Implement cookie handling with expiration — Tests for cookie handling with expiration. Cookie management.
26. [ ] 🔴 **Middleware-3**: Write tests and Implement authentication middleware — Tests for auth middleware. Auth middleware.
27. [ ] 🟡 **Middleware-4**: Write tests and Implement error handling middleware — Tests for error middleware. Centralized error → flash/status handler.
28. [ ] 🟡 **Middleware-5**: Write tests and Implement proper HTTP status responses — Tests for proper HTTP status responses. HTTP status handling.

## 📅 Phase 5 — Content Handlers (**High Priority, After Models + Auth + Middleware**)
Implement handlers for posts, comments, reactions, and filters once core systems are in place.

29. [ ] 🔴 **Post-2**: Write tests and Implement create post handler — Tests for create post handler. Handler for post creation. *Dep: Post-1, Auth-3*.
30. [ ] 🔴 **Post-3**: Write tests and Implement post listing handler — Tests for post listing handler. Handler for post listing. *Dep: Post-1*.
31. [ ] 🔴 **Post-4**: Write tests and Implement single post view handler — Tests for single post view handler. Handler for single post view. *Dep: Post-1*.
32. [ ] 🟡 **Post-5**: Write tests and Implement category association logic — Tests for category association logic. Post-category association. *Dep: Category-1, Post-1*.
33. [ ] 🔴 **Comment-2**: Write tests and Implement comment creation handler — Tests for comment creation handler. Handler for comment creation. *Dep: Comment-1, Auth-3*.
34. [ ] 🔴 **Comment-3**: Write tests and Implement comment display logic — Tests for comment display logic. Comment listing. *Dep: Comment-1*.
35. [ ] 🔴 **Comment-4**: Write tests and Implement comment validation — Tests for comment validation. Comment validation.
36. [ ] 🔴 **Reaction-2**: Write tests and Implement reaction handler — Tests for reaction handler. Handler for reactions. *Dep: Reaction-1, Auth-3*.
37. [ ] 🔴 **Reaction-3**: Write tests and Implement reaction count display — Tests for reaction count display. Reaction counting. *Dep: Reaction-1*.
38. [ ] 🔴 **Reaction-4**: Write tests and Implement reaction toggle logic — Tests for reaction toggle logic. Prevent duplicate reactions. *Dep: Reaction-1*.
39. [ ] 🟡 **Filter-1**: Write tests and Implement category filter handler — Tests for category filter handler. Category filtering. *Dep: Post-5, Category-1*.
40. [ ] 🟡 **Filter-2**: Write tests and Implement created posts filter — Tests for created posts filter. Created posts filtering. *Dep: Post-5, Auth-3*.
41. [ ] 🟡 **Filter-3**: Write tests and Implement liked posts filter — Tests for liked posts filter. Liked posts filtering. *Dep: Post-5, Reaction-2*.

## 📅 Phase 6 — Frontend, Templates & Static (**Medium Priority, Parallel with Backend**)
Polish the UI and client-side features. Can be worked on in parallel with backend development.

42. [ ] 🟡 **Templates-1**: Create base HTML template — Base template.
43. [ ] 🟡 **Templates-2**: Create home page template — Home page.
44. [ ] 🟡 **Templates-3**: Create registration template — Registration page.
45. [ ] 🟡 **Templates-4**: Create login template — Login page.
46. [ ] 🟡 **Templates-5**: Create post view template — Post view page.
47. [ ] 🟡 **Templates-6**: Create create post template — Create post page.
48. [ ] 🟡 **Static-1**: Create CSS styles — CSS styles.
49. [ ] 🟡 **Static-2**: Create JavaScript for client-side interactions — Client-side JS.

## 📅 Phase 7 — Docker Integration (**Medium Priority**)
Containerize the application.

50. [ ] 🟡 **Docker-1**: Test Docker build — Test Docker build.
51. [ ] 🟡 **Docker-2**: Test Docker deployment — Test Docker deployment.
52. [ ] 🟡 **Docker-3**: Update README with Docker instructions — Update README.

## 📅 Phase 8 — Integration, Docs, and Release (**High Priority for Testing**)
Final testing, documentation, and deployment.

53. [ ] 🔴 **Integration-1**: Write integration tests based on audit.md — Full E2E tests covering main flows, DB reset per test.
54. [ ] 🔴 **Integration-2**: Test all user flows — Test user flows.
55. [ ] 🔴 **Integration-3**: Test edge cases — Test edge cases.
56. [ ] 🔴 **Integration-4**: Test error scenarios — Test error scenarios.
57. [ ] 🔴 **Integration-5**: Test session expiration — Test session expiration.
58. [ ] 🔴 **Integration-6**: Test concurrent access — Test concurrent access.
59. [ ] 🟡 **Docs-1**: Update README with final documentation — Document endpoints, migration  steps, test commands, Docker usage.
60. [ ] 🟢 **Release-1**: Review all code for best practices — Code review.
61. [ ] 🟢 **Release-2**: Verify all requirements are met — Verify requirements.
62. [ ] 🟢 **Release-3**: Performance testing — Performance review.
63. [ ] 🟢 **Release-4**: Security review — Security checks.

## 🏃 Sprint Suggestions (Flexible)
- **Sprint 1**: Focus on Phase 1 and start Phase 2 (DB-1 to User-1). Goal: DB and basic models working.
- **Sprint 2**: Complete Phase 2, Phase 3, and start Phase 4 (Post-1 to Middleware-5). Goal: Auth and middleware ready.
- **Sprint 3**: Finish Phase 4, Phase 5, and start Phase 6 (Middleware-4 to Templates-6). Goal: Full backend and basic frontend.
- **Sprint 4**: Complete Phase 6, Phase 7, and Phase 8. Goal: Polished app with tests and docs.

## ✅ Acceptance Criteria
- **Tests**: Each unit ticket must add failing tests first, then implementation to pass.
- **Handlers**: Return appropriate HTTP status codes and use middleware for auth.
- **DB**: Migrations must be idempotent and run automatically in InitDB.

## Priority Legend
🔴 High Priority
🟡 Medium Priority
🟢 Low Priority