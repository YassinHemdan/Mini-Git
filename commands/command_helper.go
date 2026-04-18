package commands

import (
	"JIT/internals"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	fmt.Println(argv)
	h.Cmd = Execute(h.repoPath, h.Env, argv, h.Stdin, h.Stdout, h.Stderr)

	return h.Cmd
}

func (h *CommandHelper) Commit(t *testing.T, message string) {
	// h.setStdin(message)
	h.setEnv("JIT_AUTHOR_NAME", "A. U. Thor")
	h.setEnv("JIT_AUTHOR_EMAIL", "author@example.com")
	h.JitCommand("commit", "-m", message)
}

// func (h *CommandHelper) setStdin(content string) {
// 	h.Stdin = strings.NewReader(content)
// }

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
		t.Errorf("Error: expected message '%s' but found '%s'", message, h.Stdout.String())
	}
}
func (h *CommandHelper) AssertStderr(t *testing.T, message string) {
	t.Helper()
	if message != h.Stderr.String() {
		t.Errorf("Error: expected message '%s' but found '%s'", message, h.Stdout.String())
	}
}
