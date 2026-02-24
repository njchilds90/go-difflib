// Package difflib provides sequence comparison, unified diff generation,
// patch application, and similarity scoring for Go programs and AI agents.
//
// Inspired by Python's difflib and Rust's similar crate, this package
// offers deterministic, context-aware text diffing with structured output.
//
// Basic usage:
//
//	a := "foo\nbar\nbaz\n"
//	b := "foo\nBAR\nbaz\n"
//	diff := difflib.UnifiedDiff(difflib.DiffInput{
//	    A:       difflib.SplitLines(a),
//	    B:       difflib.SplitLines(b),
//	    FromFile: "original",
//	    ToFile:   "modified",
//	    Context: 3,
//	})
//	fmt.Print(diff.String())
package difflib

import (
	"fmt"
	"math"
	"strings"
)

// Op represents a single diff operation kind.
type Op int

const (
	// OpEqual indicates the segment is equal in both sequences.
	OpEqual Op = iota
	// OpInsert indicates the segment was inserted in B.
	OpInsert
	// OpDelete indicates the segment was deleted from A.
	OpDelete
	// OpReplace indicates the segment differs between A and B.
	OpReplace
)

// String returns a human-readable label for the operation.
func (o Op) String() string {
	switch o {
	case OpEqual:
		return "equal"
	case OpInsert:
		return "insert"
	case OpDelete:
		return "delete"
	case OpReplace:
		return "replace"
	default:
		return "unknown"
	}
}

// OpCode describes a contiguous block of changes between two sequences.
// It mirrors Python's difflib SequenceMatcher opcode format.
type OpCode struct {
	// Tag is the operation kind.
	Tag Op
	// I1, I2 are the start and end indices in sequence A (exclusive end).
	I1, I2 int
	// J1, J2 are the start and end indices in sequence B (exclusive end).
	J1, J2 int
}

// Hunk represents a contiguous group of changed lines in a unified diff,
// along with surrounding context lines.
type Hunk struct {
	// OldStart is the 1-based start line in the original file.
	OldStart int
	// OldLines is the number of lines from the original file in this hunk.
	OldLines int
	// NewStart is the 1-based start line in the new file.
	NewStart int
	// NewLines is the number of lines from the new file in this hunk.
	NewLines int
	// Lines contains the raw diff lines prefixed with ' ', '+', or '-'.
	Lines []string
}

// DiffResult holds a complete unified diff result.
type DiffResult struct {
	// FromFile is the label for the original file.
	FromFile string
	// ToFile is the label for the modified file.
	ToFile string
	// Hunks contains the diff hunks.
	Hunks []Hunk
}

// String renders the DiffResult as a standard unified diff string.
func (d DiffResult) String() string {
	if len(d.Hunks) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("--- %s\n", d.FromFile))
	b.WriteString(fmt.Sprintf("+++ %s\n", d.ToFile))
	for _, h := range d.Hunks {
		b.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			h.OldStart, h.OldLines, h.NewStart, h.NewLines))
		for _, l := range h.Lines {
			b.WriteString(l)
		}
	}
	return b.String()
}

// IsEmpty reports whether the diff contains no changes.
func (d DiffResult) IsEmpty() bool {
	return len(d.Hunks) == 0
}

// DiffInput holds the parameters for generating a unified diff.
type DiffInput struct {
	// A is the original sequence of lines.
	A []string
	// B is the modified sequence of lines.
	B []string
	// FromFile is the label for the original content (e.g., "a/file.go").
	FromFile string
	// ToFile is the label for the modified content (e.g., "b/file.go").
	ToFile string
	// Context is the number of unchanged lines to include around each change.
	// Defaults to 3 if zero.
	Context int
}

// SplitLines splits a string into lines preserving line endings.
// Each line retains its trailing newline if present. This matches
// the behavior expected by UnifiedDiff and makes round-tripping safe.
//
// Example:
//
//	SplitLines("foo\nbar\n") // => []string{"foo\n", "bar\n"}
func SplitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.SplitAfter(s, "\n")
	// SplitAfter leaves a trailing empty string if s ends with \n
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// JoinLines joins a slice of lines (as returned by SplitLines) into a single string.
func JoinLines(lines []string) string {
	return strings.Join(lines, "")
}

// UnifiedDiff computes a unified diff between input.A and input.B.
// Returns a DiffResult with IsEmpty() == true if the sequences are equal.
//
// Example:
//
//	result := difflib.UnifiedDiff(difflib.DiffInput{
//	    A:        difflib.SplitLines("one\ntwo\nthree\n"),
//	    B:        difflib.SplitLines("one\nTWO\nthree\n"),
//	    FromFile: "original",
//	    ToFile:   "modified",
//	    Context:  3,
//	})
//	fmt.Print(result.String())
func UnifiedDiff(input DiffInput) DiffResult {
	ctx := input.Context
	if ctx == 0 {
		ctx = 3
	}

	matcher := newMatcher(input.A, input.B)
	opcodes := matcher.GetOpCodes()

	result := DiffResult{
		FromFile: input.FromFile,
		ToFile:   input.ToFile,
	}

	// Group opcodes into hunks separated by context
	groups := groupOpcodes(opcodes, ctx)
	for _, group := range groups {
		hunk := buildHunk(input.A, input.B, group)
		result.Hunks = append(result.Hunks, hunk)
	}
	return result
}

// SequenceMatch holds information about a matching block between two sequences.
type SequenceMatch struct {
	// A is the start index in sequence A.
	A int
	// B is the start index in sequence B.
	B int
	// Size is the length of the matching block.
	Size int
}

// GetMatchingBlocks returns a list of matching blocks between two line sequences.
// The final entry is always a sentinel with Size == 0.
//
// Example:
//
//	blocks := difflib.GetMatchingBlocks(
//	    difflib.SplitLines("a\nb\nc\n"),
//	    difflib.SplitLines("a\nX\nc\n"),
//	)
func GetMatchingBlocks(a, b []string) []SequenceMatch {
	m := newMatcher(a, b)
	return m.GetMatchingBlocks()
}

// GetOpCodes returns the opcodes describing how to transform sequence A into B.
//
// Example:
//
//	codes := difflib.GetOpCodes(
//	    difflib.SplitLines("foo\nbar\n"),
//	    difflib.SplitLines("foo\nbaz\n"),
//	)
func GetOpCodes(a, b []string) []OpCode {
	m := newMatcher(a, b)
	return m.GetOpCodes()
}

// SequenceRatio returns a similarity ratio in [0.0, 1.0] between two line sequences.
// 1.0 means identical; 0.0 means completely different.
//
// Example:
//
//	ratio := difflib.SequenceRatio(
//	    difflib.SplitLines("foo\nbar\n"),
//	    difflib.SplitLines("foo\nbaz\n"),
//	)
func SequenceRatio(a, b []string) float64 {
	m := newMatcher(a, b)
	return m.Ratio()
}

// StringRatio returns a similarity ratio in [0.0, 1.0] between two raw strings
// compared character by character.
//
// Example:
//
//	ratio := difflib.StringRatio("kitten", "sitting") // ~0.615
func StringRatio(a, b string) float64 {
	ar := []rune(a)
	br := []rune(b)
	as := make([]string, len(ar))
	bs := make([]string, len(br))
	for i, r := range ar {
		as[i] = string(r)
	}
	for i, r := range br {
		bs[i] = string(r)
	}
	return SequenceRatio(as, bs)
}

// ContextDiff generates a context diff (like `diff -c`) between A and B.
// Returns lines suitable for display, each prefixed with '  ', '+ ', '- ', or '! '.
//
// Example:
//
//	lines := difflib.ContextDiff(difflib.DiffInput{
//	    A:        difflib.SplitLines("one\ntwo\nthree\n"),
//	    B:        difflib.SplitLines("one\nTWO\nthree\n"),
//	    FromFile: "original",
//	    ToFile:   "modified",
//	    Context:  3,
//	})
//	fmt.Println(strings.Join(lines, ""))
func ContextDiff(input DiffInput) []string {
	ctx := input.Context
	if ctx == 0 {
		ctx = 3
	}
	matcher := newMatcher(input.A, input.B)
	opcodes := matcher.GetOpCodes()
	groups := groupOpcodes(opcodes, ctx)

	if len(groups) == 0 {
		return nil
	}

	var out []string
	out = append(out, fmt.Sprintf("*** %s\n", input.FromFile))
	out = append(out, fmt.Sprintf("--- %s\n", input.ToFile))

	for _, group := range groups {
		first, last := group[0], group[len(group)-1]
		out = append(out, "***************\n")
		out = append(out, fmt.Sprintf("*** %d,%d ****\n", first.I1+1, last.I2))
		for _, op := range group {
			switch op.Tag {
			case OpEqual:
				for _, l := range input.A[op.I1:op.I2] {
					out = append(out, "  "+l)
				}
			case OpReplace, OpDelete:
				for _, l := range input.A[op.I1:op.I2] {
					out = append(out, "! "+l)
				}
			}
		}
		out = append(out, fmt.Sprintf("--- %d,%d ----\n", first.J1+1, last.J2))
		for _, op := range group {
			switch op.Tag {
			case OpEqual:
				for _, l := range input.B[op.J1:op.J2] {
					out = append(out, "  "+l)
				}
			case OpReplace, OpInsert:
				for _, l := range input.B[op.J1:op.J2] {
					out = append(out, "! "+l)
				}
			}
		}
	}
	return out
}

// NDiff generates a delta-format diff similar to Python's ndiff,
// showing every line with a prefix: '  ' (equal), '+ ' (insert), '- ' (delete).
//
// Example:
//
//	lines := difflib.NDiff(
//	    difflib.SplitLines("one\ntwo\nthree\n"),
//	    difflib.SplitLines("one\nTWO\nthree\n"),
//	)
func NDiff(a, b []string) []string {
	matcher := newMatcher(a, b)
	opcodes := matcher.GetOpCodes()
	var out []string
	for _, op := range opcodes {
		switch op.Tag {
		case OpEqual:
			for _, l := range a[op.I1:op.I2] {
				out = append(out, "  "+l)
			}
		case OpInsert:
			for _, l := range b[op.J1:op.J2] {
				out = append(out, "+ "+l)
			}
		case OpDelete:
			for _, l := range a[op.I1:op.I2] {
				out = append(out, "- "+l)
			}
		case OpReplace:
			for _, l := range a[op.I1:op.I2] {
				out = append(out, "- "+l)
			}
			for _, l := range b[op.J1:op.J2] {
				out = append(out, "+ "+l)
			}
		}
	}
	return out
}

// ApplyPatch applies a unified diff string to the original lines A,
// returning the patched result or an error if the patch does not apply cleanly.
//
// Example:
//
//	patched, err := difflib.ApplyPatch(original, patchString)
func ApplyPatch(a []string, patch string) ([]string, error) {
	lines := strings.Split(patch, "\n")
	// Skip header lines (--- and +++)
	i := 0
	for i < len(lines) && (strings.HasPrefix(lines[i], "---") || strings.HasPrefix(lines[i], "+++")) {
		i++
	}

	result := make([]string, len(a))
	copy(result, a)
	offset := 0

	for i < len(lines) {
		line := lines[i]
		if !strings.HasPrefix(line, "@@") {
			i++
			continue
		}
		// Parse @@ -start,count +start,count @@
		var oldStart, oldCount, newStart, newCount int
		_, err := fmt.Sscanf(line, "@@ -%d,%d +%d,%d @@", &oldStart, &oldCount, &newStart, &newCount)
		if err != nil {
			// Try without counts
			_, err = fmt.Sscanf(line, "@@ -%d +%d @@", &oldStart, &newStart)
			if err != nil {
				return nil, fmt.Errorf("difflib: malformed hunk header: %q", line)
			}
			oldCount, newCount = 1, 1
		}
		i++

		pos := oldStart - 1 + offset
		var removes, inserts []string

		for i < len(lines) {
			l := lines[i]
			if strings.HasPrefix(l, "@@") || (strings.HasPrefix(l, "---") && i > 0) {
				break
			}
			if strings.HasPrefix(l, "-") {
				removes = append(removes, strings.TrimPrefix(l, "-"))
				i++
			} else if strings.HasPrefix(l, "+") {
				inserts = append(inserts, strings.TrimPrefix(l, "+"))
				i++
			} else if strings.HasPrefix(l, " ") {
				i++
			} else {
				i++
			}
		}

		// Verify removes match
		for ri, rem := range removes {
			actual := strings.TrimRight(result[pos+ri], "")
			expected := strings.TrimRight(rem, "")
			if actual != expected {
				return nil, fmt.Errorf("difflib: patch mismatch at line %d: expected %q, got %q",
					pos+ri+1, expected, actual)
			}
		}

		// Apply: splice out removes, splice in inserts
		before := result[:pos]
		after := result[pos+len(removes):]
		next := make([]string, 0, len(before)+len(inserts)+len(after))
		next = append(next, before...)
		next = append(next, inserts...)
		next = append(next, after...)
		result = next
		offset += len(inserts) - len(removes)
	}
	return result, nil
}

// Restore returns either the A or B sequence reconstructed from ndiff output.
// which must be 1 (original) or 2 (modified).
//
// Example:
//
//	ndiff := difflib.NDiff(a, b)
//	original := difflib.Restore(ndiff, 1)
//	modified := difflib.Restore(ndiff, 2)
func Restore(delta []string, which int) []string {
	var tag string
	if which == 1 {
		tag = "- "
	} else {
		tag = "+ "
	}
	var out []string
	for _, l := range delta {
		if strings.HasPrefix(l, "  ") {
			out = append(out, strings.TrimPrefix(l, "  "))
		} else if strings.HasPrefix(l, tag) {
			out = append(out, strings.TrimPrefix(l, tag))
		}
	}
	return out
}

// ClosestMatch finds the string from candidates most similar to target,
// returning the best match and its similarity ratio.
// Returns ("", 0) if candidates is empty.
//
// Example:
//
//	best, ratio := difflib.ClosestMatch("appel", []string{"apple", "mango", "apply"})
//	// best == "apple", ratio ~= 0.888
func ClosestMatch(target string, candidates []string) (string, float64) {
	best := ""
	bestRatio := -1.0
	for _, c := range candidates {
		r := StringRatio(target, c)
		if r > bestRatio {
			bestRatio = r
			best = c
		}
	}
	if best == "" {
		return "", 0
	}
	return best, bestRatio
}

// ClosestMatches returns up to n candidates from the list sorted by similarity
// to target, highest first.
//
// Example:
//
//	matches := difflib.ClosestMatches("appel", []string{"apple", "mango", "apply", "apt"}, 2)
func ClosestMatches(target string, candidates []string, n int) []string {
	type ranked struct {
		s string
		r float64
	}
	ranked_list := make([]ranked, 0, len(candidates))
	for _, c := range candidates {
		ranked_list = append(ranked_list, ranked{c, StringRatio(target, c)})
	}
	// Simple insertion sort (n is typically small)
	for i := 1; i < len(ranked_list); i++ {
		for j := i; j > 0 && ranked_list[j].r > ranked_list[j-1].r; j-- {
			ranked_list[j], ranked_list[j-1] = ranked_list[j-1], ranked_list[j]
		}
	}
	if n > len(ranked_list) {
		n = len(ranked_list)
	}
	out := make([]string, n)
	for i := range out {
		out[i] = ranked_list[i].s
	}
	return out
}

// --- Internal: sequence matcher ---

type matcher struct {
	a, b    []string
	b2j     map[string][]int
	matches []SequenceMatch
}

func newMatcher(a, b []string) *matcher {
	m := &matcher{a: a, b: b}
	m.buildB2J()
	return m
}

func (m *matcher) buildB2J() {
	m.b2j = make(map[string][]int, len(m.b))
	for i, s := range m.b {
		m.b2j[s] = append(m.b2j[s], i)
	}
}

func (m *matcher) findLongestMatch(alo, ahi, blo, bhi int) SequenceMatch {
	bestI, bestJ, bestSize := alo, blo, 0
	j2len := make(map[int]int)
	for i := alo; i < ahi; i++ {
		newJ2len := make(map[int]int)
		for _, j := range m.b2j[m.a[i]] {
			if j < blo {
				continue
			}
			if j >= bhi {
				break
			}
			k := j2len[j-1] + 1
			newJ2len[j] = k
			if k > bestSize {
				bestI, bestJ, bestSize = i-k+1, j-k+1, k
			}
		}
		j2len = newJ2len
	}
	return SequenceMatch{bestI, bestJ, bestSize}
}

func (m *matcher) GetMatchingBlocks() []SequenceMatch {
	queue := [][4]int{{0, len(m.a), 0, len(m.b)}}
	var blocks []SequenceMatch
	for len(queue) > 0 {
		q := queue[0]
		queue = queue[1:]
		alo, ahi, blo, bhi := q[0], q[1], q[2], q[3]
		match := m.findLongestMatch(alo, ahi, blo, bhi)
		if match.Size > 0 {
			blocks = append(blocks, match)
			if alo < match.A && blo < match.B {
				queue = append(queue, [4]int{alo, match.A, blo, match.B})
			}
			if match.A+match.Size < ahi && match.B+match.Size < bhi {
				queue = append(queue, [4]int{match.A + match.Size, ahi, match.B + match.Size, bhi})
			}
		}
	}
	// Sort blocks by A index
	sortMatchBlocks(blocks)
	// Merge adjacent
	var merged []SequenceMatch
	i1, j1, k1 := 0, 0, 0
	for _, b := range blocks {
		if i1+k1 == b.A && j1+k1 == b.B {
			k1 += b.Size
		} else {
			if k1 > 0 {
				merged = append(merged, SequenceMatch{i1, j1, k1})
			}
			i1, j1, k1 = b.A, b.B, b.Size
		}
	}
	if k1 > 0 {
		merged = append(merged, SequenceMatch{i1, j1, k1})
	}
	merged = append(merged, SequenceMatch{len(m.a), len(m.b), 0})
	return merged
}

func sortMatchBlocks(blocks []SequenceMatch) {
	// Insertion sort by A
	for i := 1; i < len(blocks); i++ {
		for j := i; j > 0 && blocks[j].A < blocks[j-1].A; j-- {
			blocks[j], blocks[j-1] = blocks[j-1], blocks[j]
		}
	}
}

func (m *matcher) GetOpCodes() []OpCode {
	blocks := m.GetMatchingBlocks()
	var codes []OpCode
	i, j := 0, 0
	for _, b := range blocks {
		tag := OpEqual
		if i < b.A && j < b.B {
			tag = OpReplace
		} else if i < b.A {
			tag = OpDelete
		} else if j < b.B {
			tag = OpInsert
		} else {
			tag = OpEqual
		}
		if i < b.A || j < b.B {
			codes = append(codes, OpCode{tag, i, b.A, j, b.B})
		}
		i, j = b.A+b.Size, b.B+b.Size
		if b.Size > 0 {
			codes = append(codes, OpCode{OpEqual, b.A, i, b.B, j})
		}
	}
	return codes
}

func (m *matcher) Ratio() float64 {
	blocks := m.GetMatchingBlocks()
	matches := 0
	for _, b := range blocks {
		matches += b.Size
	}
	total := len(m.a) + len(m.b)
	if total == 0 {
		return 1.0
	}
	return 2.0 * float64(matches) / float64(total)
}

// groupOpcodes groups opcodes into hunks, each surrounded by up to `ctx` equal lines.
func groupOpcodes(codes []OpCode, ctx int) [][]OpCode {
	if len(codes) == 0 {
		return nil
	}
	// Filter leading/trailing equal blocks
	var filtered []OpCode
	for _, c := range codes {
		if c.Tag == OpEqual {
			i1 := int(math.Max(float64(c.I1), float64(c.I2-ctx)))
			i2 := int(math.Min(float64(c.I2), float64(c.I1+ctx)))
			j1 := c.J1 + (i1 - c.I1)
			j2 := c.J1 + (i2 - c.I1)
			filtered = append(filtered, OpCode{c.Tag, i1, i2, j1, j2})
		} else {
			filtered = append(filtered, c)
		}
	}

	var groups [][]OpCode
	var group []OpCode
	for _, c := range filtered {
		if c.Tag == OpEqual && c.I2-c.I1 > ctx*2 {
			// End of hunk: keep only first ctx lines
			head := OpCode{OpEqual, c.I1, c.I1 + ctx, c.J1, c.J1 + ctx}
			group = append(group, head)
			groups = append(groups, group)
			group = nil
			// Start new hunk with last ctx lines
			tail := OpCode{OpEqual, c.I2 - ctx, c.I2, c.J2 - ctx, c.J2}
			group = append(group, tail)
		} else {
			group = append(group, c)
		}
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}
	return groups
}

func buildHunk(a, b []string, group []OpCode) Hunk {
	first, last := group[0], group[len(group)-1]
	hunk := Hunk{
		OldStart: first.I1 + 1,
		OldLines: last.I2 - first.I1,
		NewStart: first.J1 + 1,
		NewLines: last.J2 - first.J1,
	}
	for _, op := range group {
		switch op.Tag {
		case OpEqual:
			for _, l := range a[op.I1:op.I2] {
				hunk.Lines = append(hunk.Lines, " "+l)
			}
		case OpInsert:
			for _, l := range b[op.J1:op.J2] {
				hunk.Lines = append(hunk.Lines, "+"+l)
			}
		case OpDelete:
			for _, l := range a[op.I1:op.I2] {
				hunk.Lines = append(hunk.Lines, "-"+l)
			}
		case OpReplace:
			for _, l := range a[op.I1:op.I2] {
				hunk.Lines = append(hunk.Lines, "-"+l)
			}
			for _, l := range b[op.J1:op.J2] {
				hunk.Lines = append(hunk.Lines, "+"+l)
			}
		}
	}
	return hunk
}
