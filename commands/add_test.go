package commands

import (
	"testing"
)

func TestAdd_RegularFile(t *testing.T) {
	helper := NewCommandHelper(t)
	helper.WriteFile(t, "file1.txt", "Hello from file1.txt\n")
	helper.JitCommand("add", "file1.txt")

	assertIndex(t, helper, [][]any{
		{"100644", "file1.txt"},
	})
}

func TestAdd_ExecutableFile(t *testing.T) {
	helper := NewCommandHelper(t)
	helper.WriteFile(t, "file1.txt", "Hello from file1.txt\n")
	helper.MakeExecutable(t, "file1.txt")
	helper.JitCommand("add", "file1.txt")

	assertIndex(t, helper, [][]any{
		{"100755", "file1.txt"},
	})
}

func assertIndex(t *testing.T, helper *CommandHelper, expected [][]any) {
	t.Helper()
	repo := helper.Repo(t)
	if _, err := repo.Index().Load(); err != nil {
		t.Fatalf("Could not load index - %v", err)
	}

	entries := repo.Index().GetEntries()

	if len(entries) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(entries))
	}

	for i, entry := range entries {
		expectedMode := expected[i][0].(string)
		expectedPath := expected[i][1].(string)
		if entry.GetMode() != expectedMode {
			t.Errorf("entry %d: expected mode %s, got %s", i, expectedMode, entry.GetMode())
		}
		if entry.GetName() != expectedPath {
			t.Errorf("entry %d: expected path %s, got %s", i, expectedPath, entry.GetName())
		}
	}
}
