# Changelog

All notable changes to go-difflib will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-23

### Added
- `SplitLines` / `JoinLines` — safe, round-trip line splitting
- `UnifiedDiff` — unified diff generation with configurable context
- `ContextDiff` — context diff format (like `diff -c`)
- `NDiff` — delta format diff
- `GetMatchingBlocks` — longest common subsequence matching blocks
- `GetOpCodes` — raw opcodes (equal/insert/delete/replace)
- `SequenceRatio` — similarity ratio for line sequences
- `StringRatio` — similarity ratio for raw strings
- `ClosestMatch` — find best match from a candidate list
- `ClosestMatches` — ranked top-N matches from a candidate list
- `ApplyPatch` — apply a unified diff patch string to a line slice
- `Restore` — reconstruct A or B from NDiff output
- `DiffResult.IsEmpty` — check if sequences were equal
- `DiffResult.String` — render unified diff as string
- Full table-driven test suite
- GitHub Actions CI (Go 1.21 / 1.22 / 1.23)
- Zero external dependencies
