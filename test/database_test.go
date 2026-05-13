package test

import (
	"JIT/commands"
	"JIT/internals"
	database "JIT/internals/database"
	"testing"
)

func TestDatabase_LoadCommitObject(t *testing.T) {
	t.Run("LoadingCommitObject", func(t *testing.T) {
		helper := commands.NewCommandHelper(t)
		commitMessage := "First Commit"
		helper.WriteFile(t, "file.txt", "Hello from file.txt")
		helper.JitCommand("add", ".")
		helper.Commit(t, commitMessage)

		repo := helper.Repo(t)
		commitOid, err := repo.Refs().ReadHead()
		if err != nil {
			t.Fatalf("Coult not read commit's oid - %v", err)
		}

		object, err := repo.Database().Load(commitOid)

		if err != nil {
			t.Fatalf("Coult not load object from database - %v", err)
		}

		if object.Type() != "commit" {
			t.Fatalf("Coult not load commit from database - %v", err)
		}
	})

	t.Run("LoadingCommitObjectWithParent", func(t *testing.T) {
		helper := commands.NewCommandHelper(t)
		commitMessage := "First Commit"
		helper.WriteFile(t, "file.txt", "Hello from file.txt")
		helper.JitCommand("add", ".")
		helper.Commit(t, commitMessage)

		repo := helper.Repo(t)
		parentCommitOid, err := repo.Refs().ReadHead()
		if err != nil {
			t.Fatalf("Coult not read commit's oid - %v", err)
		}

		helper.WriteFile(t, "file2.txt", "Hello from file2.txt")
		helper.JitCommand("add", ".")
		helper.Commit(t, commitMessage+"v2")

		latestCommitOid, err := repo.Refs().ReadHead()
		if err != nil {
			t.Fatalf("Coult not read commit's oid - %v", err)
		}

		object, err := repo.Database().Load(latestCommitOid)
		if err != nil {
			t.Fatalf("Coult not load object from database - %v", err)
		}

		if object.Type() != "commit" {
			t.Fatalf("Coult not load commit object from database - %v", err)
		}

		commit := object.(*database.Commit)

		if len(string(commit.GetParentOid())) == 0 || string(commit.GetParentOid()) != string(parentCommitOid) {
			t.Fatalf("Coult not load commit object from database - %v", err)
		}
	})
}

func TestDatabase_LoadTreeObject(t *testing.T) {
	t.Run("LoadingTreeObjectWithOneBlob", func(t *testing.T) {
		helper := commands.NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "Hello from file.txt")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")
		repo := helper.Repo(t)
		commitOid, err := repo.Refs().ReadHead()
		if err != nil {
			t.Fatalf("Couldn't read refs - %v\n", err)
		}

		commitObject, err := repo.Database().Load(commitOid)
		if err != nil {
			t.Fatalf("Couldn't load commit object - %v\n", err)
		}

		commit := commitObject.(*database.Commit)
		treeOid := commit.GetTreeOid()

		entries := loadTree(t, repo, treeOid)
		for _, entry := range entries {
			blobObject, err := repo.Database().Load(entry.GetOid())
			if err != nil {
				t.Fatalf("Couldn't load blob object - %v\n", err)
			}
			if blobObject.Type() != "blob" {
				t.Fatalf("Couldn't load tree object - %v\n", err)
			}

			blob := blobObject.(*database.Blob)

			if string(blob.GetData()) != "Hello from file.txt" {
				t.Fatalf("Expected content 'Hello from file.txt' but found '%s'\n", string(blob.GetData()))
			}
		}
	})
	t.Run("LoadingTreeObjectWithMultipleBlobs", func(t *testing.T) {
		helper := commands.NewCommandHelper(t)
		fileContentMap := make(map[string]string)
		fileContentMap["file2.txt"] = "Hello from file2.txt"
		fileContentMap["file3.txt"] = "Hello from file3.txt"
		fileContentMap["file1.txt"] = "Hello from file1.txt"

		helper.WriteFile(t, "file2.txt", fileContentMap["file2.txt"])
		helper.WriteFile(t, "file3.txt", fileContentMap["file3.txt"])
		helper.WriteFile(t, "file1.txt", fileContentMap["file1.txt"])
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")
		repo := helper.Repo(t)
		commitOid, err := repo.Refs().ReadHead()
		if err != nil {
			t.Fatalf("Couldn't read refs - %v\n", err)
		}

		commitObject, err := repo.Database().Load(commitOid)
		if err != nil {
			t.Fatalf("Couldn't load commit object - %v\n", err)
		}

		commit := commitObject.(*database.Commit)
		treeOid := commit.GetTreeOid()

		entries := loadTree(t, repo, treeOid)

		if len(entries) != len(fileContentMap) {
			t.Fatalf("Expected %d entries but found %d\n", len(fileContentMap), len(entries))
		}
		for _, entry := range entries {
			blobObject, err := repo.Database().Load(entry.GetOid())
			if err != nil {
				t.Fatalf("Couldn't load blob object - %v\n", err)
			}
			if blobObject.Type() != "blob" {
				t.Fatalf("Couldn't load tree object - %v\n", err)
			}

			blob := blobObject.(*database.Blob)

			if string(blob.GetData()) != fileContentMap[entry.GetName()] {
				t.Fatalf("Expected content 'Hello from file.txt' but found '%s'\n", string(blob.GetData()))
			}
		}
	})
	t.Run("LoadingNestedTreeObject", func(t *testing.T) {
		helper := commands.NewCommandHelper(t)

		fileContentMap := make(map[string]string)
		fileContentMap["file2.txt"] = "Hello from file2.txt\n"
		fileContentMap["file3.txt"] = "Hello from file3.txt"
		fileContentMap["file1.txt"] = "Hello from file1.txt\n"

		helper.WriteFile(t, "dir/file1.txt", fileContentMap["file1.txt"])
		helper.WriteFile(t, "dir/file2.txt", fileContentMap["file2.txt"])
		helper.WriteFile(t, "dir/file3.txt", fileContentMap["file3.txt"])
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")
		repo := helper.Repo(t)
		commitOid, err := repo.Refs().ReadHead()
		if err != nil {
			t.Fatalf("Couldn't read refs - %v\n", err)
		}

		commitObject, err := repo.Database().Load(commitOid)
		if err != nil {
			t.Fatalf("Couldn't load commit object - %v\n", err)
		}

		commit := commitObject.(*database.Commit)
		treeOid := commit.GetTreeOid()

		entries := loadTree(t, repo, treeOid)
		var traverse func(entries map[string]database.Entry)

		traverse = func(entries map[string]database.Entry) {
			for _, entry := range entries {
				if entry.Type() == "tree" {
					subTreeEntries := loadTree(t, repo, entry.GetOid())
					traverse(subTreeEntries)
				} else if entry.Type() == "blob" {
					objectBlob, err := repo.Database().Load(entry.GetOid())
					if err != nil {
						t.Fatalf("Could not load blob - %v\n", err)
					}

					if objectBlob.Type() != "blob" {
						t.Fatalf("Expected type blob but found '%s'\n", objectBlob.Type())
					}

					blob := objectBlob.(*database.Blob)
					if string(blob.GetData()) != fileContentMap[entry.GetName()] {
						t.Fatalf("Expected content '%s' but found '%s'\n", fileContentMap[entry.GetName()], string(blob.GetData()))
					}
				}
			}
		}
		// fmt.Println(entries)
		traverse(entries)
	})
}

func loadTree(t *testing.T, repo *internals.Repository, oid []byte) map[string]database.Entry {
	t.Helper()
	treeObject, err := repo.Database().Load(oid)
	if err != nil {
		t.Fatalf("Couldn't load tree object - %v\n", err)
	}

	if treeObject.Type() != "tree" {
		t.Fatalf("Couldn't load tree object - %v\n", err)
	}

	tree := treeObject.(*database.Tree)
	return tree.GetEntries()

}
