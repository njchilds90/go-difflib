# Contributing to go-difflib

Thank you for your interest in contributing!

## Getting Started

1. Fork the repository on GitHub.
2. Clone your fork and create a new branch.
3. Make your changes, including tests.
4. Run `go test ./...` and `go vet ./...` locally.
5. Open a pull request against `main`.

## Code Standards

- All public functions must have GoDoc comments.
- All new features require table-driven tests.
- No external dependencies may be introduced.
- `go vet` must pass cleanly.
- Prefer deterministic, pure functions wherever possible.

## Reporting Issues

Please open a GitHub Issue describing:
- What you expected
- What actually happened
- A minimal reproducible example if possible

## License

By contributing, you agree your contributions will be licensed under the MIT License.
