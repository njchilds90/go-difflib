package difflib_test

import (
	"strings"
	"testing"

	difflib "github.com/njchilds90/go-difflib"
)

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"single no newline", "foo", []string{"foo"}},
		{"single with newline", "foo\n", []string{"foo\n"}},
		{"multiple", "foo\nbar\nbaz\n", []string{"foo\n", "bar\n", "baz\n"}},
		{"no trailing newline", "foo\nbar", []string{"foo\n", "bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := difflib.SplitLines(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("SplitLines(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SplitLines(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestJoinLines(t *testing.T) {
	input := "foo\nbar\nbaz\n"
	lines := difflib.SplitLines(input)
	got := difflib.JoinLines(lines)
	if got != input {
		t.Errorf("JoinLines(SplitLines(%q)) = %q, want %q", input, got, input)
	}
}

func TestUnifiedDiffEqual(t *testing.T) {
	a := difflib.SplitLines("foo\nbar\nbaz\n")
	result := difflib.UnifiedDiff(difflib.DiffInput{
		A: a, B: a,
		FromFile: "a", ToFile: "b",
		Context: 3,
	})
	if !result.IsEmpty() {
		t.Errorf("expected empty diff for equal sequences, got:\n%s", result.String())
	}
}

func TestUnifiedDiffReplace(t *testing.T) {
	a := difflib.SplitLines("one\ntwo\nthree\n")
	b := difflib.SplitLines("one\nTWO\nthree\n")
	result := difflib.UnifiedDiff(difflib.DiffInput{
		A: a, B: b,
		FromFile: "original", ToFile: "modified",
		Context: 3,
	})
	s := result.String()
	if !strings.Contains(s, "-two\n") {
		t.Errorf("expected '-two\\n' in diff, got:\n%s", s)
	}
	if !strings.Contains(s, "+TWO\n") {
		t.Errorf("expected '+TWO\\n' in diff, got:\n%s", s)
	}
}

func TestUnifiedDiffInsert(t *testing.T) {
	a := difflib.SplitLines("one\nthree\n")
	b := difflib.SplitLines("one\ntwo\nthree\n")
	result := difflib.UnifiedDiff(difflib.DiffInput{
		A: a, B: b,
		FromFile: "a", ToFile: "b",
	})
	if !strings.Contains(result.String(), "+two\n") {
		t.Errorf("expected '+two\\n' in diff:\n%s", result.String())
	}
}

func TestUnifiedDiffDelete(t *testing.T) {
	a := difflib.SplitLines("one\ntwo\nthree\n")
	b := difflib.SplitLines("one\nthree\n")
	result := difflib.UnifiedDiff(difflib.DiffInput{
		A: a, B: b,
		FromFile: "a", ToFile: "b",
	})
	if !strings.Contains(result.String(), "-two\n") {
		t.Errorf("expected '-two\\n' in diff:\n%s", result.String())
	}
}

func TestUnifiedDiffHeaders(t *testing.T) {
	a := difflib.SplitLines("x\n")
	b := difflib.SplitLines("y\n")
	result := difflib.UnifiedDiff(difflib.DiffInput{
		A: a, B: b,
		FromFile: "old.txt", ToFile: "new.txt",
	})
	s := result.String()
	if !strings.HasPrefix(s, "--- old.txt\n+++ new.txt\n") {
		t.Errorf("unexpected headers:\n%s", s)
	}
}

func TestSequenceRatioIdentical(t *testing.T) {
	a := difflib.SplitLines("foo\nbar\n")
	ratio := difflib.SequenceRatio(a, a)
	if ratio != 1.0 {
		t.Errorf("expected 1.0 for identical sequences, got %f", ratio)
	}
}

func TestSequenceRatioEmpty(t *testing.T) {
	ratio := difflib.SequenceRatio(nil, nil)
	if ratio != 1.0 {
		t.Errorf("expected 1.0 for empty sequences, got %f", ratio)
	}
}

func TestSequenceRatioDifferent(t *testing.T) {
	a := difflib.SplitLines("foo\n")
	b := difflib.SplitLines("bar\n")
	ratio := difflib.SequenceRatio(a, b)
	if ratio >= 1.0 || ratio < 0.0 {
		t.Errorf("unexpected ratio %f for different sequences", ratio)
	}
}

func TestStringRatio(t *testing.T) {
	ratio := difflib.StringRatio("kitten", "kitten")
	if ratio != 1.0 {
		t.Errorf("expected 1.0, got %f", ratio)
	}
	ratio2 := difflib.StringRatio("kitten", "sitting")
	if ratio2 <= 0 || ratio2 >= 1 {
		t.Errorf("expected ratio in (0,1) for kitten/sitting, got %f", ratio2)
	}
}

func TestGetMatchingBlocks(t *testing.T) {
	a := difflib.SplitLines("a\nb\nc\n")
	b := difflib.SplitLines("a\nX\nc\n")
	blocks := difflib.GetMatchingBlocks(a, b)
	// Sentinel block at the end always has Size 0
	last := blocks[len(blocks)-1]
	if last.Size != 0 {
		t.Errorf("last matching block should be sentinel with Size=0, got %+v", last)
	}
}

func TestGetOpCodes(t *testing.T) {
	a := difflib.SplitLines("foo\nbar\nbaz\n")
	b := difflib.SplitLines("foo\nBAR\nbaz\n")
	codes := difflib.GetOpCodes(a, b)
	found := false
	for _, c := range codes {
		if c.Tag == difflib.OpReplace {
			found = true
		}
	}
	if !found {
		t.Errorf("expected at least one OpReplace code, got %v", codes)
	}
}

func TestOpString(t *testing.T) {
	cases := []struct {
		op   difflib.Op
		want string
	}{
		{difflib.OpEqual, "equal"},
		{difflib.OpInsert, "insert"},
		{difflib.OpDelete, "delete"},
		{difflib.OpReplace, "replace"},
	}
	for _, c := range cases {
		if c.op.String() != c.want {
			t.Errorf("Op(%d).String() = %q, want %q", c.op, c.op.String(), c.want)
		}
	}
}

func TestNDiff(t *testing.T) {
	a := difflib.SplitLines("one\ntwo\nthree\n")
	b := difflib.SplitLines("one\nTWO\nthree\n")
	lines := difflib.NDiff(a, b)
	s := strings.Join(lines, "")
	if !strings.Contains(s, "- two\n") {
		t.Errorf("expected '- two\\n' in NDiff output:\n%s", s)
	}
	if !strings.Contains(s, "+ TWO\n") {
		t.Errorf("expected '+ TWO\\n' in NDiff output:\n%s", s)
	}
}

func TestRestore(t *testing.T) {
	a := difflib.SplitLines("one\ntwo\nthree\n")
	b := difflib.SplitLines("one\nTWO\nthree\n")
	delta := difflib.NDiff(a, b)

	restoredA := difflib.Restore(delta, 1)
	if difflib.JoinLines(restoredA) != difflib.JoinLines(a) {
		t.Errorf("Restore(1) mismatch: got %v, want %v", restoredA, a)
	}

	restoredB := difflib.Restore(delta, 2)
	if difflib.JoinLines(restoredB) != difflib.JoinLines(b) {
		t.Errorf("Restore(2) mismatch: got %v, want %v", restoredB, b)
	}
}

func TestContextDiff(t *testing.T) {
	a := difflib.SplitLines("one\ntwo\nthree\n")
	b := difflib.SplitLines("one\nTWO\nthree\n")
	lines := difflib.ContextDiff(difflib.DiffInput{
		A: a, B: b,
		FromFile: "orig", ToFile: "new",
		Context: 3,
	})
	if len(lines) == 0 {
		t.Error("expected non-empty context diff")
	}
	s := strings.Join(lines, "")
	if !strings.Contains(s, "! two\n") && !strings.Contains(s, "! TWO\n") {
		t.Errorf("expected '! ' markers in context diff:\n%s", s)
	}
}

func TestClosestMatch(t *testing.T) {
	best, ratio := difflib.ClosestMatch("appel", []string{"apple", "mango", "apply"})
	if best != "apple" && best != "apply" {
		t.Errorf("ClosestMatch: expected apple or apply, got %q (ratio %f)", best, ratio)
	}
	if ratio <= 0 || ratio > 1 {
		t.Errorf("ClosestMatch: ratio out of range: %f", ratio)
	}
}

func TestClosestMatchEmpty(t *testing.T) {
	best, ratio := difflib.ClosestMatch("foo", nil)
	if best != "" || ratio != 0 {
		t.Errorf("expected empty match for empty candidates, got %q %f", best, ratio)
	}
}

func TestClosestMatches(t *testing.T) {
	matches := difflib.ClosestMatches("appel", []string{"apple", "mango", "apply", "apt"}, 2)
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d: %v", len(matches), matches)
	}
}

func TestClosestMatchesN_Larger(t *testing.T) {
	matches := difflib.ClosestMatches("foo", []string{"foo", "bar"}, 10)
	if len(matches) != 2 {
		t.Errorf("expected 2 matches when n>len(candidates), got %d", len(matches))
	}
}

func TestApplyPatch_Identity(t *testing.T) {
	a := difflib.SplitLines("one\ntwo\nthree\n")
	b := difflib.SplitLines("one\nTWO\nthree\n")
	result := difflib.UnifiedDiff(difflib.DiffInput{
		A: a, B: b,
		FromFile: "a", ToFile: "b",
	})
	patched, err := difflib.ApplyPatch(a, result.String())
	if err != nil {
		t.Fatalf("ApplyPatch error: %v", err)
	}
	if difflib.JoinLines(patched) != difflib.JoinLines(b) {
		t.Errorf("ApplyPatch result mismatch: got %q, want %q",
			difflib.JoinLines(patched), difflib.JoinLines(b))
	}
}

func TestDiffResultIsEmpty(t *testing.T) {
	r := difflib.DiffResult{}
	if !r.IsEmpty() {
		t.Error("empty DiffResult should report IsEmpty=true")
	}
}

// Example for GoDoc
func ExampleUnifiedDiff() {
	a := difflib.SplitLines("one\ntwo\nthree\n")
	b := difflib.SplitLines("one\nTWO\nthree\n")
	result := difflib.UnifiedDiff(difflib.DiffInput{
		A:        a,
		B:        b,
		FromFile: "original",
		ToFile:   "modified",
		Context:  3,
	})
	_ = result.String()
}

func ExampleStringRatio() {
	ratio := difflib.StringRatio("kitten", "sitting")
	_ = ratio
}

func ExampleClosestMatch() {
	best, _ := difflib.ClosestMatch("appel", []string{"apple", "mango", "apply"})
	_ = best
}
