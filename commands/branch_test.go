package commands

import (
	"JIT/internals"
	database "JIT/internals/database"
	"fmt"
	"testing"
)

func TestBranch(t *testing.T) {
	t.Run("TestingRevisions-NoErrorsExpected-HEAD", func(t *testing.T) {
		helper := NewCommandHelper(t)
		makeCommits(helper, t, 10)

		tests := []struct {
			branchName string
			revision   string
			nthParent  int
		}{
			{"branch0", "HEAD", 0},
			{"branch1", "HEAD^", 1},
			{"branch2", "HEAD^^", 2},
			{"branch3", "HEAD^^^", 3},
			{"branch4", "HEAD~1", 1},
			{"branch5", "HEAD~2", 2},
			{"branch6", "HEAD~3", 3},
			{"branch7", "HEAD~4", 4},
			{"branch8", "HEAD~5", 5},
			{"branch9", "HEAD~6", 6},
			{"branch10", "HEAD~7", 7},
			{"branch11", "HEAD~8", 8},
			{"branch12", "HEAD~9", 9},
			{"branch13", "HEAD^~1", 2},
			{"branch14", "HEAD^~2", 3},
			{"branch15", "HEAD^~3", 4},
			{"branch16", "HEAD^^~1", 3},
			{"branch17", "HEAD^^~2", 4},
			{"branch18", "HEAD^^~3", 5},
			{"branch19", "HEAD~3^", 4},
			{"branch20", "HEAD~4^", 5},
			{"branch21", "HEAD~5^", 6},
			{"branch22", "HEAD~2~3", 5},
			{"branch23", "HEAD~3~2", 5},
			{"branch24", "HEAD^^~3~1", 6},
			{"branch25", "HEAD^~1^~2", 5},
			{"branch26", "HEAD^^^~2", 5},
			{"branch27", "HEAD~4^^", 6},
			{"branch28", "HEAD^~2~2^", 6},
			{"branch29", "HEAD^^^^", 4},
			{"branch30", "HEAD^^^^^", 5},
			{"branch31", "HEAD^^^^^^", 6},
			{"branch32", "HEAD^^^^^^^", 7},
			{"branch33", "HEAD^^^^^^^^", 8},
			{"branch34", "HEAD^^^^^^^^^", 9},
			{"branch35", "HEAD^^^~4", 7},
			{"branch36", "HEAD^^~4", 6},
			{"branch37", "HEAD^~4", 5},
			{"branch38", "HEAD~6^", 7},
			{"branch39", "HEAD~7^", 8},
			{"branch40", "HEAD~8^", 9},
			{"branch41", "HEAD^^~2^", 5},
			{"branch42", "HEAD^~3^", 5},
			{"branch43", "HEAD^~2^^", 5},
			{"branch44", "HEAD~1~1~1", 3},
			{"branch45", "HEAD~2~1~2", 5},
			{"branch46", "HEAD^~1~1~1", 4},
			{"branch47", "HEAD^^~1~1~1", 5},
			{"branch48", "HEAD^^^~1~1~1", 6},
			{"branch49", "HEAD^^^^~5", 9},
		}

		for _, tt := range tests {
			t.Run(tt.revision, func(t *testing.T) {
				helper.JitCommand("branch", tt.branchName, tt.revision)
				assertBranch(helper, t, tt.branchName, tt.nthParent)
			})
		}
	})

	t.Run("TestingRevisions-NoErrorsExpected-AtSign", func(t *testing.T) {
		helper := NewCommandHelper(t)
		makeCommits(helper, t, 10)

		tests := []struct {
			branchName string
			revision   string
			nthParent  int
		}{
			{"atBranch0", "@", 0},
			{"atBranch1", "@^", 1},
			{"atBranch2", "@^^", 2},
			{"atBranch3", "@~3", 3},
			{"atBranch4", "@^~3", 4},
			{"atBranch5", "@~2~3", 5},
			{"atBranch6", "@^^~3~1", 6},
			{"atBranch7", "@^^^~4", 7},
			{"atBranch8", "@~8", 8},
			{"atBranch9", "@~8^", 9},
		}

		for _, tt := range tests {
			t.Run(tt.revision, func(t *testing.T) {
				helper.JitCommand("branch", tt.branchName, tt.revision)
				assertBranch(helper, t, tt.branchName, tt.nthParent)
			})
		}
	})

	t.Run("TestingRevisions-NoErrorsExpected-BranchName", func(t *testing.T) {
		helper := NewCommandHelper(t)
		makeCommits(helper, t, 10)

		startBranch := "main"
		helper.JitCommand("branch", startBranch)

		tests := []struct {
			branchName string
			revision   string
			nthParent  int
		}{
			{"namedBranch0", startBranch, 0},
			{"namedBranch1", startBranch + "^", 1},
			{"namedBranch2", startBranch + "^^", 2},
			{"namedBranch3", startBranch + "~3", 3},
			{"namedBranch4", startBranch + "^~3", 4},
			{"namedBranch5", startBranch + "~2~3", 5},
			{"namedBranch6", startBranch + "^^~3~1", 6},
			{"namedBranch7", startBranch + "^^^~4", 7},
			{"namedBranch8", startBranch + "~8", 8},
			{"namedBranch9", startBranch + "~8^", 9},
		}

		for _, tt := range tests {
			t.Run(tt.revision, func(t *testing.T) {
				helper.JitCommand("branch", tt.branchName, tt.revision)
				assertBranch(helper, t, tt.branchName, tt.nthParent)
			})
		}
	})
}

func TestBranch_v2(t *testing.T) {
	t.Run("TestingRevisions-NoErrorsExpected-FromCreatedBranches", func(t *testing.T) {
		helper := NewCommandHelper(t)
		makeCommits(helper, t, 10)
		helper.JitCommand("branch", "baseA", "HEAD")
		helper.JitCommand("branch", "baseB", "HEAD~3")
		helper.JitCommand("branch", "baseC", "HEAD~5")
		helper.JitCommand("branch", "baseD", "HEAD~8")

		tests := []struct {
			branchName string
			revision   string
			nthParent  int
		}{
			{"fromBaseA0", "baseA", 0},
			{"fromBaseA1", "baseA^", 1},
			{"fromBaseA2", "baseA^^", 2},
			{"fromBaseA3", "baseA~3", 3},
			{"fromBaseA4", "baseA^~3", 4},
			{"fromBaseA5", "baseA^^~3~1", 6},
			{"fromBaseA6", "baseA~8", 8},
			{"fromBaseA7", "baseA~9", 9},
			{"fromBaseB0", "baseB", 3},
			{"fromBaseB1", "baseB^", 4},
			{"fromBaseB2", "baseB^^", 5},
			{"fromBaseB3", "baseB~1", 4},
			{"fromBaseB4", "baseB~2", 5},
			{"fromBaseB5", "baseB~3", 6},
			{"fromBaseB6", "baseB^~2", 6},
			{"fromBaseB7", "baseB^^^", 6},
			{"fromBaseC0", "baseC", 5},
			{"fromBaseC1", "baseC^", 6},
			{"fromBaseC2", "baseC^^", 7},
			{"fromBaseC3", "baseC~1", 6},
			{"fromBaseC4", "baseC~2", 7},
			{"fromBaseC5", "baseC~3", 8},
			{"fromBaseC6", "baseC~4", 9},
			{"fromBaseD0", "baseD", 8},
			{"fromBaseD1", "baseD^", 9},
			{"fromBaseD2", "baseD~1", 9},
		}

		for _, tt := range tests {
			t.Run(tt.revision, func(t *testing.T) {
				helper.JitCommand("branch", tt.branchName, tt.revision)
				assertBranch(helper, t, tt.branchName, tt.nthParent)
			})
		}
	})
}

func makeCommits(helper *CommandHelper, t *testing.T, n int) {
	t.Helper()
	content := "hello"
	filename := "file.txt"
	helper.WriteFile(t, filename, "")
	for i := 1; i <= n; i++ {
		content = content + fmt.Sprint(i)
		helper.WriteFile(t, filename, content)
		helper.JitCommand("add", ".")
		helper.Commit(t, "commit"+fmt.Sprint(i))
	}
}

func assertBranch(helper *CommandHelper, t *testing.T, branchName string, nthParent int) {
	t.Helper()
	repo := helper.Repo(t)
	branchOid, err := getBranchOid(repo, branchName)
	if err != nil {
		t.Fatalf("Error: an error %v was not expected to happen\n", err)
	}

	nthParentOid, err := getNthParentOid(repo, nthParent)
	if err != nil {
		t.Fatalf("Error: an error (%v) was not expected to happen\n", err)
	}

	if string(branchOid) != string(nthParentOid) {
		t.Fatalf("Error: expected oid '%s' but found '%s'\n", string(nthParentOid), string(branchOid))
	}
}

func getBranchOid(repo *internals.Repository, branchName string) ([]byte, error) {
	return repo.Refs().ReadRef(branchName)
}

func getNthParentOid(repo *internals.Repository, nth int) ([]byte, error) {
	oid, err := repo.Refs().ReadHead()
	if err != nil {
		return nil, err
	}

	for range nth {
		oid, err = commitParentOid(repo, oid)
		if err != nil {
			return nil, err
		}
	}

	return oid, nil
}

func commitParentOid(repo *internals.Repository, oid []byte) ([]byte, error) {
	commit, err := repo.Database().Load(oid)
	if err != nil {
		return nil, err
	}

	return commit.(*database.Commit).GetParentOid(), nil
}
