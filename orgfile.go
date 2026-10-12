// Package orgfile provides generic org-mode file primitives shared by the
// org tools: parsing, pretty-printing via go-org's OrgWriter, subtree
// re-leveling, and atomic file replacement.
package orgfile

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
)

// Parse parses org source with go-org's default settings and returns its
// top-level nodes. go-org's default TODO spec is "TODO | DONE", so any other
// keyword lands in the headline title rather than its Status; use
// ParseWithTODO when the caller needs a wider keyword set recognized.
func Parse(source, path string) ([]org.Node, error) {
	return ParseWithTODO(source, path, "")
}

// ParseWithTODO parses org source with todoSpec registered as the default
// #+TODO: setting, so go-org populates Headline.Status for those keywords.
// todoSpec uses org's own syntax, active and done states separated by a pipe:
// "TODO NEXT | DONE CANCELLED". An empty todoSpec leaves go-org's default in
// place. A file carrying its own #+TODO: line still overrides this.
func ParseWithTODO(source, path, todoSpec string) ([]org.Node, error) {
	c := org.New().Silent()
	if todoSpec != "" {
		c.DefaultSettings["TODO"] = todoSpec
	}
	d := c.Parse(strings.NewReader(source), path)
	if d.Error != nil {
		return nil, d.Error
	}
	return d.Nodes, nil
}

// ShiftHeadline re-levels a headline and all its descendant headlines by delta.
func ShiftHeadline(h org.Headline, delta int) org.Headline {
	h.Lvl += delta
	children := make([]org.Node, len(h.Children))
	for i, c := range h.Children {
		if ch, ok := c.(org.Headline); ok {
			children[i] = ShiftHeadline(ch, delta)
		} else {
			children[i] = c
		}
	}
	h.Children = children
	return h
}

// TagsColumn is the column headline tags are right-aligned to (go-org defaults
// to 77).
const TagsColumn = 80

// Render pretty-prints nodes as a standardized org-mode string using go-org's
// OrgWriter defaults, except tags are aligned to TagsColumn.
func Render(nodes []org.Node) string {
	w := org.NewOrgWriter()
	w.TagsColumn = TagsColumn
	org.WriteNodes(w, nodes...)
	return w.String()
}

// WriteAtomic writes data to path via a temp file in the same directory then
// renames over the target, preserving mode when the file already exists.
func WriteAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode()
	}
	tmp, err := os.CreateTemp(dir, ".orgfile-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
