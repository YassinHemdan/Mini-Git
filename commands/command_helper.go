package commands

import (
	"JIT/internals"
	database "JIT/internals/database"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type CommandHelper struct {
	repoPath   string
	repository *internals.Repository
	Stdin      *strings.Reader
	Stdout     *bytes.Buffer
	Stderr     *bytes.Buffer
	Env        map[string]string
	Cmd        *CommandContext
}

func NewCommandHelper(t *testing.T) *CommandHelper {
	t.Helper()

	repoDir, err := os.MkdirTemp("", "jit-test-*")
	if err != nil {
		t.Fatalf("Could not create a temp directory - %v", err)
	}

	helper := &CommandHelper{
		repoPath: repoDir,
		Env:      make(map[string]string),
	}

	helper.JitCommand("init")

	t.Cleanup(func() {
		os.RemoveAll(repoDir)
	})
	return helper
}

func NewCommandHelperWithTestRepo(t *testing.T) *CommandHelper {
	t.Helper()
	root, _ := os.Getwd()
	root = filepath.Dir(root)
	// I think that it would be better to copy this .jit to a temp folder and work on it and after that
	// we clear it as we don't want to corrupt our current copy of it. TODO THIS LATER
	// for now we will use the current copy directly

	repoDir := filepath.Join(root, "testingRepo")
	helper := &CommandHelper{
		repoPath: repoDir,
		Env:      make(map[string]string),
	}
	return helper
}

func (h *CommandHelper) RepoPath() string {
	return h.repoPath
}

func (h *CommandHelper) Repo(t *testing.T) *internals.Repository {
	t.Helper()

	if h.repository == nil {
		jitDir := filepath.Join(h.repoPath, ".jit")
		repo, err := internals.NewRepository(jitDir)
		if err != nil {
			t.Fatalf("Could not initialize repository - %v", err)
		}

		h.repository = repo
	}

	return h.repository
}

func (h *CommandHelper) WriteFile(t *testing.T, name, contents string) {
	t.Helper()

	pathname := filepath.Join(h.repoPath, name)
	if err := os.MkdirAll(filepath.Dir(pathname), 0744); err != nil {
		t.Fatalf("Could not create directories for %s - %v", name, err)
	}

	if err := os.WriteFile(pathname, []byte(contents), 0644); err != nil {
		t.Fatalf("Could not write file %s: %v", pathname, err)
	}

}

func (h *CommandHelper) Touch(t *testing.T, name string) {
	t.Helper()
	pathname := filepath.Join(h.repoPath, name)
	now := time.Now()

	if err := os.Chtimes(pathname, now, now); err != nil {
		file, err := os.Create(pathname)

		if err != nil {
			t.Fatalf("Could not touch file %s: %v", pathname, err)
		}
		file.Close()
	}
}

func (h *CommandHelper) Delete(t *testing.T, name string) {
	t.Helper()
	pathname := filepath.Join(h.repoPath, name)
	files, err := h.Repo(t).Workspace().ListFiles("")
	if err != nil {
		t.Fatalf("%v", err)
	}
	fmt.Println(files)
	if err := os.RemoveAll(pathname); err != nil {
		t.Fatalf("Could not delete %s - %v", pathname, err)
	}
	files, _ = h.repository.Workspace().ListFiles("")
	fmt.Println(files)
}

func (h *CommandHelper) Mkdir(t *testing.T, name string) {
	t.Helper()
	pathname := filepath.Join(h.repoPath, name)
	if err := os.MkdirAll(pathname, 0744); err != nil {
		t.Fatalf("Could not create directories for %s - %v", name, err)
	}
}

func (h *CommandHelper) MakeExecutable(t *testing.T, name string) {
	t.Helper()

	pathname := filepath.Join(h.repoPath, name)
	if err := os.Chmod(pathname, 0755); err != nil {
		t.Fatalf("Could not make %s executable: %v", name, err)
	}
}

func (h *CommandHelper) MakeUnreadable(t *testing.T, name string) {
	t.Helper()

	pathname := filepath.Join(h.repoPath, name)
	if err := os.Chmod(pathname, 0000); err != nil {
		t.Fatalf("Could not make %s unreadable: %v", name, err)
	}
}

func (h *CommandHelper) JitCommand(argv ...string) *CommandContext {
	h.Stdin = strings.NewReader("")
	h.Stdout = &bytes.Buffer{}
	h.Stderr = &bytes.Buffer{}

	// fmt.Println(argv)
	h.Cmd = Execute(h.repoPath, h.Env, argv, h.Stdin, h.Stdout, h.Stderr)
	time.Sleep(5 * time.Millisecond) // flaky tests. Should we handle them in a different way ??
	return h.Cmd
}

func (h *CommandHelper) Commit(t *testing.T, message string) {
	t.Helper()
	h.setEnv("JIT_AUTHOR_NAME", "A. U. Thor")
	h.setEnv("JIT_AUTHOR_EMAIL", "author@example.com")
	h.JitCommand("commit", "-m", message)
}

func (h *CommandHelper) HashBlob(t *testing.T, pathname string) []byte {
	t.Helper()
	content, err := h.repository.Workspace().ReadFile(pathname)
	if err != nil {
		t.Fatalf("Could not read file from workspace - %v\n", err)
	}
	blob := database.NewBlob(content)
	oid, err := h.repository.Database().HashObject(blob)
	if err != nil {
		t.Fatalf("Could not hash an object - %v\n", err)
	}

	return oid
}

func (h *CommandHelper) setEnv(key, value string) {
	h.Env[key] = value
}

func (h *CommandHelper) AssertStatus(t *testing.T, status int) {
	t.Helper()
	if status != h.Cmd.Status {
		t.Errorf("Error: expected status %d but found %d", status, h.Cmd.Status)
	}
}

func (h *CommandHelper) AssertStdout(t *testing.T, message string) {
	t.Helper()
	if message != h.Stdout.String() {
		t.Errorf("Error: expected message '%s'\nbut found '%s'\n", message, h.Stdout.String())
	}
}

func (h *CommandHelper) AssertStderr(t *testing.T, message string) {
	t.Helper()
	if message != h.Stderr.String() {
		t.Errorf("Error: expected message '%s'\nbut found '%s'\n", message, h.Stderr.String())
	}
}

// func (h *CommandHelper) Load(t *testing.T)
