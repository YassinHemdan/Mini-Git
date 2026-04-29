//go:build ignore

package main

import (
	"JIT/internals"
	database "JIT/internals/database"
	"JIT/internals/utils"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	ShowHead()
}
func ShowHead() {
	repo, err := internals.NewRepository(utils.JitMetadataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize new repository - %v\n", err)
		os.Exit(1)
	}

	commitOid, err := repo.Refs().ReadHead()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read Refs - %v\n", err)
		os.Exit(1)
	}

	obj, err := repo.Database().Load(commitOid)
	if err != nil || obj.Type() != "commit" {
		fmt.Fprintf(os.Stderr, "Could not load commit object from DB - %v\n", err)
		os.Exit(1)
	}

	var showTree func(entries map[string]database.Entry, pathname string)
	showTree = func(entries map[string]database.Entry, pathname string) {
		for _, entry := range entries {
			fullpathname := filepath.Join(pathname, entry.GetName())
			if entry.Type() == "tree" {
				if entry.GetName() == "index" {
					// fmt.Printf("oiddd = %x\n", entry.GetOid())
				}
				// fmt.Println("Current tree name =", entry.GetName())
				// fmt.Printf("Current tree oid = %x\n", entry.GetOid())
				showTree(loadTree(entry.GetOid(), repo), fullpathname)
			} else if entry.Type() == "blob" {
				fmt.Printf("%s %x %s\n", entry.GetMode(), entry.GetOid(), fullpathname)
			}
		}
	}
	commitObj := obj.(*database.Commit)
	showTree(loadTree(commitObj.GetTreeOid(), repo), "")
}

func loadTree(oid []byte, repo *internals.Repository) map[string]database.Entry {
	obj, err := repo.Database().Load(oid)
	if err != nil || obj.Type() != "tree" {
		fmt.Fprintf(os.Stderr, "Could not load tree object from DB - %v\n", err)
		os.Exit(1)
	}

	treeObj := obj.(*database.Tree)
	return treeObj.GetEntries()
}
