package internals

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type Tree struct {
	oid      []byte
	entries  map[string]Entry
	pathname string
	keys     []string
}

// lets define the methods and put it empty for now
// it will implement the Object interface and the Entry interface
func (t *Tree) New() {
	t.entries = make(map[string]Entry, 0)
}

func (t *Tree) AddEntry(ParentDirectories []string, entry Entry) {
	fmt.Println("adding entry with pathname = ", entry.GetPathname())
	fmt.Println("adding entry with name = ", entry.GetName())
	fmt.Println("current parent directories = ", ParentDirectories)
	fmt.Println("Current tree to add an entry to it = " + t.pathname + " and its name is = " + t.GetName())
	fmt.Println()
	if len(ParentDirectories) == 0 {
		fmt.Println("Current entry is a blob")
		// fmt.Println("saving... ", entry.GetName())
		t.entries[entry.GetName()] = entry
		t.keys = append(t.keys, entry.GetName())
		fmt.Println("Saving blob with name = "+entry.GetName(), " in tree called = "+t.GetName())
		fmt.Println("current entries saved to the current tree: ")

		for _, key := range t.keys {
			fmt.Println(t.entries[key].GetPathname() + " with name " + t.entries[key].GetName())
		}
	} else {
		fmt.Println("Current entry is a Tree")
		childTree := &Tree{}
		childTree.New()
		val, ok := t.entries[filepath.Base(ParentDirectories[0])]
		if ok {
			fmt.Println("Tree with the name = " + filepath.Base(ParentDirectories[0]) + " is already saved")
			fmt.Println("fetching the tree to add an entry to it")
			childTree = val.(*Tree)
		} else {
			fmt.Println("Tree with the name = " + filepath.Base(ParentDirectories[0]) + " is not saved before")

		}

		fmt.Println("we will pass it parent directories = ", ParentDirectories[1:])
		childTree.SetTreePathname(ParentDirectories[0])
		childTree.AddEntry(ParentDirectories[1:], entry)

		fmt.Println("Child tree fullname = " + childTree.GetPathname())
		fmt.Println("Child tree name = " + childTree.GetName())
		fmt.Println("Child tree entries names: ")
		for _, key := range childTree.keys {
			fmt.Println("child entry")
			fmt.Println(childTree.entries[key].GetPathname() + " with name " + childTree.entries[key].GetName())
		}
		// fmt.Println("saving... ", childTree.GetName())
		t.entries[childTree.GetName()] = childTree
		if !ok {
			t.keys = append(t.keys, childTree.GetName())
		}

		fmt.Println("child tree has beed added as an entry, lets loop over all entries: ")
		for _, key := range t.keys {
			fmt.Println(t.entries[key].GetPathname() + " with name " + t.entries[key].GetName())
		}
	}
	fmt.Println("******************************")
}

func (t *Tree) SetTreePathname(pathname string) {
	t.pathname = pathname
}
func (t *Tree) GetPathname() string {
	return t.pathname
}

func (t *Tree) Build(entries []Entry) Tree {
	slices.SortFunc(entries, func(e1, e2 Entry) int {
		if e1.GetPathname() < e2.GetPathname() {
			return -1
		} else if e1.GetPathname() > e2.GetPathname() {
			return 1
		}
		return 0
	})

	root := Tree{}
	root.New()
	for _, entry := range entries {
		// fmt.Println("from tree.Build() => ", entry.ParentDirectories())
		root.AddEntry(entry.ParentDirectories(), entry)
	}

	fmt.Println("One more check")
	fmt.Println()
	print(&root, "")
	fmt.Println()
	return root
}

func print(tree *Tree, prefixSpace string) {
	fmt.Println(prefixSpace + "tree pathname = " + tree.pathname)
	for _, k := range tree.keys {
		entry := tree.entries[k]
		if entry.Type() == "tree" {
			fmt.Println("Cur entry is a tree with name = ", entry.GetName())
			childTree := entry.(*Tree)
			print(childTree, prefixSpace+"  ")
		} else {
			fmt.Println(prefixSpace + " " + entry.GetPathname())
		}
	}
}
func (t *Tree) Traverse(fn func(Entry)) {
	for _, k := range t.keys {
		entry := t.entries[k]
		if entry.Type() == "tree" {
			childTree := entry.(*Tree)
			childTree.Traverse(fn)
		} else {
			fn(entry)
		}
	}
	fn(t)
}
func (t *Tree) Type() string {
	return "tree"
}
func (t *Tree) GetOid() []byte {
	return t.oid
}

func (t *Tree) SetOid(oid []byte) {
	t.oid = oid
}

func (t *Tree) ToString() string {
	var data []byte

	for _, k := range t.keys {
		curEntry := t.entries[k]
		data = append(data, fmt.Sprintf("%s %s", curEntry.GetMode(), curEntry.GetName())...)
		data = append(data, 0x00)
		data = append(data, curEntry.GetOid()...)
	}
	// tree 74100644 file2.txt0�k}fu���v�l�ȿ�7j!100644 file3.txt0�e������ƮHH��H

	// why there is a zero appended after the file name ? because in git the format of the tree object is like this:
	// <mode> <name>\0<oid><mode> <name>\0<oid>...

	// to remove that zero we can use the following code:
	// data = bytes.ReplaceAll(data, []byte{0x00}, []byte{})

	// or from the beginning, we can use the following code to append the data without the zero:
	// data = append(data, fmt.Sprintf("%s %s", curEntry.GetMode(), curEntry.GetName())...)
	// data = append(data, 0x00)
	// data = append(data, curEntry.GetOid()...)
	return string(data)
}

func (t *Tree) GetName() string {
	return filepath.Base(t.pathname)
}

func (t *Tree) GetMode() string {
	return "40000"
}

func (t *Tree) ParentDirectories() []string {
	prefixs := strings.Split(filepath.ToSlash(t.GetPathname()), "/")
	parents := []string{}

	for i := 1; i < len(prefixs); i++ {
		parents = append(parents, strings.Join(prefixs[:i], "/"))
	}

	return parents
}
