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
func TestAdd_MultipleFiles(t *testing.T) {
	helper := NewCommandHelper(t)

	helper.WriteFile(t, "hello.txt", "Hello from hello.txt\n")
	helper.WriteFile(t, "world.txt", "Hello from world.txt\n")

	helper.JitCommand("add", "hello.txt", "world.txt")

	assertIndex(t, helper, [][]any{
		{"100644", "hello.txt"},
		{"100644", "world.txt"},
	})
}

func TestAdd_AddIncrementalFiles(t *testing.T) {
	helper := NewCommandHelper(t)

	helper.WriteFile(t, "world.txt", "Hello from world.txt\n")
	helper.JitCommand("add", "world.txt")
	assertIndex(t, helper, [][]any{
		{"100644", "world.txt"},
	})

	helper.WriteFile(t, "hello.txt", "Hello from hello.txt\n")
	helper.JitCommand("add", "hello.txt")
	assertIndex(t, helper, [][]any{
		{"100644", "hello.txt"},
		{"100644", "world.txt"},
	})
}

func TestAdd_AddDirectory(t *testing.T) {
	helper := NewCommandHelper(t)
	helper.WriteFile(t, "a-dir/nested.txt", "content")
	helper.JitCommand("add", "a-dir")
	assertIndex(t, helper, [][]any{
		{"100644", "a-dir/nested.txt"},
	})

}
func TestAdd_AddRoot(t *testing.T) {
	helper := NewCommandHelper(t)
	helper.WriteFile(t, "a-dir/nested1.txt", "content")
	helper.WriteFile(t, "a-dir/nested2.txt", "content")
	helper.WriteFile(t, "outer.txt", "content")
	helper.JitCommand("add", ".")
	assertIndex(t, helper, [][]any{
		{"100644", "a-dir/nested1.txt"},
		{"100644", "a-dir/nested2.txt"},
		{"100644", "outer.txt"},
	})
}

func TestAdd_SilentOnSuccess(t *testing.T) {
	helper := NewCommandHelper(t)
	helper.WriteFile(t, "hello.txt", "hello")
	helper.JitCommand("add", "hello.txt")

	helper.AssertStatus(t, 0)
	helper.AssertStdout(t, "")
	helper.AssertStderr(t, "")
}

func TestAdd_FailsForNonExistentFiles(t *testing.T) {
	helper := NewCommandHelper(t)
	helper.JitCommand("add", "no-such-file")

	helper.AssertStatus(t, 128)
	helper.AssertStderr(t, "fatal: pathspec 'no-such-file' did not match any files\n")

	assertIndex(t, helper, [][]any{})
}

func TestAdd_FailsForUnreadableFiles(t *testing.T) {
	helper := NewCommandHelper(t)

	helper.WriteFile(t, "secret.txt", "")
	helper.MakeUnreadable(t, "secret.txt")

	helper.JitCommand("add", "secret.txt")

	helper.AssertStatus(t, 128)
	helper.AssertStderr(t, "error: open('secret.txt'): Permission denied\nfatal: adding files failed\n")
	assertIndex(t, helper, [][]any{})
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
		if entry.GetPathname() != expectedPath {
			t.Errorf("entry %d: expected path %s, got %s", i, expectedPath, entry.GetPathname())
		}
	}
}
