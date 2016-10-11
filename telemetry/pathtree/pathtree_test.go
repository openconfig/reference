package pathtree

import (
	"reflect"
	"sort"
	"testing"
)

func TestEmptyPath(t *testing.T) {
	tree := Branch{}
	err := tree.Add(nil, "")
	if err == nil {
		t.Error("tree.Add(...): want error on empty path")
	}
}

var (
	ifPath  = Path{"interfaces", "interface[name=Ethernet3/1/1]"}
	ifPath1 = Path{"interfaces", "interface[name=Ethernet3/1/2]"}
	ifPath2 = Path{"interfaces", "interface[name=Ethernet3/2/1]"}
	bgpPath = Path{"protocols", "protocol[name=bgp]", "neighbors", "neighbor[neighbor-address=1.1.1.1]"}
)

func TestMultiInsert(t *testing.T) {
	tree := Branch{}
	want := 2
	tree.Add(ifPath, 1)
	tree.Add(ifPath, want)
	if got := tree.Get(ifPath); got != want {
		t.Errorf("tree.Get(%v): got %v, want %d", ifPath, got, want)
	}
}

func TestIntermediate(t *testing.T) {
	tree := Branch{}
	partial := Path{"interfaces"}
	tree.Add(ifPath, 1)
	got := tree.Get(partial)
	want := Branch{"interface[name=Ethernet3/1/1]": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tree.Get(%v): got %v, want %v", partial, got, want)
	}
}

func TestMultipleIntermediate(t *testing.T) {
	tree := Branch{}
	partial := Path{"interfaces"}
	tree.Add(ifPath, 1)
	tree.Add(ifPath1, 2)
	s := Branch{}
	s.Add(Path{"name"}, "vlan subinterface")
	sInt := Branch{}
	sInt.Add(Path{"subinterface[name=100]"}, s)
	sContainer := Branch{}
	sContainer.Add(Path{"subinterfaces"}, sInt)
	tree.Add(ifPath2, sContainer)
	got := tree.Get(partial)
	want := Branch{
		"interface[name=Ethernet3/1/1]": 1,
		"interface[name=Ethernet3/1/2]": 2,
		"interface[name=Ethernet3/2/1]": Branch{
			"subinterfaces": Branch{
				"subinterface[name=100]": Branch{
					"name": "vlan subinterface",
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tree.Get(%v): got %v, want %v", partial, got, want)
	}
}

var branchTest = []PathVal{
	{
		Path: ifPath,
		Val:  "a",
	}, {
		Path: bgpPath,
		Val:  "b",
	},
}

func TestBranch(t *testing.T) {
	tree := Branch{}
	for _, tt := range branchTest {
		tree.Add(tt.Path, tt.Val)
	}
	for _, tt := range branchTest {
		if got := tree.Get(tt.Path); got != tt.Val {
			t.Errorf("tree.Get(%v): got %v, want %q", tt.Path, got, tt.Val)
		}
	}
}

var sortPathValsTest = PathVals{
	{
		Path: Path{"a", "g"},
		Val:  "h",
	}, {
		Path: Path{"a", "b", "e"},
		Val:  "f",
	}, {
		Path: Path{"i"},
		Val:  "j",
	}, {
		Path: Path{"a", "b", "c"},
		Val:  "d",
	},
}

var sortPathValsWant = PathVals{
	{
		Path: Path{"a", "b", "c"},
		Val:  "d",
	}, {
		Path: Path{"a", "b", "e"},
		Val:  "f",
	}, {
		Path: Path{"a", "g"},
		Val:  "h",
	}, {
		Path: Path{"i"},
		Val:  "j",
	},
}

func TestSortPathVals(t *testing.T) {
	got := sortPathValsTest
	sort.Sort(got)
	want := sortPathValsWant
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sorting PathVals got: %v, want %v", got, want)
	}
}

func TestWalk(t *testing.T) {
	tree := Branch{}
	for _, tt := range sortPathValsTest {
		tree.Add(tt.Path, tt.Val)
	}
	got := PathVals(tree.Walk())
	// Sorting needed because walk does not guarantee order.
	sort.Sort(got)
	want := sortPathValsWant
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Walk got: %v, want %v", got, want)
	}
}

func TestDelete(t *testing.T) {
	tree := Branch{}
	for _, tt := range sortPathValsTest {
		tree.Add(tt.Path, tt.Val)
	}
	if tree.Delete(nil) {
		t.Error("tree.Delete(nil) got true, want false")
	}
	if tree.Delete(Path{"foo"}) {
		t.Error(`tree.Delete("foo") got true, want false`)
	}
	if !tree.Delete(sortPathValsWant[0].Path) {
		t.Errorf("tree.Delete(%q) got false, want true", sortPathValsWant[0].Path)
	}
	got := PathVals(tree.Walk())
	// Sorting needed because walk does not guarantee order.
	sort.Sort(got)
	want := sortPathValsWant[1:]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("After Delete(%q) got: %v, want %v", sortPathValsWant[0], got, want)
	}
	if !tree.Delete(sortPathValsWant[3].Path) {
		t.Errorf("tree.Delete(%q) got false, want true", sortPathValsWant[3].Path)
	}
	got = PathVals(tree.Walk())
	sort.Sort(got)
	want = want[:2]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("After Delete(%q) got: %v, want %v", sortPathValsWant[3], got, want)
	}
	if !tree.Delete(Path{"a"}) {
		t.Error(`tree.Delete("a") got false, want true`)
	}
	got = PathVals(tree.Walk())
	want = PathVals{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("After Delete(%q) got: %v, want %v", "a", got, want)
	}
}

var sortPathsTest = Paths{
	{"x"},
	{"a", "d"},
	{"a", "b", "x"},
	{"A"},
}

var sortPathsWant = Paths{
	{"A"},
	{"a", "b", "x"},
	{"a", "d"},
	{"x"},
}

func TestSortPaths(t *testing.T) {
	got := sortPathsTest
	sort.Sort(got)
	want := sortPathsWant
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sorting Paths got: %v, want %v", got, want)
	}
}

func TestPathLess(t *testing.T) {
	paths := sortPathsWant
	for x := 0; x < len(paths)-1; x++ {
		if !paths[x].Less(paths[x+1]) {
			t.Errorf("(%v).Less(%v): got false, want true", paths[x], paths[x+1])
		}
	}
	for x := len(paths) - 1; x > 0; x-- {
		if paths[x].Less(paths[x-1]) {
			t.Errorf("(%v).Less(%v): got true, want false", paths[x], paths[x-1])
		}
	}
	for _, p := range paths {
		if p.Less(p) {
			t.Errorf("(%v).Less(%v): got true, want false", p, p)
		}
	}
}

func TestPathEqual(t *testing.T) {
	paths := sortPathsWant
	for x := 0; x < len(paths)-1; x++ {
		if paths[x].Equal(paths[x+1]) {
			t.Errorf("(%v).Equal(%v): got true, want false", paths[x], paths[x+1])
		}
	}
	for x := len(paths) - 1; x > 0; x-- {
		if paths[x].Equal(paths[x-1]) {
			t.Errorf("(%v).Equal(%v): got true, want false", paths[x], paths[x-1])
		}
	}
	for _, p := range paths {
		if !p.Equal(p) {
			t.Errorf("(%v).Equal(%v): got false, want true", p, p)
		}
	}
}
