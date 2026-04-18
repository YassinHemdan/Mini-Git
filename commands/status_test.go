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
		// dir1: fully untracked
		helper.WriteFile(t, "dir1/file1.txt", "")
		helper.WriteFile(t, "dir1/file2.txt", "")
		// dir2: partially tracked
		helper.WriteFile(t, "dir2/tracked.txt", "")
		helper.WriteFile(t, "dir2/untracked.txt", "")
		// dir3: fully tracked
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
}

func TestStatus_ListUntrackedDirectoriesV2(t *testing.T) {
	t.Run("V1", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "")
		helper.WriteFile(t, "dir/another.txt", "")
		assertStatus(t, helper, format("??", "dir/")+format("??", "file.txt"))

		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.WriteFile(t, "file2.txt", "")
		helper.WriteFile(t, "dir/another2.txt", "")
		assertStatus(t, helper, format("??", "dir/another2.txt")+format("??", "file2.txt"))
	})

	t.Run("V2", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "")
		helper.WriteFile(t, "dir/another1.txt", "")
		helper.WriteFile(t, "dir/another2.txt", "")
		assertStatus(t, helper, format("??", "dir/")+format("??", "file.txt"))

		helper.JitCommand("add", "dir/another1.txt")
		helper.Commit(t, "first commit")

		assertStatus(t, helper, format("??", "dir/another2.txt")+format("??", "file.txt"))
	})
	t.Run("V3", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file1.txt", "")
		helper.WriteFile(t, "file2.txt", "")
		helper.WriteFile(t, "dir/file4.txt", "")
		helper.WriteFile(t, "dir/sub1/file5.txt", "")
		helper.WriteFile(t, "dir/sub1/file6.txt", "")
		helper.WriteFile(t, "dir/sub1/file7.txt", "")
		helper.WriteFile(t, "dir/sub2/file8.txt", "")
		helper.WriteFile(t, "dir/sub2/file9.txt", "")
		helper.WriteFile(t, "dir/sub2/file10.txt", "")
		assertStatus(t, helper,
			format("??", "dir/")+
				format("??", "file1.txt")+
				format("??", "file2.txt"))

		helper.JitCommand("add", "file1.txt", "file2.txt")
		helper.Commit(t, "first commit")

		assertStatus(t, helper, format("??", "dir/"))

		helper.JitCommand("add", "dir/file4.txt")
		helper.Commit(t, "second commit")

		assertStatus(t, helper, format("??", "dir/sub1/")+format("??", "dir/sub2/"))
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
