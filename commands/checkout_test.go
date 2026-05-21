package commands

import (
	"slices"
	"testing"
)

type file struct {
	pathname string
	content  string
}

func workspaceFilesSorted(t *testing.T, helper *CommandHelper) []file {
	t.Helper()
	repo := helper.Repo(t)
	files, err := repo.Workspace().ListFiles("")

	if err != nil {
		t.Fatal(err)
	}

	slices.Sort(files)
	result := make([]file, 0)
	for _, filepath := range files {
		content, err := repo.Workspace().ReadFile(filepath)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, file{pathname: filepath, content: string(content)})
	}

	return result
}
func TestCheckout(t *testing.T) {
	t.Run("ChangeContents", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Repo(t)

		name1, content1 := "a/b/file1.txt", "hello1"
		helper.WriteFile(t, name1, content1)
		name2, content2 := "outer.txt", "outer"
		helper.WriteFile(t, name2, content2)
		name3, content3 := "a/nested.txt", "hello world"
		helper.WriteFile(t, name3, content3) // we will not modify this one

		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.WriteFile(t, name1, content1+"changed")
		helper.WriteFile(t, name2, content2+"changed")

		helper.JitCommand("add", ".")
		helper.Commit(t, "second commit")

		helper.JitCommand("checkout", "@^")
		workspaceFiles := workspaceFilesSorted(t, helper)

		expected := make([]file, 0)
		expected = append(expected, file{pathname: name1, content: content1})
		expected = append(expected, file{pathname: name2, content: content2})
		expected = append(expected, file{pathname: name3, content: content3})

		slices.SortFunc(expected, func(a, b file) int {
			if a.pathname < b.pathname {
				return -1
			} else if a.pathname > b.pathname {
				return 1
			}
			return 0
		})

		assertCheckoutCmd(t, workspaceFiles, expected)
	})
	t.Run("RemoveFilesWithChangeContents", func(t *testing.T) {
		/*
			remove files means that we created new files in the latest (or current commit) that does not exist in the
			target commit that we want to switch to it, so these new files should be removed before moving switching
			also after removing a file, if any of its parent directories becomes empty, we delete it as well
		*/
		helper := NewCommandHelper(t)
		helper.Repo(t)

		name1, content1 := "a/b/file1.txt", "hello1"
		helper.WriteFile(t, name1, content1)
		name2, content2 := "outer.txt", "outer"
		helper.WriteFile(t, name2, content2)
		name3, content3 := "a/nested.txt", "hello world"
		helper.WriteFile(t, name3, content3) // we will not modify this one

		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.WriteFile(t, name1, content1+"changed")
		helper.WriteFile(t, name2, content2+"changed")
		helper.WriteFile(t, "sub1/sub2/sub3/newfile.txt", "hello from new file") // this one should not be included to our result

		helper.JitCommand("add", ".")
		helper.Commit(t, "second commit")

		helper.JitCommand("checkout", "@^")
		workspaceFiles := workspaceFilesSorted(t, helper)

		expected := make([]file, 0)
		expected = append(expected, file{pathname: name1, content: content1})
		expected = append(expected, file{pathname: name2, content: content2})
		expected = append(expected, file{pathname: name3, content: content3})

		slices.SortFunc(expected, func(a, b file) int {
			if a.pathname < b.pathname {
				return -1
			} else if a.pathname > b.pathname {
				return 1
			}
			return 0
		})

		assertCheckoutCmd(t, workspaceFiles, expected)
	})
	t.Run("CreateFilesWithChangeContents", func(t *testing.T) {
		/*
			This is the opposite of remove files
			create files means there was some files in the target commit that we want to move to it but they got removed in
			the current commit, we we want to CREATE these files before switching to the target commit
		*/
		helper := NewCommandHelper(t)
		helper.Repo(t)

		name1, content1 := "a/b/file1.txt", "hello1"
		helper.WriteFile(t, name1, content1)
		name2, content2 := "outer.txt", "outer"
		helper.WriteFile(t, name2, content2)
		name3, content3 := "a/nested.txt", "hello world"
		helper.WriteFile(t, name3, content3) // we will not modify this one

		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.WriteFile(t, name1, content1+"changed")
		helper.WriteFile(t, name2, content2+"changed")
		helper.Delete(t, name3)

		/*
			for now we will delete the .jit/index before making the second add and commit as we are not handling the
			index file during the switching, this will be handled soon but for now we will delete the index file before every new commit
			except for the first one for sure
		*/

		/*
			what will happen if we did not delete the index file ?
			the differences between the two trees of the two commits won't be recognized and by that there won't be any changes
			done to prepare the workspace to match the tree in the target commit(that contains 3 files) and the current
			tree will remain the same (with 2 files)
		*/
		helper.Delete(t, ".jit/index")
		helper.JitCommand("add", ".")
		helper.Commit(t, "second commit")

		helper.JitCommand("checkout", "@^")
		workspaceFiles := workspaceFilesSorted(t, helper)

		expected := make([]file, 0)
		expected = append(expected, file{pathname: name1, content: content1})
		expected = append(expected, file{pathname: name2, content: content2})
		expected = append(expected, file{pathname: name3, content: content3})

		slices.SortFunc(expected, func(a, b file) int {
			if a.pathname < b.pathname {
				return -1
			} else if a.pathname > b.pathname {
				return 1
			}
			return 0
		})

		assertCheckoutCmd(t, workspaceFiles, expected)
	})

	t.Run("Mix", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Repo(t)

		helper.WriteFile(t, "a/todo1.txt", "hello from todo1")
		helper.WriteFile(t, "a/todo2.txt", "hello from todo2")
		helper.WriteFile(t, "a/todo3.txt", "hello from todo3")
		helper.WriteFile(t, "a/todo4.txt", "hello from todo4")
		helper.WriteFile(t, "a/todo5.txt", "hello from todo5")

		helper.WriteFile(t, "remove.txt", "remove me") // will be removed

		// in the next commit, the b.txt directory will be removed and we will create file a/b.txt
		// in other words, replacing a directory with a file
		helper.WriteFile(t, "a/b.txt/file1.txt", "hello from file1")
		helper.WriteFile(t, "a/b.txt/file2.txt", "hello from file2")
		helper.WriteFile(t, "a/b.txt/file3.txt", "hello from file3")

		// this file will be removed and we will create a directory in the new commit
		// replacing a file with a directory
		helper.WriteFile(t, "a/nested.txt", "hello from nested.txt")

		expectedC1 := workspaceFilesSorted(t, helper)

		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")
		helper.JitCommand("branch", "c1", "@")

		helper.Delete(t, ".jit/index")
		helper.Delete(t, "remove.txt")
		helper.WriteFile(t, "a/todo1.txt", "hello from todo1"+"changed")
		helper.WriteFile(t, "a/todo2.txt", "hello from todo2"+"changed")
		helper.WriteFile(t, "a/todo3.txt", "hello from todo3"+"changed")

		helper.Delete(t, "a/b.txt")
		helper.WriteFile(t, "a/b.txt", "Hello from b.txt")

		helper.Delete(t, "a/nested.txt")
		helper.WriteFile(t, "a/nested.txt/nes1.txt", "Hello from nes1.txt")
		helper.WriteFile(t, "a/nested.txt/nes2.txt", "Hello from nes2.txt")
		helper.WriteFile(t, "a/nested.txt/nes3.txt", "Hello from nes3.txt")

		expectedC2 := workspaceFilesSorted(t, helper)

		helper.JitCommand("add", ".")
		helper.Commit(t, "second commit")
		helper.JitCommand("branch", "c2", "@") // we will be back to it

		helper.JitCommand("checkout", "c1")
		workspaceFiles := workspaceFilesSorted(t, helper)
		assertCheckoutCmd(t, workspaceFiles, expectedC1)
		// helper.Delete(t, ".jit/index")
		helper.JitCommand("checkout", "c2")
		workspaceFiles = workspaceFilesSorted(t, helper)
		assertCheckoutCmd(t, workspaceFiles, expectedC2)
	})
}

func assertCheckoutCmd(t *testing.T, workspace, expected []file) {
	t.Helper()
	if len(workspace) != len(expected) {
		t.Fatalf("expected %d files but found %d\n", len(expected), len(workspace))
	}
	for i := 0; i < len(workspace); i++ {
		expectedFile := expected[i]
		workspaceFile := workspace[i]

		if expectedFile.pathname != workspaceFile.pathname {
			t.Fatalf("expected pathname '%s' but found '%s'\n", expectedFile.pathname, workspaceFile.pathname)
		}
		if expectedFile.content != workspaceFile.content {
			t.Fatalf("for file %s: expected content '%s' but found '%s'\n", expectedFile.pathname, expectedFile.content, workspaceFile.content)
		}
	}
}
