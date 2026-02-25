Codebase Stats

- Total Go files: 125
- Test Go files: 51
- LOC (including tests): 19,971
- LOC (excluding tests): 7,872
- Measured at (UTC): 2025-11-18 23:41:39Z

Notes

- "LOC" is a raw line count (includes comments and blank lines).
- Non-test Go files = 125 - 51 = 74

Commands used to collect these metrics

```bash
# Count Go files
find . -name "*.go" | wc -l

# Count test files
find . -name "*_test.go" | wc -l

# LOC including tests
find . -name "*.go" -print0 | xargs -0 wc -l | tail -n1 | awk '{print $1}'

# LOC excluding tests
find . -name "*.go" -not -name "*_test.go" -print0 | xargs -0 wc -l | tail -n1 | awk '{print $1}'
```