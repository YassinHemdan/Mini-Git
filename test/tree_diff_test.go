package test

import (
	"JIT/commands"
	database "JIT/internals/database"
	"testing"
)

const (
	CREATED  = "created"
	DELETED  = "deleted"
	MODIFIED = "modified"
)

type file struct {
	path    string
	content string
}

type TestEntry struct {
	before          []func()
	after           []func()
	expectedResults map[string]string
}

func getTestEntry(t *testing.T, helper *commands.CommandHelper) TestEntry {
	t.Helper()
	before := make([]func(), 0)
	after := make([]func(), 0)

	before = append(before, (func() func() {
		file := file{path: "README.txt", content: "this is a readme file"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}

	})())
	before = append(before, (func() func() {
		file := file{path: "TODO.txt", content: "this is todo file"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}
	})())

	before = append(before, (func() func() {
		file := file{path: "config/database/development.txt", content: "dev file"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}
	})())

	before = append(before, (func() func() {
		file := file{path: "config/database/production.txt", content: "prod file"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}
	})())

	before = append(before, (func() func() {
		file := file{path: "lib/app.txt", content: "app file"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}
	})())

	after = append(after, (func() func() {
		file := file{path: "lib/app.txt", content: "app file v2"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}
	})())

	after = append(after, (func() func() {
		file := file{path: "lib/models/repository.txt", content: "repo file"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}
	})())

	after = append(after, (func() func() {
		file := file{path: "lib/models/user.txt", content: "user file"}
		return func() {
			helper.WriteFile(t, file.path, file.content)
		}
	})())

	after = append(after, (func() func() {
		file := file{path: "TODO.txt"}
		return func() {
			helper.Delete(t, file.path)
		}
	})())

	expected := make(map[string]string)

	expected["lib/models/user.txt"] = CREATED
	expected["lib/models/repository.txt"] = CREATED
	expected["lib/app.txt"] = MODIFIED
	expected["TODO.txt"] = DELETED

	return TestEntry{before: before, after: after, expectedResults: expected}
}

func Test_TreeDiff(t *testing.T) {
	t.Run("Comparing OIDS", func(t *testing.T) {
		helper := commands.NewCommandHelper(t)
		repo := helper.Repo(t)
		testEntry := getTestEntry(t, helper)

		for _, fn := range testEntry.before {
			fn()
		}

		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		for _, fn := range testEntry.after {
			fn()
		}

		helper.Delete(t, ".jit/index") // to make sure deleted files got removed from the index
		helper.JitCommand("add", ".")
		helper.Commit(t, "second second")

		oid, err := repo.Refs().ReadHead()
		if err != nil {
			t.Fatalf("Could not read HEAD - %v\n", err)
		}

		commitObj, err := repo.Database().Load(oid)
		if err != nil || commitObj.Type() != "commit" {
			t.Fatalf("Could not load commit obj - %v\n", err)
		}

		commit := commitObj.(*database.Commit)

		changes, err := repo.Database().TreeDiff(commit.GetParentOid(), commit.GetOid())
		if err != nil {
			t.Fatalf("Could not compare two oids - %v\n", err)
		}

		if len(changes) != len(testEntry.expectedResults) {
			t.Fatalf("expected %d entries but found %d\n", len(testEntry.expectedResults), len(changes))
		}

		for expectedName, expectedStatus := range testEntry.expectedResults {
			entries, found := changes[expectedName]
			if !found {
				t.Fatalf("path %s not found\n", expectedName)
			}

			switch expectedStatus {
			case MODIFIED:
				checkModifiedEntries(t, entries)
			case DELETED:
				checkDeletedFiles(t, entries)
			case CREATED:
				checkAddedFiles(t, entries)
			}
		}
	})
}

func checkModifiedEntries(t *testing.T, entries []database.Entry) {
	t.Helper()

	// both before and after should exist with either different oids or modes
	if len(entries) != 2 {
		t.Fatalf("expeted 2 entries but found %d\n", len(entries))
	}

	if entries[0] == nil || entries[1] == nil {
		t.Fatal("nil entry was not expected\n")
	}
	if entries[0].GetMode() == entries[1].GetMode() && string(entries[0].GetOid()) == string(entries[1].GetOid()) {
		t.Fatal("expected different mode or oids\n")
	}
}
func checkDeletedFiles(t *testing.T, entries []database.Entry) {
	t.Helper()
	if entries[0] == nil {
		t.Fatal("first entry should not be nil\n")
	}
	if entries[1] != nil {
		t.Fatal("second entry should be nil\n")
	}
}
func checkAddedFiles(t *testing.T, entries []database.Entry) {
	t.Helper()
	if entries[0] != nil {
		t.Fatal("first entry should be nil\n")
	}
	if entries[1] == nil {
		t.Fatal("second entry should not be nil\n")
	}
}
