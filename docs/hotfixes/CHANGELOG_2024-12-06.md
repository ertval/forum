# Changelog - December 6, 2024

## Bug Fixes & Improvements

### Test Runner Improvements
- **Fixed**: Quiet mode bug in `scripts/tests/run_all_tests.sh` where database verification would hang
- **Improved**: Replaced simple spinner with nicer braille dot spinner (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏) with blue color
- **Added**: Spinner now shows during database verification step in quiet mode
- **Added**: Pass/fail indicators (✓/✗) shown immediately after each test completes in quiet mode

### New Test Command
- **Added**: `make test-fail` command that runs all tests and displays only failed test output
- **Created**: `scripts/tests/run_failed_tests.sh` - captures and displays full output for failed tests only

### Database Migrations
- **Created**: `scripts/seed/run_migrations.sh` - proper bash-based migration runner
  - Automatically applies pending SQL migrations
  - Tracks applied migrations in `schema_migrations` table
  - Skips already-applied migrations
  - Supports transaction-based application
- **Updated**: `make migrate` now calls `bash ./scripts/seed/run_migrations.sh` instead of non-existent Go file

### Makefile Cleanup
- **Fixed**: Moved all `.PHONY` declarations to **after** their corresponding comment (not before)
- **Removed**: `dev-setup` and `dev` targets (removed `air` hot-reload dependency)
- **Updated**: `make help` output to include new `test-fail` command and remove dev commands

### Documentation Updates
All documentation updated with migration information:

1. **migrations/MIGRATIONS_GUIDE.md**
   - Added "Running Migrations" section
   - Documented `make migrate` and manual script usage
   - Explained migration tracking system

2. **docs/ARCHITECTURE.md**
   - Added "Running Migrations" section with manual options
   - Documented `schema_migrations` tracking table
   - Referenced MIGRATIONS_GUIDE.md for detailed info

3. **docs/IMPLEMENTATION_ROADMAP.md**
   - Added manual migration script to Phase 1 checklist

4. **README.md**
   - Added comprehensive "Database Migrations" section
   - Documented manual migration workflow
   - Included migration creation and file structure examples

5. **.github/copilot-instructions.md**
   - Updated Testing Commands section with `make test-fail`
   - Updated Development Commands with migration commands
   - Added migration runner and guide to Key Files Reference

6. **docs/MAKEFILE_ISSUES.md**
   - Updated to show all issues as RESOLVED
   - Documented resolutions for `air` and migration script issues

## Files Changed

**New Files:**
- `scripts/seed/run_migrations.sh` - Migration runner script
- `scripts/tests/run_failed_tests.sh` - Failed tests output script
- `docs/CHANGELOG_2024-12-06.md` - This file

**Modified Files:**
- `scripts/tests/run_all_tests.sh` - Fixed quiet mode, improved spinner
- `Makefile` - Fixed .PHONY placement, removed air, updated migrate, added test-fail
- `migrations/MIGRATIONS_GUIDE.md` - Added running migrations section
- `docs/ARCHITECTURE.md` - Added migration documentation
- `docs/IMPLEMENTATION_ROADMAP.md` - Added migration script to checklist
- `README.md` - Added Database Migrations section
- `.github/copilot-instructions.md` - Updated commands and file reference
- `docs/MAKEFILE_ISSUES.md` - Marked all issues as resolved

## Testing

All scripts verified with bash syntax checking. Makefile help command confirmed working with updated target list.
