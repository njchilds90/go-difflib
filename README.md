# go-difflib

> Sequence diffing, patching, and unified diff generation for Go.

[![CI](https://github.com/njchilds90/go-difflib/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/go-difflib/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/go-difflib.svg)](https://pkg.go.dev/github.com/njchilds90/go-difflib)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Go port of the capabilities of Python's `difflib` and Rust's `similar` — designed for both human developers and AI agents.

Zero dependencies. Deterministic. Context-aware.

---

## Install
```bash
go get github.com/njchilds90/go-difflib
```

## Quick Start
```go
import difflib "github.com/njchilds90/go-difflib"

a := difflib.SplitLines("one\ntwo\nthree\n")
b := difflib.SplitLines("one\nTWO\nthree\n")

result := difflib.UnifiedDiff(difflib.DiffInput{
    A:        a,
    B:        b,
    FromFile: "original",
    ToFile:   "modified",
    Context:  3,
})
fmt.Print(result.String())
```

Output:
```
--- original
+++ modified
@@ -1,3 +1,3 @@
 one
-two
+TWO
 three
```

## API Overview

| Function | Description |
|---|---|
| `SplitLines(s)` | Split string into lines preserving newlines |
| `JoinLines(lines)` | Rejoin lines into a string |
| `UnifiedDiff(input)` | Generate a unified diff |
| `ContextDiff(input)` | Generate a context diff |
| `NDiff(a, b)` | Delta-format diff |
| `GetOpCodes(a, b)` | Raw equal/insert/delete/replace opcodes |
| `GetMatchingBlocks(a, b)` | Longest common subsequence blocks |
| `SequenceRatio(a, b)` | Similarity ratio for line slices |
| `StringRatio(a, b)` | Similarity ratio for strings |
| `ClosestMatch(target, candidates)` | Best match from a candidate list |
| `ClosestMatches(target, candidates, n)` | Top-N ranked matches |
| `ApplyPatch(a, patch)` | Apply a unified diff patch |
| `Restore(delta, which)` | Reconstruct A or B from NDiff output |

## License

MIT
```

---
