package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	repoDir, err := os.MkdirTemp("", "jit-test-*")
	if err != nil {
		t.Fatalf("Could not create a temp directoru - %v", err)
	}

	defer os.RemoveAll(repoDir)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	ctx := Execute(repoDir, map[string]string{}, []string{"init"}, strings.NewReader(""), stdout, stderr)

	if ctx.Status != 0 {
		t.Fatalf("init failed: %s", stdout.String())
	}

	jitDir := filepath.Join(repoDir, ".jit")
	objectsDir := filepath.Join(jitDir, "objects")
	refsDir := filepath.Join(jitDir, "refs")

	assertDirExists(t, jitDir)
	assertDirExists(t, objectsDir)
	assertDirExists(t, refsDir)

}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("expected directory to exist: %s", path)
		} else {
			t.Errorf("error checking path %s: %v", path, err)
		}

		return
	}

	if !info.IsDir() {
		t.Errorf("expected %s to be a directory, but it's a file", path)
	}

}
