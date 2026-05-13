package commands

// import (
// 	"encoding/hex"
// 	"fmt"
// 	"path/filepath"
// 	"testing"
// )

// func TestDiff(t *testing.T) {
// 	t.Run("ModifiyContent - WorkspaceFile", func(t *testing.T) {
// 		helper := NewCommandHelper(t)
// 		helper.WriteFile(t, "file.txt", "Hello World")
// 		helper.JitCommand("add", ".")
// 		helper.Commit(t, "first commit")
// 		assertDiff(t, helper, "")

// 		helper.WriteFile(t, "file.txt", "Modified")

// 		helper.Repo(t).Index().Load()

// 		entry := helper.Repo(t).Index().GetEntry("file.txt")

// 		a_oid := entry.GetOid()
// 		a_mode := "100644"

// 		b_oid := helper.HashBlob(t, "file.txt")
// 		b_mode := "100644"
// 		expectedMessage := headerMessage("file.txt", a_oid, a_mode, b_oid, b_mode)
// 		fmt.Println(expectedMessage)
// 		assertDiff(t, helper, expectedMessage)
// 	})

// 	t.Run("ModifiyMode - WorkspaceFile", func(t *testing.T) {
// 		helper := NewCommandHelper(t)
// 		helper.WriteFile(t, "file.txt", "Hello World")
// 		helper.JitCommand("add", ".")
// 		helper.Commit(t, "first commit")
// 		assertDiff(t, helper, "")

// 		// helper.WriteFile(t, "file.txt", "Modified")
// 		helper.MakeExecutable(t, "file.txt")
// 		helper.Repo(t).Index().Load()

// 		entry := helper.Repo(t).Index().GetEntry("file.txt")

// 		a_oid := entry.GetOid()
// 		a_mode := "100644"

// 		b_oid := helper.HashBlob(t, "file.txt")
// 		b_mode := "100755"
// 		expectedMessage := headerMessage("file.txt", a_oid, a_mode, b_oid, b_mode)
// 		fmt.Println(expectedMessage)
// 		assertDiff(t, helper, expectedMessage)
// 	})

// 	t.Run("ModifiyModeAndContent - WorkspaceFile", func(t *testing.T) {
// 		helper := NewCommandHelper(t)
// 		helper.WriteFile(t, "file.txt", "Hello World")
// 		helper.JitCommand("add", ".")
// 		helper.Commit(t, "first commit")
// 		assertDiff(t, helper, "")

// 		helper.WriteFile(t, "file.txt", "Modified")
// 		helper.MakeExecutable(t, "file.txt")
// 		helper.Repo(t).Index().Load()

// 		entry := helper.Repo(t).Index().GetEntry("file.txt")

// 		a_oid := entry.GetOid()
// 		a_mode := "100644"

// 		b_oid := helper.HashBlob(t, "file.txt")
// 		b_mode := "100755"
// 		expectedMessage := headerMessage("file.txt", a_oid, a_mode, b_oid, b_mode)
// 		fmt.Println(expectedMessage)
// 		assertDiff(t, helper, expectedMessage)
// 	})

// 	t.Run("Delete - WorkspaceFile", func(t *testing.T) {
// 		helper := NewCommandHelper(t)
// 		helper.WriteFile(t, "file.txt", "Hello World")
// 		helper.JitCommand("add", ".")
// 		helper.Delete(t, "file.txt")

// 		helper.Repo(t).Index().Load()

// 		entry := helper.Repo(t).Index().GetEntry("file.txt")

// 		a_oid := entry.GetOid()
// 		a_mode := "100644"
// 		a_path := filepath.Join("a", "file.txt")

// 		null_oid := make([]byte, 40)
// 		b_path := filepath.Join("b", "file.txt")

// 		nullPath := "/dev/null"

// 		expectedMessage := fmt.Sprintf("diff --git %s %s\n", a_path, b_path)
// 		expectedMessage += fmt.Sprintf("deleted file mode %s\n", a_mode)
// 		expectedMessage += fmt.Sprintf("index %s..%s\n", short(a_oid), short(null_oid))
// 		expectedMessage += fmt.Sprintf("--- %s\n+++ %s\n", a_path, nullPath)

// 		// fmt.Println(expectedMessage)
// 		assertDiff(t, helper, expectedMessage)
// 	})
// }

// func assertDiff(t *testing.T, helper *CommandHelper, expectedMessage string) {
// 	t.Helper()
// 	helper.JitCommand("diff")
// 	helper.AssertStdout(t, expectedMessage)
// }

// func headerMessage(path string, a_oid []byte, a_mode string, b_oid []byte, b_mode string) string {
// 	a_path := filepath.Join("a", path)
// 	b_path := filepath.Join("b", path)

// 	message := fmt.Sprintf("diff --git %s %s\n", a_path, b_path)
// 	if a_mode != b_mode {
// 		message += fmt.Sprintf("old mode %s\nnew mode %s\n", a_mode, b_mode)
// 	}
// 	if string(a_oid) == string(b_oid) {
// 		return message
// 	}
// 	message += fmt.Sprintf("index %s..%s", short(a_oid), short(b_oid))
// 	if a_mode == b_mode {
// 		message += fmt.Sprintf(" %s", a_mode)
// 	}
// 	message += "\n"
// 	message += fmt.Sprintf("--- %s\n", a_path)
// 	message += fmt.Sprintf("+++ %s\n", b_path)

// 	return message
// }
// func short(oid []byte) string {
// 	return hex.EncodeToString(oid)[:7]
// }
