package commands

import (
	"fmt"
	"testing"
)

func TestStatus_UntrackedFiles(t *testing.T) {
	t.Run("EmptyRepository", func(t *testing.T) {
		helper := NewCommandHelper(t)
		assertStatus(t, helper, "")
	})

	t.Run("SingleUntrackedFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "")
		assertStatus(t, helper, format("??", "file.txt"))
	})

	t.Run("MultipleUntrackedFilesInAlphabeticalOrder", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "zebra.txt", "")
		helper.WriteFile(t, "apple.txt", "")
		helper.WriteFile(t, "mango.txt", "")
		assertStatus(t, helper,
			format("??", "apple.txt")+
				format("??", "mango.txt")+
				format("??", "zebra.txt"))
	})

	t.Run("AllFilesTrackedShowsNothing", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "")
		helper.WriteFile(t, "dir/file2.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")
		assertStatus(t, helper, "")
	})

	t.Run("CollapsesUntrackedDirectoryWithSingleFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("CollapsesUntrackedDirectoryWithMultipleFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file1.txt", "")
		helper.WriteFile(t, "dir/file2.txt", "")
		helper.WriteFile(t, "dir/file3.txt", "")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("CollapsesDeepNestedUntrackedDirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/b/c/d/file.txt", "")
		assertStatus(t, helper, format("??", "a/"))
	})

	t.Run("CollapsesDirectoryWithMixOfFilesAndSubdirectories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "")
		helper.WriteFile(t, "dir/sub1/file1.txt", "")
		helper.WriteFile(t, "dir/sub2/file2.txt", "")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("CollapsesMultipleNestedLevelsAllUntracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/sub1/sub2/sub3/file.txt", "")
		helper.WriteFile(t, "dir/sub1/sub2/file.txt", "")
		helper.WriteFile(t, "dir/sub1/file.txt", "")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("ExpandsDirectoryWhenSomeFilesAreTracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/tracked.txt", "")
		helper.WriteFile(t, "dir/untracked.txt", "")
		helper.JitCommand("add", "dir/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper, format("??", "dir/untracked.txt"))
	})

	t.Run("ExpandsDirectoryShowsUntrackedSubdirsWhenParentPartiallyTracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/tracked.txt", "")
		helper.WriteFile(t, "dir/sub1/file1.txt", "")
		helper.WriteFile(t, "dir/sub1/file2.txt", "")
		helper.WriteFile(t, "dir/sub2/file3.txt", "")
		helper.JitCommand("add", "dir/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "dir/sub1/")+
				format("??", "dir/sub2/"))
	})

	t.Run("ExpandsDirectoryShowsMixOfFilesAndSubdirs", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/tracked.txt", "")
		helper.WriteFile(t, "dir/untracked.txt", "")
		helper.WriteFile(t, "dir/sub/file.txt", "")
		helper.JitCommand("add", "dir/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "dir/sub/")+
				format("??", "dir/untracked.txt"))
	})

	t.Run("ExpandsOnlyAtLevelWhereTrackingExists", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/sub/tracked.txt", "")
		helper.WriteFile(t, "dir/sub/untracked.txt", "")
		helper.WriteFile(t, "dir/sub/deep/file.txt", "")
		helper.JitCommand("add", "dir/sub/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "dir/sub/deep/")+
				format("??", "dir/sub/untracked.txt"))
	})

	t.Run("ExpandsDeepNestedWhenIntermediateFileIsTracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/tracked.txt", "")
		helper.WriteFile(t, "a/b/tracked.txt", "")
		helper.WriteFile(t, "a/b/c/file1.txt", "")
		helper.WriteFile(t, "a/b/c/file2.txt", "")
		helper.JitCommand("add", "a/tracked.txt", "a/b/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper, format("??", "a/b/c/"))
	})

	t.Run("ShowsMultipleUntrackedSiblingDirectories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "alpha/file.txt", "")
		helper.WriteFile(t, "beta/file.txt", "")
		helper.WriteFile(t, "gamma/file.txt", "")
		assertStatus(t, helper,
			format("??", "alpha/")+
				format("??", "beta/")+
				format("??", "gamma/"))
	})

	t.Run("ShowsMultipleUntrackedSubdirsAfterParentFileTracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/tracked.txt", "")
		helper.WriteFile(t, "dir/sub1/file.txt", "")
		helper.WriteFile(t, "dir/sub2/file.txt", "")
		helper.WriteFile(t, "dir/sub3/file.txt", "")
		helper.JitCommand("add", "dir/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "dir/sub1/")+
				format("??", "dir/sub2/")+
				format("??", "dir/sub3/"))
	})

	t.Run("SortsFilesAndDirectoriesTogetherAlphabetically", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "aaa.txt", "")
		helper.WriteFile(t, "dir/file.txt", "")
		helper.WriteFile(t, "zzz.txt", "")
		assertStatus(t, helper,
			format("??", "aaa.txt")+
				format("??", "dir/")+
				format("??", "zzz.txt"))
	})

	t.Run("DirectoryNameSortsCorrectlyAmongFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "b.txt", "")
		helper.WriteFile(t, "a/file.txt", "")
		helper.WriteFile(t, "c.txt", "")
		assertStatus(t, helper,
			format("??", "a/")+
				format("??", "b.txt")+
				format("??", "c.txt"))
	})

	t.Run("TrackedFilesDisappearFromStatusAfterCommit", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file1.txt", "")
		helper.WriteFile(t, "file2.txt", "")
		assertStatus(t, helper,
			format("??", "file1.txt")+
				format("??", "file2.txt"))

		helper.JitCommand("add", "file1.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper, format("??", "file2.txt"))
	})

	t.Run("DirectoryCollapsesAfterUntrackedFileAdded", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file1.txt", "")
		helper.WriteFile(t, "dir/file2.txt", "")
		helper.JitCommand("add", "dir/file1.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper, format("??", "dir/file2.txt"))

		helper.JitCommand("add", "dir/file2.txt")
		helper.Commit(t, "commit 2")
		assertStatus(t, helper, "")
	})

	t.Run("DirectoryExpandsToSubdirsAfterPartialCommit", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "")
		helper.WriteFile(t, "dir/sub1/file1.txt", "")
		helper.WriteFile(t, "dir/sub2/file2.txt", "")
		assertStatus(t, helper, format("??", "dir/"))

		helper.JitCommand("add", "dir/file.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "dir/sub1/")+
				format("??", "dir/sub2/"))
	})

	t.Run("SubdirExpandsToFilesAfterPartialCommitInSubdir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "")
		helper.WriteFile(t, "dir/sub/file1.txt", "")
		helper.WriteFile(t, "dir/sub/file2.txt", "")
		assertStatus(t, helper, format("??", "dir/"))

		helper.JitCommand("add", "dir/file.txt", "dir/sub/file1.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper, format("??", "dir/sub/file2.txt"))
	})

	t.Run("RootLevelMixOfTrackedAndUntrackedFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked1.txt", "")
		helper.WriteFile(t, "tracked2.txt", "")
		helper.WriteFile(t, "untracked1.txt", "")
		helper.WriteFile(t, "untracked2.txt", "")
		helper.JitCommand("add", "tracked1.txt", "tracked2.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "untracked1.txt")+
				format("??", "untracked2.txt"))
	})

	t.Run("NewFileInAlreadyFullyTrackedDirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file1.txt", "")
		helper.WriteFile(t, "dir/file2.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "dir/file3.txt", "")
		assertStatus(t, helper, format("??", "dir/file3.txt"))
	})

	t.Run("NewSubdirInAlreadyFullyTrackedDirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file1.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "dir/sub/file2.txt", "")
		helper.WriteFile(t, "dir/sub/file3.txt", "")
		assertStatus(t, helper, format("??", "dir/sub/"))
	})

	t.Run("ComplexTreeWithMixedTrackingAtMultipleLevels", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/tracked.txt", "")
		helper.WriteFile(t, "a/b/tracked.txt", "")
		helper.WriteFile(t, "a/b/untracked.txt", "")
		helper.WriteFile(t, "a/b/c/file.txt", "")
		helper.WriteFile(t, "a/d/file.txt", "")
		helper.JitCommand("add", "a/tracked.txt", "a/b/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "a/b/c/")+
				format("??", "a/b/untracked.txt")+
				format("??", "a/d/"))
	})

	t.Run("SiblingDirsWithDifferentTrackingStates", func(t *testing.T) {
		helper := NewCommandHelper(t)

		helper.WriteFile(t, "dir1/file1.txt", "")
		helper.WriteFile(t, "dir1/file2.txt", "")

		helper.WriteFile(t, "dir2/tracked.txt", "")
		helper.WriteFile(t, "dir2/untracked.txt", "")

		helper.WriteFile(t, "dir3/file1.txt", "")

		helper.JitCommand("add", "dir2/tracked.txt", "dir3/file1.txt")
		helper.Commit(t, "commit")

		assertStatus(t, helper,
			format("??", "dir1/")+
				format("??", "dir2/untracked.txt"))
	})

	t.Run("DeeplyNestedPartialTrackingOnlyExpandsToCorrectLevel", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/b/c/tracked.txt", "")
		helper.WriteFile(t, "a/b/c/untracked.txt", "")
		helper.WriteFile(t, "a/b/c/d/file.txt", "")
		helper.WriteFile(t, "a/b/c/e/file.txt", "")
		helper.JitCommand("add", "a/b/c/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "a/b/c/d/")+
				format("??", "a/b/c/e/")+
				format("??", "a/b/c/untracked.txt"))
	})

	t.Run("MultipleNewFilesAcrossTrackedAndUntrackedDirs", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file1.txt", "")
		helper.WriteFile(t, "dir1/file2.txt", "")
		helper.WriteFile(t, "dir2/file3.txt", "")
		helper.WriteFile(t, "dir2/file4.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file5.txt", "")
		helper.WriteFile(t, "dir1/file6.txt", "")
		helper.WriteFile(t, "dir2/sub/file7.txt", "")
		helper.WriteFile(t, "newdir/file8.txt", "")
		assertStatus(t, helper,
			format("??", "dir1/file6.txt")+
				format("??", "dir2/sub/")+
				format("??", "file5.txt")+
				format("??", "newdir/"))
	})

	t.Run("ignores empty directory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "empty-dir")
		assertStatus(t, helper, "")
	})

	t.Run("ignores nested empty directories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "outer/inner")
		assertStatus(t, helper, "")
	})

	t.Run("ignores deeply nested empty directories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "a/b/c/d/e")
		assertStatus(t, helper, "")
	})

	t.Run("ignores multiple empty directories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "empty1")
		helper.Mkdir(t, "empty2")
		helper.Mkdir(t, "empty3")
		assertStatus(t, helper, "")
	})

	t.Run("ignores sibling empty directories inside a parent", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "parent/child1")
		helper.Mkdir(t, "parent/child2")
		helper.Mkdir(t, "parent/child3")
		assertStatus(t, helper, "")
	})

	t.Run("lists file but ignores empty sibling directory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "hello.txt", "hello")
		helper.Mkdir(t, "empty-dir")
		assertStatus(t, helper, "?? hello.txt\n")
	})

	t.Run("lists directory with file but ignores empty sibling directory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "full-dir/file.txt", "content")
		helper.Mkdir(t, "empty-dir")
		assertStatus(t, helper, "?? full-dir/\n")
	})

	t.Run("lists file inside directory that also has empty subdirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "parent/file.txt", "content")
		helper.Mkdir(t, "parent/empty-child")
		assertStatus(t, helper, "?? parent/\n")
	})

	t.Run("ignores empty directory alongside tracked files", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked.txt", "content")
		helper.JitCommand("add", "tracked.txt")
		helper.Commit(t, "add tracked file")

		helper.Mkdir(t, "empty-dir")
		assertStatus(t, helper, "")
	})

	t.Run("ignores empty directory but shows untracked file alongside tracked files", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked.txt", "content")
		helper.JitCommand("add", "tracked.txt")
		helper.Commit(t, "add tracked file")

		helper.Mkdir(t, "empty-dir")
		helper.WriteFile(t, "untracked.txt", "new content")
		assertStatus(t, helper, "?? untracked.txt\n")
	})

	t.Run("directory with one file and multiple empty siblings", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "project/src/main.go", "package main")
		helper.Mkdir(t, "project/docs")
		helper.Mkdir(t, "project/tests")
		helper.Mkdir(t, "project/build")
		assertStatus(t, helper, "?? project/\n")
	})

	t.Run("file at root with many nested empty directories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "readme.txt", "hello")
		helper.Mkdir(t, "a/b/c")
		helper.Mkdir(t, "x/y/z")
		helper.Mkdir(t, "empty")
		assertStatus(t, helper, "?? readme.txt\n")
	})

	t.Run("previously empty directory becomes non-empty", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "dir")
		assertStatus(t, helper, "")

		helper.WriteFile(t, "dir/file.txt", "content")
		assertStatus(t, helper, "?? dir/\n")
	})

	t.Run("tree with empty leaves only", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "root/branch1/leaf1")
		helper.Mkdir(t, "root/branch1/leaf2")
		helper.Mkdir(t, "root/branch2/leaf1")
		assertStatus(t, helper, "")
	})

	t.Run("tree with one file deep in nested directories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "root/branch1/leaf1")
		helper.Mkdir(t, "root/branch1/leaf2")
		helper.WriteFile(t, "root/branch2/leaf1/file.txt", "content")
		assertStatus(t, helper, "?? root/\n")
	})

	t.Run("complex: tracked files, untracked files, and empty directories", func(t *testing.T) {
		helper := NewCommandHelper(t)

		helper.WriteFile(t, "tracked1.txt", "content1")
		helper.WriteFile(t, "src/tracked2.txt", "content2")
		helper.JitCommand("add", ".")
		helper.Commit(t, "initial commit")

		helper.WriteFile(t, "untracked.txt", "new")
		helper.WriteFile(t, "lib/helper.go", "package lib")
		helper.Mkdir(t, "empty1")
		helper.Mkdir(t, "empty2/nested")
		helper.Mkdir(t, "src/empty-child")

		assertStatus(t, helper, "?? lib/\n?? untracked.txt\n")
	})
}

func assertStatus(t *testing.T, helper *CommandHelper, statusOutput string) {
	t.Helper()
	helper.JitCommand("status")
	helper.AssertStdout(t, statusOutput)
}

func format(status, pathname string) string {
	return fmt.Sprintf("%s %s\n", status, pathname)
}
