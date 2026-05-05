package commands

import (
	"fmt"
	"testing"
)

func TestStatus_Untracked_Basic(t *testing.T) {
	t.Run("EmptyRepositoryShowsNothing", func(t *testing.T) {
		helper := NewCommandHelper(t)
		assertStatus(t, helper, "")
	})

	t.Run("SingleUntrackedFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "hello")
		assertStatus(t, helper, format("??", "file.txt"))
	})

	t.Run("MultipleUntrackedFilesSortedAlphabetically", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "z.txt", "")
		helper.WriteFile(t, "a.txt", "")
		helper.WriteFile(t, "m.txt", "")
		assertStatus(t, helper,
			format("??", "a.txt")+
				format("??", "m.txt")+
				format("??", "z.txt"))
	})

	t.Run("AllTrackedShowsNothing", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a.txt", "")
		helper.WriteFile(t, "b.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")
		assertStatus(t, helper, "")
	})
}

func TestStatus_Untracked_DirectoryCollapsing(t *testing.T) {
	t.Run("CollapsesUntrackedDirWithOneFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("CollapsesUntrackedDirWithMultipleFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/a.txt", "")
		helper.WriteFile(t, "dir/b.txt", "")
		helper.WriteFile(t, "dir/c.txt", "")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("CollapsesDeeplyNestedUntrackedDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/b/c/d/e/file.txt", "")
		assertStatus(t, helper, format("??", "a/"))
	})

	t.Run("CollapsesDirWithMixOfFilesAndSubdirs", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "")
		helper.WriteFile(t, "dir/sub1/a.txt", "")
		helper.WriteFile(t, "dir/sub2/b.txt", "")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("MultipleCollapsedSiblingDirs", func(t *testing.T) {
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
}

func TestStatus_Untracked_DirectoryExpanding(t *testing.T) {
	t.Run("ExpandsDirWhenSomeFilesAreTracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/tracked.txt", "")
		helper.WriteFile(t, "dir/untracked.txt", "")
		helper.JitCommand("add", "dir/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper, format("??", "dir/untracked.txt"))
	})

	t.Run("ExpandsDirShowsUntrackedSubdirs", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/tracked.txt", "")
		helper.WriteFile(t, "dir/sub1/file.txt", "")
		helper.WriteFile(t, "dir/sub2/file.txt", "")
		helper.JitCommand("add", "dir/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper,
			format("??", "dir/sub1/")+
				format("??", "dir/sub2/"))
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

	t.Run("ExpandsDeepNestedWhenIntermediateTracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/tracked.txt", "")
		helper.WriteFile(t, "a/b/tracked.txt", "")
		helper.WriteFile(t, "a/b/c/file1.txt", "")
		helper.WriteFile(t, "a/b/c/file2.txt", "")
		helper.JitCommand("add", "a/tracked.txt", "a/b/tracked.txt")
		helper.Commit(t, "commit")
		assertStatus(t, helper, format("??", "a/b/c/"))
	})

	t.Run("NewFileInFullyTrackedDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file1.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "dir/file2.txt", "new")
		assertStatus(t, helper, format("??", "dir/file2.txt"))
	})

	t.Run("NewSubdirInFullyTrackedDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "dir/sub/a.txt", "")
		helper.WriteFile(t, "dir/sub/b.txt", "")
		assertStatus(t, helper, format("??", "dir/sub/"))
	})

	t.Run("DirCollapsesAfterAllFilesTracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file1.txt", "")
		helper.WriteFile(t, "dir/file2.txt", "")
		helper.JitCommand("add", "dir/file1.txt")
		helper.Commit(t, "commit 1")
		assertStatus(t, helper, format("??", "dir/file2.txt"))

		helper.JitCommand("add", "dir/file2.txt")
		helper.Commit(t, "commit 2")
		assertStatus(t, helper, "")
	})

	t.Run("DirExpandsAfterPartialCommit", func(t *testing.T) {
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
}

func TestStatus_Untracked_EmptyDirectories(t *testing.T) {
	t.Run("IgnoresEmptyDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "empty")
		assertStatus(t, helper, "")
	})

	t.Run("IgnoresNestedEmptyDirs", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "a/b/c")
		assertStatus(t, helper, "")
	})

	t.Run("IgnoresMultipleEmptyDirs", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "empty1")
		helper.Mkdir(t, "empty2")
		helper.Mkdir(t, "empty3")
		assertStatus(t, helper, "")
	})

	t.Run("IgnoresEmptySiblingDirs", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "parent/child1")
		helper.Mkdir(t, "parent/child2")
		assertStatus(t, helper, "")
	})

	t.Run("ShowsFileButIgnoresEmptySiblingDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "hello.txt", "hello")
		helper.Mkdir(t, "empty-dir")
		assertStatus(t, helper, format("??", "hello.txt"))
	})

	t.Run("IgnoresEmptyDirAlongsideTrackedFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")
		helper.Mkdir(t, "empty")
		assertStatus(t, helper, "")
	})

	t.Run("PreviouslyEmptyDirBecomesNonEmpty", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "dir")
		assertStatus(t, helper, "")

		helper.WriteFile(t, "dir/file.txt", "content")
		assertStatus(t, helper, format("??", "dir/"))
	})

	t.Run("TreeWithEmptyLeavesOnly", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "root/b1/leaf1")
		helper.Mkdir(t, "root/b1/leaf2")
		helper.Mkdir(t, "root/b2/leaf1")
		assertStatus(t, helper, "")
	})

	t.Run("TreeWithOneFileDeepInEmptyNesting", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.Mkdir(t, "root/b1/leaf1")
		helper.Mkdir(t, "root/b1/leaf2")
		helper.WriteFile(t, "root/b2/leaf1/file.txt", "content")
		assertStatus(t, helper, format("??", "root/"))
	})

	t.Run("FileInsideDirWithEmptySubdir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "parent/file.txt", "content")
		helper.Mkdir(t, "parent/empty-child")
		assertStatus(t, helper, format("??", "parent/"))
	})
}

func TestStatus_Modified_Content(t *testing.T) {
	t.Run("NothingModifiedShowsNothing", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.WriteFile(t, "a/2.txt", "two")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")
		assertStatus(t, helper, "")
	})

	t.Run("SingleModifiedFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file.txt", "changed")
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("MultipleModifiedFilesSorted", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "c.txt", "one")
		helper.WriteFile(t, "a.txt", "two")
		helper.WriteFile(t, "b.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "c.txt", "changed1")
		helper.WriteFile(t, "a.txt", "changed2")
		assertStatus(t, helper,
			format(" M", "a.txt")+
				format(" M", "c.txt"))
	})

	t.Run("ModifiedFileInSubdirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "dir/file.txt", "modified")
		assertStatus(t, helper, format(" M", "dir/file.txt"))
	})

	t.Run("ModifiedFilesAcrossMultipleDirectories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/file.txt", "one")
		helper.WriteFile(t, "b/file.txt", "two")
		helper.WriteFile(t, "c/file.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "a/file.txt", "changed1")
		helper.WriteFile(t, "c/file.txt", "changed2")
		assertStatus(t, helper,
			format(" M", "a/file.txt")+
				format(" M", "c/file.txt"))
	})

	t.Run("ModifiedContentSameSize", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "aaa")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file.txt", "bbb")
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("ModifiedDeeplyNestedFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/b/c/d/file.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "a/b/c/d/file.txt", "modified")
		assertStatus(t, helper, format(" M", "a/b/c/d/file.txt"))
	})

	t.Run("UnmodifiedFileNotReported", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "changed.txt", "one")
		helper.WriteFile(t, "unchanged.txt", "two")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "changed.txt", "modified")
		assertStatus(t, helper, format(" M", "changed.txt"))
	})

	t.Run("ModifyThenRevertBackShowsNothing", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file.txt", "changed")
		assertStatus(t, helper, format(" M", "file.txt"))

		helper.WriteFile(t, "file.txt", "original")
		assertStatus(t, helper, "")
	})
}

func TestStatus_Modified_Mode(t *testing.T) {
	t.Run("UnmodifiedModeShowsNothing", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")
		assertStatus(t, helper, "")
	})

	t.Run("MakeFileExecutable", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.MakeExecutable(t, "file.txt")
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("MakeMultipleFilesExecutable", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a.txt", "one")
		helper.WriteFile(t, "b.txt", "two")
		helper.WriteFile(t, "c.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.MakeExecutable(t, "a.txt")
		helper.MakeExecutable(t, "c.txt")
		assertStatus(t, helper,
			format(" M", "a.txt")+
				format(" M", "c.txt"))
	})

	t.Run("MakeNestedFileExecutable", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/script.sh", "#!/bin/bash")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.MakeExecutable(t, "dir/script.sh")
		assertStatus(t, helper, format(" M", "dir/script.sh"))
	})
}

func TestStatus_Modified_Timestamps(t *testing.T) {
	t.Run("TouchFileDoesNotReportAsModified", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Touch(t, "file.txt")
		assertStatus(t, helper, "")
	})

	t.Run("TouchThenModifyReportsAsModified", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Touch(t, "file.txt")
		assertStatus(t, helper, "")

		helper.WriteFile(t, "file.txt", "changed")
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("TouchMultipleFilesNoneReportedAsModified", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a.txt", "one")
		helper.WriteFile(t, "b.txt", "two")
		helper.WriteFile(t, "c.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Touch(t, "a.txt")
		helper.Touch(t, "b.txt")
		helper.Touch(t, "c.txt")
		assertStatus(t, helper, "")
	})

	t.Run("TouchOneFileModifyAnother", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a.txt", "one")
		helper.WriteFile(t, "b.txt", "two")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Touch(t, "a.txt")
		helper.WriteFile(t, "b.txt", "changed")
		assertStatus(t, helper, format(" M", "b.txt"))
	})

	t.Run("SecondStatusAfterTouchIsAlsoClean", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Touch(t, "file.txt")
		assertStatus(t, helper, "") // first call updates index timestamps
		assertStatus(t, helper, "") // second call should also be clean
	})
}

func TestStatus_Deleted(t *testing.T) {
	t.Run("SingleDeletedFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "file.txt")
		assertStatus(t, helper, format(" D", "file.txt"))
	})

	t.Run("MultipleDeletedFilesSorted", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "c.txt", "one")
		helper.WriteFile(t, "a.txt", "two")
		helper.WriteFile(t, "b.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a.txt")
		helper.Delete(t, "c.txt")
		assertStatus(t, helper,
			format(" D", "a.txt")+
				format(" D", "c.txt"))
	})

	t.Run("DeleteEntireDirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/file1.txt", "one")
		helper.WriteFile(t, "dir/file2.txt", "two")
		helper.WriteFile(t, "dir/file3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "dir")
		assertStatus(t, helper,
			format(" D", "dir/file1.txt")+
				format(" D", "dir/file2.txt")+
				format(" D", "dir/file3.txt"))
	})

	t.Run("DeleteNestedDirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/b/file1.txt", "one")
		helper.WriteFile(t, "a/b/file2.txt", "two")
		helper.WriteFile(t, "a/file3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a/b")
		assertStatus(t, helper,
			format(" D", "a/b/file1.txt")+
				format(" D", "a/b/file2.txt"))
	})

	t.Run("DeleteDeeplyNestedFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/b/c/d/file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a/b/c/d/file.txt")
		assertStatus(t, helper, format(" D", "a/b/c/d/file.txt"))
	})

	t.Run("DeleteSomeFilesInDirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/keep.txt", "keep")
		helper.WriteFile(t, "dir/remove.txt", "remove")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "dir/remove.txt")
		assertStatus(t, helper, format(" D", "dir/remove.txt"))
	})

	t.Run("DeleteAllFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a.txt", "one")
		helper.WriteFile(t, "b.txt", "two")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a.txt")
		helper.Delete(t, "b.txt")
		assertStatus(t, helper,
			format(" D", "a.txt")+
				format(" D", "b.txt"))
	})

	t.Run("DeleteFilesAcrossMultipleDirectories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/file.txt", "one")
		helper.WriteFile(t, "b/file.txt", "two")
		helper.WriteFile(t, "c/file.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a/file.txt")
		helper.Delete(t, "c/file.txt")
		assertStatus(t, helper,
			format(" D", "a/file.txt")+
				format(" D", "c/file.txt"))
	})
}

func TestStatus_Mixed_ModifiedAndUntracked(t *testing.T) {
	t.Run("ModifiedAndUntrackedFilesTogether", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "tracked.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "tracked.txt", "changed")
		helper.WriteFile(t, "new.txt", "untracked")
		assertStatus(t, helper,
			format(" M", "tracked.txt")+
				format("??", "new.txt"))
	})

	t.Run("ModifiedFileAndUntrackedDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file.txt", "modified")
		helper.WriteFile(t, "newdir/file.txt", "")
		assertStatus(t, helper,
			format(" M", "file.txt")+
				format("??", "newdir/"))
	})

	t.Run("MultipleModifiedAndMultipleUntracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a.txt", "one")
		helper.WriteFile(t, "b.txt", "two")
		helper.WriteFile(t, "c.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "a.txt", "changed")
		helper.WriteFile(t, "c.txt", "changed")
		helper.WriteFile(t, "d.txt", "new")
		helper.WriteFile(t, "e.txt", "new")
		assertStatus(t, helper,
			format(" M", "a.txt")+
				format(" M", "c.txt")+
				format("??", "d.txt")+
				format("??", "e.txt"))
	})
}

func TestStatus_Mixed_DeletedAndUntracked(t *testing.T) {
	t.Run("DeletedAndUntrackedFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "old.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "old.txt")
		helper.WriteFile(t, "new.txt", "content")
		assertStatus(t, helper,
			format(" D", "old.txt")+
				format("??", "new.txt"))
	})

	t.Run("DeleteDirAndAddNewUntrackedDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "old/file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "old")
		helper.WriteFile(t, "new/file.txt", "content")
		assertStatus(t, helper,
			format(" D", "old/file.txt")+
				format("??", "new/"))
	})
}

func TestStatus_Mixed_DeletedAndModified(t *testing.T) {
	t.Run("DeletedAndModifiedFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "delete-me.txt", "content")
		helper.WriteFile(t, "modify-me.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "delete-me.txt")
		helper.WriteFile(t, "modify-me.txt", "changed")
		assertStatus(t, helper,
			format(" D", "delete-me.txt")+
				format(" M", "modify-me.txt"))
	})

	t.Run("DeleteSomeModifyOthersInSameDir", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/a.txt", "one")
		helper.WriteFile(t, "dir/b.txt", "two")
		helper.WriteFile(t, "dir/c.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "dir/a.txt")
		helper.WriteFile(t, "dir/c.txt", "modified")
		assertStatus(t, helper,
			format(" D", "dir/a.txt")+
				format(" M", "dir/c.txt"))
	})
}

func TestStatus_Mixed_AllThree(t *testing.T) {
	t.Run("DeletedModifiedAndUntracked", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "delete.txt", "content")
		helper.WriteFile(t, "modify.txt", "original")
		helper.WriteFile(t, "keep.txt", "unchanged")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "delete.txt")
		helper.WriteFile(t, "modify.txt", "changed")
		helper.WriteFile(t, "untracked.txt", "new")
		assertStatus(t, helper,
			format(" D", "delete.txt")+
				format(" M", "modify.txt")+
				format("??", "untracked.txt"))
	})

	t.Run("AllThreeAcrossMultipleDirectories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/file1.txt", "one")
		helper.WriteFile(t, "b/file2.txt", "two")
		helper.WriteFile(t, "c/file3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a/file1.txt")
		helper.WriteFile(t, "b/file2.txt", "modified")
		helper.WriteFile(t, "d/file4.txt", "new")
		assertStatus(t, helper,
			format(" D", "a/file1.txt")+
				format(" M", "b/file2.txt")+
				format("??", "d/"))
	})

	t.Run("AllThreeWithNestedDirectories", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "src/main.go", "package main")
		helper.WriteFile(t, "src/utils/helper.go", "package utils")
		helper.WriteFile(t, "docs/readme.md", "# README")
		helper.WriteFile(t, "config.yml", "key: value")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "docs")
		helper.WriteFile(t, "config.yml", "key: newvalue")
		helper.WriteFile(t, "src/main.go", "package main\nfunc main() {}")
		helper.WriteFile(t, "tests/test1.go", "package tests")
		helper.WriteFile(t, "newfile.txt", "hello")
		assertStatus(t, helper,
			format(" M", "config.yml")+
				format(" D", "docs/readme.md")+
				format(" M", "src/main.go")+
				format("??", "newfile.txt")+
				format("??", "tests/"))
	})

	t.Run("AllThreeWithTouchAndModeChange", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a.txt", "one")
		helper.WriteFile(t, "b.txt", "two")
		helper.WriteFile(t, "c.txt", "three")
		helper.WriteFile(t, "d.txt", "four")
		helper.WriteFile(t, "e.txt", "five")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a.txt")             // deleted
		helper.WriteFile(t, "b.txt", "mod")   // content modified
		helper.MakeExecutable(t, "c.txt")     // mode modified
		helper.Touch(t, "d.txt")              // just touched, no change
		helper.WriteFile(t, "new.txt", "new") // untracked
		assertStatus(t, helper,
			format(" D", "a.txt")+
				format(" M", "b.txt")+
				format(" M", "c.txt")+
				format("??", "new.txt"))
	})
}

func TestStatus_Complex(t *testing.T) {
	t.Run("LargeNumberOfFilesAllStates", func(t *testing.T) {
		helper := NewCommandHelper(t)
		// Track many files
		for i := 0; i < 20; i++ {
			helper.WriteFile(t, fmt.Sprintf("file%02d.txt", i), fmt.Sprintf("content%d", i))
		}
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "file00.txt")
		helper.Delete(t, "file02.txt")
		helper.Delete(t, "file04.txt")

		helper.WriteFile(t, "file01.txt", "changed")
		helper.WriteFile(t, "file03.txt", "changed")
		helper.WriteFile(t, "file05.txt", "changed")

		helper.WriteFile(t, "untracked1.txt", "new")
		helper.WriteFile(t, "untracked2.txt", "new")

		assertStatus(t, helper,
			format(" D", "file00.txt")+
				format(" M", "file01.txt")+
				format(" D", "file02.txt")+
				format(" M", "file03.txt")+
				format(" D", "file04.txt")+
				format(" M", "file05.txt")+
				format("??", "untracked1.txt")+
				format("??", "untracked2.txt"))
	})

	t.Run("DeleteDirThenAddNewUntrackedDirSameName", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "dir/old.txt", "old content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "dir")
		helper.WriteFile(t, "dir/new.txt", "new content")
		assertStatus(t, helper,
			format(" D", "dir/old.txt")+
				format("??", "dir/new.txt"))
	})

	//TODO LATER
	// t.Run("DeleteFileAndCreateDirWithSameName", func(t *testing.T) {
	// 	helper := NewCommandHelper(t)
	// 	helper.WriteFile(t, "name.txt", "content")
	// 	helper.JitCommand("add", ".")
	// 	helper.Commit(t, "commit")

	// 	helper.Delete(t, "name.txt")
	// 	helper.WriteFile(t, "name.txt/file.txt", "nested")
	// 	assertStatus(t, helper,
	// 		format(" D", "name.txt")+
	// 			format("??", "name.txt/"))
	// })

	t.Run("DeeplyNestedMixedStates", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "a/b/c/d1.txt", "one")
		helper.WriteFile(t, "a/b/c/d2.txt", "two")
		helper.WriteFile(t, "a/b/e/f1.txt", "three")
		helper.WriteFile(t, "a/b/e/f2.txt", "four")
		helper.WriteFile(t, "a/g.txt", "five")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "a/b/c/d1.txt")
		helper.WriteFile(t, "a/b/c/d2.txt", "modified")
		helper.Delete(t, "a/b/e")
		helper.WriteFile(t, "a/b/new/file.txt", "new")
		helper.Touch(t, "a/g.txt")
		assertStatus(t, helper,
			format(" D", "a/b/c/d1.txt")+
				format(" M", "a/b/c/d2.txt")+
				format(" D", "a/b/e/f1.txt")+
				format(" D", "a/b/e/f2.txt")+
				format("??", "a/b/new/"))
	})

	t.Run("StatusIsCleanAfterMultipleCommits", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "v1")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit 1")

		helper.WriteFile(t, "file.txt", "v2")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit 2")

		helper.WriteFile(t, "file.txt", "v3")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit 3")

		assertStatus(t, helper, "")
	})

	t.Run("ConsecutiveStatusCallsAreIdempotent", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file.txt", "modified")
		assertStatus(t, helper, format(" M", "file.txt"))
		assertStatus(t, helper, format(" M", "file.txt"))
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("ModifyDeleteRecreateFile", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "original")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "file.txt")
		helper.WriteFile(t, "file.txt", "recreated")
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("DeleteAndRecreateWithSameContent", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "file.txt")
		helper.WriteFile(t, "file.txt", "content")

		assertStatus(t, helper, "")
	})

	t.Run("EmptyFileTrackedThenPopulated", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file.txt", "now has content")
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("PopulatedFileThenEmptied", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "file.txt", "has content")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.WriteFile(t, "file.txt", "")
		assertStatus(t, helper, format(" M", "file.txt"))
	})

	t.Run("SortingMixedStatusesAndPaths", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "alpha.txt", "a")
		helper.WriteFile(t, "beta.txt", "b")
		helper.WriteFile(t, "gamma.txt", "g")
		helper.WriteFile(t, "delta.txt", "d")
		helper.WriteFile(t, "dir/inner.txt", "i")
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit")

		helper.Delete(t, "beta.txt")
		helper.WriteFile(t, "delta.txt", "modified")
		helper.WriteFile(t, "gamma.txt", "modified")
		helper.Delete(t, "dir/inner.txt")
		helper.WriteFile(t, "aaa.txt", "new")
		helper.WriteFile(t, "zzz/file.txt", "new")
		assertStatus(t, helper,
			format(" D", "beta.txt")+
				format(" M", "delta.txt")+
				format(" D", "dir/inner.txt")+
				format(" M", "gamma.txt")+
				format("??", "aaa.txt")+
				format("??", "zzz/"))
	})

	t.Run("RealWorldProjectSimulation", func(t *testing.T) {
		helper := NewCommandHelper(t)

		helper.WriteFile(t, "README.md", "# My Project")
		helper.WriteFile(t, "go.mod", "module myproject")
		helper.WriteFile(t, "main.go", "package main")
		helper.WriteFile(t, "cmd/root.go", "package cmd")
		helper.WriteFile(t, "cmd/version.go", "package cmd")
		helper.WriteFile(t, "internal/db/db.go", "package db")
		helper.WriteFile(t, "internal/db/migrations.go", "package db")
		helper.WriteFile(t, "internal/api/handler.go", "package api")
		helper.WriteFile(t, "internal/api/middleware.go", "package api")
		helper.WriteFile(t, "configs/dev.yml", "env: dev")
		helper.WriteFile(t, "configs/prod.yml", "env: prod")
		helper.JitCommand("add", ".")
		helper.Commit(t, "initial commit")

		// Simulate development work:
		helper.WriteFile(t, "main.go", "package main\nimport \"fmt\"")
		helper.WriteFile(t, "internal/api/handler.go", "package api\n// new")
		helper.Delete(t, "cmd/version.go")
		helper.Delete(t, "configs/dev.yml")
		helper.WriteFile(t, "internal/api/router.go", "package api")
		helper.WriteFile(t, "internal/cache/cache.go", "package cache")
		helper.WriteFile(t, "scripts/deploy.sh", "#!/bin/bash")
		helper.WriteFile(t, "test/api_test.go", "package test")
		helper.Mkdir(t, "build")
		helper.Touch(t, "README.md")

		assertStatus(t, helper,
			format(" D", "cmd/version.go")+
				format(" D", "configs/dev.yml")+
				format(" M", "internal/api/handler.go")+
				format(" M", "main.go")+
				format("??", "internal/api/router.go")+
				format("??", "internal/cache/")+
				format("??", "scripts/")+
				format("??", "test/"))
	})
}

func TestStatus_HeadIndexChanges(t *testing.T) {
	t.Run("Head/Index changes_AddOnly", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.WriteFile(t, "a/2.txt", "two")
		helper.WriteFile(t, "a/b/3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.WriteFile(t, "a/4.txt", "four")
		assertStatus(t, helper, format("??", "a/4.txt"))

		helper.JitCommand("add", ".")
		assertStatus(t, helper, format("A ", "a/4.txt"))

		helper.WriteFile(t, "d/e/5.txt", "five")
		assertStatus(t, helper, format("A ", "a/4.txt")+format("??", "d/"))

		helper.JitCommand("add", ".")
		assertStatus(t, helper, format("A ", "a/4.txt")+format("A ", "d/e/5.txt"))
	})
	t.Run("Head/Index changes_AddAndModify", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.WriteFile(t, "a/2.txt", "two")
		helper.WriteFile(t, "a/b/3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.WriteFile(t, "a/4.txt", "four")
		assertStatus(t, helper, format("??", "a/4.txt"))

		helper.JitCommand("add", ".")
		assertStatus(t, helper, format("A ", "a/4.txt"))

		helper.WriteFile(t, "a/4.txt", "a change")
		assertStatus(t, helper, format("AM", "a/4.txt"))

		helper.JitCommand("add", ".")
		assertStatus(t, helper, format("A ", "a/4.txt"))
	})
	t.Run("Head/Index changes_ModifyIndexMode", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.WriteFile(t, "a/2.txt", "two")
		helper.WriteFile(t, "a/b/3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.MakeExecutable(t, "1.txt")
		helper.JitCommand("add", ".")
		assertStatus(t, helper, "M  1.txt\n")
	})

	t.Run("Head/Index changes_ModifyIndexContent", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.WriteFile(t, "a/2.txt", "two")
		helper.WriteFile(t, "a/b/3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.WriteFile(t, "a/b/3.txt", "modify")
		helper.JitCommand("add", ".")
		assertStatus(t, helper, "M  a/b/3.txt\n")
	})
	t.Run("Head/Index changes_ModifyIndexAndWorkspace", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")
		assertStatus(t, helper, "")

		// 1- modify the content and added it
		helper.WriteFile(t, "1.txt", "modification1")
		helper.JitCommand("add", ".")
		assertStatus(t, helper, "M  1.txt\n")

		// 2- modify it again
		helper.WriteFile(t, "1.txt", "modification2")
		assertStatus(t, helper, "MM 1.txt\n")
	})
	t.Run("Head/Index changes_DeleteFiles", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.WriteFile(t, "a/2.txt", "two")
		helper.WriteFile(t, "a/b/3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.Delete(t, "1.txt")
		helper.Delete(t, ".jit/index")
		helper.JitCommand("add", ".")

		assertStatus(t, helper, "D  1.txt\n")
	})
	t.Run("Head/Index changes_DeleteDirectory", func(t *testing.T) {
		helper := NewCommandHelper(t)
		helper.WriteFile(t, "1.txt", "one")
		helper.WriteFile(t, "a/2.txt", "two")
		helper.WriteFile(t, "a/b/3.txt", "three")
		helper.JitCommand("add", ".")
		helper.Commit(t, "first commit")

		helper.Delete(t, "a")
		helper.Delete(t, ".jit/index")
		helper.JitCommand("add", ".")

		assertStatus(t, helper, format("D ", "a/2.txt")+format("D ", "a/b/3.txt"))
	})
}
func assertStatus(t *testing.T, helper *CommandHelper, statusOutput string) {
	t.Helper()
	helper.JitCommand("status", "--porcelain")
	helper.AssertStdout(t, statusOutput)
}

func format(status, pathname string) string {
	return fmt.Sprintf("%s %s\n", status, pathname)
}
