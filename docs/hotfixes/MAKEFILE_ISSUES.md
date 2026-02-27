<!--
This file lists missing commands/files discovered while verifying `Makefile` targets.
All issues have been RESOLVED.
-->

# Makefile - Issues Found (RESOLVED)

During a verification run the following items referenced by the `Makefile` were not found in the environment/workspace.

**All issues have been fixed:**

## ✅ RESOLVED: Missing command `air`

- **Description**: The `make dev-setup` target attempted to install `air` and the `dev` target expected `air` to be available for hot reload.
- **Resolution**: Removed `air`-related targets (`dev-setup`, `dev`) from Makefile as they are not essential for core development workflow. Developers can install `air` manually if desired: `go install github.com/cosmtrek/air@latest`

## ✅ RESOLVED: Missing file `scripts/run_migrations.go`

- **Description**: The `make migrate` target ran `go run ./scripts/run_migrations.go` but the file didn't exist.
- **Resolution**: Created `scripts/seed/run_migrations.sh` - a bash script that applies SQL migrations from the `migrations/` directory, tracks applied migrations in a `schema_migrations` table, and skips already-applied migrations. Updated Makefile to call this script: `bash ./scripts/seed/run_migrations.sh`

## Additional Improvements

- **New command**: `make test-fail` - Runs all tests and displays only failed test output for easier debugging
- **Fixed .PHONY placement**: Moved all `.PHONY` declarations to after their corresponding comment/target definition (not before)
- **Updated documentation**: Added migration documentation to `migrations/MIGRATIONS_GUIDE.md`, `docs/ARCHITECTURE.md`, `docs/IMPLEMENTATION_ROADMAP.md`, `README.md`, and `.github/copilot-instructions.md`
- **Improved test runner**: Fixed quiet mode spinner bug in `scripts/tests/run_all_tests.sh`, added braille dot spinner with color, shows database verification progress

