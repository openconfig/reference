// Package pathtree provides a tree structure for storing Collector updates. In
// this tree, all terminal leaves have a value, and no intermediate branches may
// have a value (they will be assigned a sub-branch according to ValKey if
// supplied).
package pathtree

import "fmt"

// ValKey is the map key used when a value is assigned to an intermediate node
// as Branch only supports a single value, either a terminal value or a branch.
// It is expected that intermediate nodes will generally not have values but
// when they do, this will ensure they will be representable even when converted
// to JSON.
var ValKey = "_VALUE"

// Path wraps a slice of strings representing a path.
type Path []string

// A Branch stores a tree with branches for internal nodes and values only in
// its leaf nodes. This allows a Branch to store results by Collector path.
type Branch map[string]interface{}

// Add modifies the tree to contain all branches in the path to the terminal
// leaf containing val.
func (b Branch) Add(path Path, val interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path for %v", val)
	}
	br := b
	for len(path) > 1 {
		tmp := br[path[0]]
		var b Branch
		if tmp == nil {
			b = Branch{}
			br[path[0]] = b
		} else {
			var ok bool
			b, ok = tmp.(Branch)
			if !ok {
				// Attempt to overwrite a previous terminal leaf.
				b = Branch{}
				br[path[0]] = b
				br[ValKey] = tmp
			}
		}
		br = b
		path = path[1:]
	}
	tmp := br[path[0]]
	if b, ok := tmp.(Branch); ok {
		b[ValKey] = val
	} else {
		br[path[0]] = val
	}
	return nil
}

// Get returns a subtree for intermediate paths, a terminal value for complete
// paths, or nil if the path is not in the tree.
func (b Branch) Get(path Path) interface{} {
	var node interface{} = b
	for _, p := range path {
		br, _ := node.(Branch)
		if br == nil {
			return nil
		}
		node = br[p]
	}
	return node
}

// Delete removes path from the tree if it exists, and returns true if so. It
// does not clean up any empty parent nodes in path.
func (b Branch) Delete(path Path) bool {
	switch len(path) {
	case 0:
		return false
	case 1:
		if _, ok := b[path[0]]; !ok {
			return false
		}
		delete(b, path[0])
		return true
	default:
		if br, ok := b[path[0]].(Branch); ok && br != nil {
			return br.Delete(path[1:])
		}
		return false
	}
}

// PathVal is a container for a fully walked path and a terminal value.
type PathVal struct {
	Path Path
	Val  interface{}
}

// Walk will do a full walk of the tree and return a slice of path-value pairs.
// Order of the outputs is not guaranteed and may differ between calls on the
// same b.
func (b Branch) Walk() []PathVal {
	ret := []PathVal{}
	for p, v := range b {
		if br, ok := v.(Branch); ok && br != nil {
			for _, pv := range br.Walk() {
				pv.Path = append(Path{p}, pv.Path...)
				ret = append(ret, pv)
			}
		} else {
			ret = append(ret, PathVal{
				Path: Path{p},
				Val:  v,
			})
		}
	}
	return ret
}

// Less returns true if p sorts before p2.
func (p Path) Less(p2 Path) bool {
	for x := 0; x < len(p) && x < len(p2); x++ {
		if p[x] < p2[x] {
			return true
		}
		if p[x] > p2[x] {
			return false
		}
	}
	return len(p) < len(p2)
}

// Equal returns true if p is equivalent to p2.
func (p Path) Equal(p2 Path) bool {
	if len(p) != len(p2) {
		return false
	}
	for x := 0; x < len(p); x++ {
		if p[x] != p2[x] {
			return false
		}
	}
	return true
}

// Paths is a slice of paths that has a sorting implementation.
type Paths []Path

// Sorting interface for Paths.
func (p Paths) Len() int      { return len(p) }
func (p Paths) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Paths) Less(i, j int) bool {
	return p[i].Less(p[j])
}

// PathVals is a slice of PathVal that has a sorting implementation over paths.
type PathVals []PathVal

// Sorting interface for PathVals.
func (p PathVals) Len() int      { return len(p) }
func (p PathVals) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PathVals) Less(i, j int) bool {
	return Path(p[i].Path).Less(p[j].Path)
}
