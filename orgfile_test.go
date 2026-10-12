package orgfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklasfasching/go-org/org"
)

func TestRenderStandardizesKeysAndCase(t *testing.T) {
	nodes, err := Parse("#+title: hi\n* head\n:properties:\n:id: x\n:end:\n", "t.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := Render(nodes)
	want := "#+TITLE: hi\n* head\n:PROPERTIES:\n:ID: x\n:END:\n"
	if got != want {
		t.Fatalf("render mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestRenderAlignsTagsToColumn80(t *testing.T) {
	nodes, err := Parse("* head :work:\n", "t.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := Render(nodes)
	// tag string ":work:" (6 chars) right-aligned so it ends at column 80,
	// i.e. starts at column 75 (0-indexed 74) -> 80 total line width.
	want := "* head" + strings.Repeat(" ", 80-len("* head")-len(":work:")) + ":work:\n"
	if got != want {
		t.Fatalf("tag alignment mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
	if len(got) != len(want) || len(strings.TrimRight(got, "\n")) != 80 {
		t.Fatalf("line width = %d, want 80", len(strings.TrimRight(got, "\n")))
	}
}

func TestWriteAtomicCreatesAndReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.org")

	if err := WriteAtomic(path, []byte("one\n")); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := WriteAtomic(path, []byte("two\n")); err != nil {
		t.Fatalf("replace: %v", err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "two\n" {
		t.Fatalf("content = %q, want %q", got, "two\n")
	}
	// no temp files left behind
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("dir has %d entries, want 1 (temp file leaked)", len(entries))
	}
}

func TestWriteAtomicPreservesMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.org")
	if err := os.WriteFile(path, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, []byte("y\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	fi, _ := os.Stat(path)
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", fi.Mode().Perm())
	}
}

func TestParseDefaultOnlyRecognizesTodoAndDone(t *testing.T) {
	// go-org's default spec is "TODO | DONE", so NEXT is title text, not Status.
	nodes, err := Parse("* NEXT ship it\n", "t.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	h := nodes[0].(org.Headline)
	if h.Status != "" {
		t.Fatalf("Status = %q, want empty (NEXT unrecognized by default)", h.Status)
	}
}

func TestParseWithTODORecognizesCustomKeywords(t *testing.T) {
	const spec = "TODO MAYBE UPCOMING NEXT WAITING PROG PROJECT | CANCELLED DONE"
	for _, kw := range []string{"TODO", "MAYBE", "UPCOMING", "NEXT", "WAITING", "PROG", "PROJECT", "CANCELLED", "DONE"} {
		nodes, err := ParseWithTODO("* "+kw+" ship it\n", "t.org", spec)
		if err != nil {
			t.Fatalf("parse %s: %v", kw, err)
		}
		h := nodes[0].(org.Headline)
		if h.Status != kw {
			t.Fatalf("Status = %q, want %q", h.Status, kw)
		}
		if got := org.String(h.Title...); got != "ship it" {
			t.Fatalf("Title = %q, want %q", got, "ship it")
		}
	}
}

func TestParseWithTODORoundTripsUnchanged(t *testing.T) {
	// Registering keywords must not alter rendered output: a keyword is written
	// back as Status instead of title text, producing the same bytes.
	const spec = "TODO NEXT PROG | DONE CANCELLED"
	src := "* PROG write the parser :work:\n* NEXT review it\n* DONE ship it\n"
	nodes, err := ParseWithTODO(src, "t.org", spec)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	withKw := Render(nodes)
	nodes, err = Parse(src, "t.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if withDefault := Render(nodes); withKw != withDefault {
		t.Fatalf("render diverges by TODO spec:\n--- keywords ---\n%q\n--- default ---\n%q", withKw, withDefault)
	}
}

func TestShiftHeadlineRelevelsDescendants(t *testing.T) {
	nodes, err := Parse("*** parent\n**** child\n***** grandchild\n", "t.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	parent := nodes[0].(org.Headline)

	// promote the level-3 root to level 1; descendants follow
	got := Render([]org.Node{ShiftHeadline(parent, 1-parent.Lvl)})
	want := "* parent\n** child\n*** grandchild\n"
	if got != want {
		t.Fatalf("promote mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}

	// demote by one
	got = Render([]org.Node{ShiftHeadline(parent, 1)})
	want = "**** parent\n***** child\n****** grandchild\n"
	if got != want {
		t.Fatalf("demote mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
