package index

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
)

type testEntry struct {
	path string
	oid  []byte
}

func createEntries(paths ...string) []testEntry {
	entries := make([]testEntry, len(paths))

	for i, p := range paths {
		oid := make([]byte, 20)
		rand.Read(oid)
		entries[i] = testEntry{
			path: p,
			oid:  oid,
		}
	}

	return entries
}

func TestIndex_Add_SingleAndMultipleFiles(t *testing.T) {
	fileInfo, err := os.Stat("index_test.go")
	if err != nil {
		t.Errorf("failed to stat file: %v", err)
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		t.Errorf("failed to stat file: %v", err)
	}
	tests := []struct {
		name          string
		entries       []testEntry
		expectedPaths []string
		wantError     bool
	}{
		{
			name:          "Adding a single file",
			entries:       createEntries("alice.txt"),
			expectedPaths: []string{"alice.txt"},
			wantError:     false,
		},
		{
			name:          "Adding mutiple files",
			entries:       createEntries("alice.txt", "Bob.txt", "Cameron.txt"),
			expectedPaths: []string{"alice.txt", "Bob.txt", "Cameron.txt"},
			wantError:     false,
		},
		{
			name:          "Adding empty files",
			entries:       createEntries(""),
			expectedPaths: []string{},
			wantError:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempPath := os.TempDir()
			indexPath := filepath.Join(tempPath, "index")

			idx, err := NewIndex(indexPath)
			if err != nil {
				t.Errorf("Failed to create index object")
			}

			for _, entry := range tt.entries {
				err := idx.Add(entry.path, entry.oid, stat)
				if tt.wantError && err == nil {
					t.Errorf("An error expected with pathname '%s' but got nil", entry.path)
				} else if !tt.wantError && err != nil {
					t.Errorf("Unexpected error with pathname '%s'", entry.path)
				}
			}

			resultPaths := idx.getKeysSlice()
			if len(resultPaths) != len(tt.expectedPaths) {
				t.Fatalf("Expected %d entries but got %d", len(tt.expectedPaths), len(resultPaths))
			}

			// slices.Sort(resultPaths)
			slices.Sort(tt.expectedPaths)
			for i, expectedPath := range tt.expectedPaths {
				if expectedPath != resultPaths[i] {
					t.Errorf("Entry[%d]: expected %s path but got %s", i, expectedPath, resultPaths[i])
				}
			}
		})
	}
}

func TestIndex_Add_ReplaceFileWithDirectory(t *testing.T) {
	fileInfo, err := os.Stat("index_test.go")

	if err != nil {
		t.Errorf("failed to stat file: %v", err)
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		t.Errorf("failed to stat file: %v", err)
	}

	tests := []struct {
		name          string
		entries       []testEntry
		expectedPaths []string
		wantError     bool
	}{
		{
			name:          "Adding a single file",
			entries:       createEntries("alice.txt"),
			expectedPaths: []string{"alice.txt"},
			wantError:     false,
		},
		{
			name:          "Adding mutiple files",
			entries:       createEntries("alice.txt", "Bob.txt", "Cameron.txt"),
			expectedPaths: []string{"alice.txt", "Bob.txt", "Cameron.txt"},
			wantError:     false,
		},
		{
			name:          "Adding empty files",
			entries:       createEntries(""),
			expectedPaths: []string{},
			wantError:     true,
		},
		{
			name:          "Replacing a file with a directory v1",
			entries:       createEntries("bob.txt", "internals/file.txt", "internals/file.txt/sub/nested.txt"),
			expectedPaths: []string{"bob.txt", "internals/file.txt/sub/nested.txt"},
			wantError:     false,
		},
		{
			name:          "Replacing a file with a directory v2",
			entries:       createEntries("bob.txt", "alice.txt", "alice.txt/nested.txt"),
			expectedPaths: []string{"bob.txt", "alice.txt/nested.txt"},
			wantError:     false,
		},
		{
			name: "Replacing a file with a directory v3",
			entries: createEntries(
				"internals/sub1/file1.txt",
				"internals/sub1/sub2/file2.txt",
				"internals/sub1/sub2/sub3/file3.txt",
				"internals/sub1/sub2/file2.txt/sub3/file3.txt",
			),
			expectedPaths: []string{
				"internals/sub1/file1.txt",
				"internals/sub1/sub2/file2.txt/sub3/file3.txt",
				"internals/sub1/sub2/sub3/file3.txt",
			},
			wantError: false,
		},
		{
			name:    "Replacing a file with a directory v4 - simple single level",
			entries: createEntries("foo.txt", "foo.txt/bar.txt"),
			expectedPaths: []string{
				"foo.txt/bar.txt",
			},
			wantError: false,
		},
		{
			name:    "Replacing a file with a directory v5 - multiple files under replaced path",
			entries: createEntries("data", "data/a.txt", "data/b.txt"),
			expectedPaths: []string{
				"data/a.txt",
				"data/b.txt",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v6 - deep nesting",
			entries: createEntries(
				"a/b/c/d/e.txt",
				"a/b/c/d/e.txt/f/g/h.txt",
			),
			expectedPaths: []string{
				"a/b/c/d/e.txt/f/g/h.txt",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v7 - two files become directories",
			entries: createEntries(
				"a/b.txt",
				"a/b.txt/c.txt",
				"a/b.txt/c.txt/d.txt",
			),
			expectedPaths: []string{
				"a/b.txt/c.txt/d.txt",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v8 - siblings unaffected",
			entries: createEntries(
				"src/main.go",
				"src/utils.go",
				"src/helpers.go",
				"src/main.go/sub/file.txt",
			),
			expectedPaths: []string{
				"src/helpers.go",
				"src/main.go/sub/file.txt",
				"src/utils.go",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v9 - top-level file replaced",
			entries: createEntries(
				"README",
				"LICENSE",
				"README/content.md",
			),
			expectedPaths: []string{
				"LICENSE",
				"README/content.md",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v10 - cascading replacements",
			entries: createEntries(
				"a.txt",
				"a.txt/b.txt",
				"a.txt/b.txt/c.txt",
			),
			expectedPaths: []string{
				"a.txt/b.txt/c.txt",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v11 - multiple independent replacements",
			entries: createEntries(
				"lib/parser.go",
				"lib/lexer.go",
				"cmd/main.go",
				"cmd/cli.go",
				"lib/parser.go/v2/parser.go",
				"cmd/main.go/internal/run.go",
			),
			expectedPaths: []string{
				"cmd/cli.go",
				"cmd/main.go/internal/run.go",
				"lib/lexer.go",
				"lib/parser.go/v2/parser.go",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v12 - conflict deep in the path",
			entries: createEntries(
				"project/src/components/button/index.js",
				"project/src/components/button/styles.css",
				"project/src/components/button/index.js/variants/primary.js",
			),
			expectedPaths: []string{
				"project/src/components/button/index.js/variants/primary.js",
				"project/src/components/button/styles.css",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v13 - similar names unaffected",
			entries: createEntries(
				"docs/guide.md",
				"docs/guide.md.bak",
				"docs/guide.md.old",
				"docs/guide.md/chapter1.md",
			),
			expectedPaths: []string{
				"docs/guide.md.bak",
				"docs/guide.md.old",
				"docs/guide.md/chapter1.md",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v14 - intermediate component conflict",
			entries: createEntries(
				"a/b/c",
				"x/y/z",
				"a/b/c/d/e/f.txt",
			),
			expectedPaths: []string{
				"a/b/c/d/e/f.txt",
				"x/y/z",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v15 - three files replaced by single add",
			entries: createEntries(
				"a",
				"a/b",
				"a/b/c",
				"a/b/c/d/e/f.txt",
			),
			expectedPaths: []string{
				"a/b/c/d/e/f.txt",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v16 - add multiple under replaced dir",
			entries: createEntries(
				"config.json",
				"config.json/dev.json",
				"config.json/prod.json",
				"config.json/staging.json",
			),
			expectedPaths: []string{
				"config.json/dev.json",
				"config.json/prod.json",
				"config.json/staging.json",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v17 - exact match only",
			entries: createEntries(
				"src/app",
				"src/app.test",
				"src/app.config",
				"src/application",
				"src/app/main.go",
			),
			expectedPaths: []string{
				"src/app.config",
				"src/app.test",
				"src/app/main.go",
				"src/application",
			},
			wantError: false,
		},
		{
			name: "Replacing a file with a directory v18 - original v3 extended",
			entries: createEntries(
				"internals/sub1/file1.txt",
				"internals/sub1/sub2/file2.txt",
				"internals/sub1/sub2/sub3/file3.txt",
				"internals/sub1/sub2/sub3/file4.txt",
				"internals/sub1/sub2/file2.txt/sub3/file3.txt",
			),
			expectedPaths: []string{
				"internals/sub1/file1.txt",
				"internals/sub1/sub2/file2.txt/sub3/file3.txt",
				"internals/sub1/sub2/sub3/file3.txt",
				"internals/sub1/sub2/sub3/file4.txt",
			},
			wantError: false,
		},

		// Replacement at root level with many siblings
		{
			name: "Replacing a file with a directory v19 - root level many siblings",
			entries: createEntries(
				"a.txt",
				"b.txt",
				"c.txt",
				"d.txt",
				"e.txt",
				"b.txt/nested/deep/file.go",
			),
			expectedPaths: []string{
				"a.txt",
				"b.txt/nested/deep/file.go",
				"c.txt",
				"d.txt",
				"e.txt",
			},
			wantError: false,
		},

		// Same directory used after replacement, with multiple levels
		{
			name: "Replacing a file with a directory v20 - replacement with multiple nested levels",
			entries: createEntries(
				"pkg/models",
				"pkg/models/user/user.go",
				"pkg/models/user/user_test.go",
				"pkg/models/post/post.go",
			),
			expectedPaths: []string{
				"pkg/models/post/post.go",
				"pkg/models/user/user.go",
				"pkg/models/user/user_test.go",
			},
			wantError: false,
		},

		// Two separate replacements happening at different depths
		{
			name: "Replacing a file with a directory v21 - two replacements at different depths",
			entries: createEntries(
				"a/file1.txt",
				"a/b/c/file2.txt",
				"a/file1.txt/sub.txt",
				"a/b/c/file2.txt/deep/sub.txt",
			),
			expectedPaths: []string{
				"a/b/c/file2.txt/deep/sub.txt",
				"a/file1.txt/sub.txt",
			},
			wantError: false,
		},

		// Edge case: very long path where a mid-level file is replaced
		{
			name: "Replacing a file with a directory v22 - mid-level replacement in long path",
			entries: createEntries(
				"a/b/c/d/e/f/g/h.txt",
				"a/b/c/d/midfile.txt",
				"a/b/c/d/midfile.txt/x/y/z.txt",
			),
			expectedPaths: []string{
				"a/b/c/d/e/f/g/h.txt",
				"a/b/c/d/midfile.txt/x/y/z.txt",
			},
			wantError: false,
		},

		// File at one level replaced, sibling directories remain intact
		{
			name: "Replacing a file with a directory v23 - sibling directories unaffected",
			entries: createEntries(
				"src/index.js",
				"src/components/Header.js",
				"src/components/Footer.js",
				"src/index.js/polyfills/core.js",
			),
			expectedPaths: []string{
				"src/components/Footer.js",
				"src/components/Header.js",
				"src/index.js/polyfills/core.js",
			},
			wantError: false,
		},

		// Replacing and then the replacement itself is valid (no further conflict)
		{
			name: "Replacing a file with a directory v24 - no cascading after replacement",
			entries: createEntries(
				"build/output",
				"build/output/dist/bundle.js",
				"build/output/dist/bundle.css",
			),
			expectedPaths: []string{
				"build/output/dist/bundle.css",
				"build/output/dist/bundle.js",
			},
			wantError: false,
		},
		// {
		// 	name: "Replacing a directory with a file v1",
		// 	entries: createEntries(
		// 		"internals/sub1/file1.txt",
		// 		"internals/sub1/file2.txt",
		// 		"internals/sub1/file3.txt",
		// 		"internals/sub1",
		// 	),
		// 	expectedPaths: []string{"internals/sub1"},
		// 	wantError:     false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempPath := os.TempDir()
			indexPath := filepath.Join(tempPath, "index")

			idx, err := NewIndex(indexPath)
			if err != nil {
				t.Errorf("Failed to create index object")
			}

			for _, entry := range tt.entries {
				err := idx.Add(entry.path, entry.oid, stat)
				if tt.wantError && err == nil {
					t.Errorf("An error expected with pathname '%s' but got nil", entry.path)
				} else if !tt.wantError && err != nil {
					t.Errorf("Unexpected error with pathname '%s'", entry.path)
				}
			}

			resultPaths := idx.getKeysSlice()
			if len(resultPaths) != len(tt.expectedPaths) {
				t.Fatalf("Expected %d entries but got %d", len(tt.expectedPaths), len(resultPaths))
			}

			// slices.Sort(resultPaths)
			slices.Sort(tt.expectedPaths)
			for i, expectedPath := range tt.expectedPaths {
				if expectedPath != resultPaths[i] {
					t.Errorf("Entry[%d]: expected %s path but got %s", i, expectedPath, resultPaths[i])
				}
			}
		})
	}
}

func TestIndex_Add_ReplaceDirectoryWithFile(t *testing.T) {
	fileInfo, err := os.Stat("index_test.go")
	if err != nil {
		t.Errorf("failed to stat file: %v", err)
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		t.Errorf("failed to stat file: %v", err)
	}
	tests := []struct {
		name          string
		entries       []testEntry
		expectedPaths []string
		wantError     bool
	}{
		{
			name: "Replace directory with file v1 - single child removed",
			entries: createEntries(
				"src/main.go",
				"src",
			),
			expectedPaths: []string{"src"},
			wantError:     false,
		},
		{
			name: "Replace directory with file v2 - multiple children removed",
			entries: createEntries(
				"internals/sub1/file1.txt",
				"internals/sub1/file2.txt",
				"internals/sub1/file3.txt",
				"internals/sub1",
			),
			expectedPaths: []string{"internals/sub1"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v3 - deep nesting removed",
			entries: createEntries(
				"a/b/c/d/e.txt",
				"a/b/c/d/f.txt",
				"a/b/c/g.txt",
				"a/b",
			),
			expectedPaths: []string{"a/b"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v4 - siblings unaffected",
			entries: createEntries(
				"lib/parser/parse.go",
				"lib/parser/ast.go",
				"lib/lexer/lex.go",
				"lib/parser",
			),
			expectedPaths: []string{
				"lib/lexer/lex.go",
				"lib/parser",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v5 - root level directory",
			entries: createEntries(
				"docs/guide.md",
				"docs/api.md",
				"docs/tutorial.md",
				"docs",
			),
			expectedPaths: []string{"docs"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v6 - only matching subtree removed",
			entries: createEntries(
				"project/src/components/Button.js",
				"project/src/components/Header.js",
				"project/src/utils/helpers.js",
				"project/tests/test1.js",
				"project/src/components",
			),
			expectedPaths: []string{
				"project/src/components",
				"project/src/utils/helpers.js",
				"project/tests/test1.js",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v7 - intermediate level replacement",
			entries: createEntries(
				"a/b/c/file1.txt",
				"a/b/c/file2.txt",
				"a/b/c/d/file3.txt",
				"a/b/c/d/e/file4.txt",
				"a/b/c",
			),
			expectedPaths: []string{"a/b/c"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v8 - two independent replacements",
			entries: createEntries(
				"pkg/models/user.go",
				"pkg/models/post.go",
				"cmd/server/main.go",
				"cmd/server/routes.go",
				"pkg/models",
				"cmd/server",
			),
			expectedPaths: []string{
				"cmd/server",
				"pkg/models",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v9 - similar prefix unaffected",
			entries: createEntries(
				"app/config/dev.json",
				"app/config/prod.json",
				"app/configs/extra.json",
				"app/configuration/base.json",
				"app/config",
			),
			expectedPaths: []string{
				"app/config",
				"app/configs/extra.json",
				"app/configuration/base.json",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v10 - very deep tree collapsed",
			entries: createEntries(
				"a/b/c/d/e/f/g/h/i/j.txt",
				"a/b/c/d/e/f/g/h/i/k.txt",
				"a/b/c/d/e/f/g/l.txt",
				"a",
			),
			expectedPaths: []string{"a"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v11 - mixed files and subdirs",
			entries: createEntries(
				"build/output/bundle.js",
				"build/output/bundle.css",
				"build/output/assets/logo.png",
				"build/output/assets/icon.svg",
				"build/output",
			),
			expectedPaths: []string{"build/output"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v12 - one sibling dir replaced",
			entries: createEntries(
				"src/a/file1.txt",
				"src/b/file2.txt",
				"src/c/file3.txt",
				"src/b",
			),
			expectedPaths: []string{
				"src/a/file1.txt",
				"src/b",
				"src/c/file3.txt",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v13 - deep replacement siblings survive",
			entries: createEntries(
				"a/x.txt",
				"a/b/y.txt",
				"a/b/c/z.txt",
				"a/b/c/d/file1.txt",
				"a/b/c/d/file2.txt",
				"a/b/c/d",
			),
			expectedPaths: []string{
				"a/b/c/d",
				"a/b/c/z.txt",
				"a/b/y.txt",
				"a/x.txt",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v14 - replace then add new entries",
			entries: createEntries(
				"data/users/alice.json",
				"data/users/bob.json",
				"data/users",
				"data/posts/post1.json",
			),
			expectedPaths: []string{
				"data/posts/post1.json",
				"data/users",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v15 - sequential nested replacements",
			entries: createEntries(
				"a/b/c/file1.txt",
				"a/b/c/file2.txt",
				"a/b/c",
				"a/b/d/file3.txt",
				"a/b",
			),
			expectedPaths: []string{"a/b"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v16 - many children removed",
			entries: createEntries(
				"dir/file1.txt",
				"dir/file2.txt",
				"dir/file3.txt",
				"dir/file4.txt",
				"dir/file5.txt",
				"dir/file6.txt",
				"dir/file7.txt",
				"dir/file8.txt",
				"dir",
			),
			expectedPaths: []string{"dir"},
			wantError:     false,
		},

		{
			name: "Replace directory with file v17 - top-level files survive",
			entries: createEntries(
				"README.md",
				"LICENSE",
				"src/main.go",
				"src/util.go",
				"src/handler/api.go",
				"src",
			),
			expectedPaths: []string{
				"LICENSE",
				"README.md",
				"src",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v18 - replace then re-expand",
			entries: createEntries(
				"pkg/http/server.go",
				"pkg/http/client.go",
				"pkg/http",
				"pkg/http/v2/server.go",
			),
			expectedPaths: []string{
				"pkg/http/v2/server.go",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v19 - substring names unaffected",
			entries: createEntries(
				"test/unit/a.go",
				"test/unit/b.go",
				"testing/integration/c.go",
				"test_utils/helper.go",
				"test/unit",
			),
			expectedPaths: []string{
				"test/unit",
				"test_utils/helper.go",
				"testing/integration/c.go",
			},
			wantError: false,
		},

		{
			name: "Replace directory with file v20 - replace only existing directory",
			entries: createEntries(
				"only/dir/file1.txt",
				"only/dir/file2.txt",
				"only/dir/sub/file3.txt",
				"only",
			),
			expectedPaths: []string{"only"},
			wantError:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempPath := os.TempDir()
			indexPath := filepath.Join(tempPath, "index")

			idx, err := NewIndex(indexPath)
			if err != nil {
				t.Errorf("Failed to create index object")
			}

			for _, entry := range tt.entries {
				err := idx.Add(entry.path, entry.oid, stat)
				if tt.wantError && err == nil {
					t.Errorf("An error expected with pathname '%s' but got nil", entry.path)
				} else if !tt.wantError && err != nil {
					t.Errorf("Unexpected error with pathname '%s'", entry.path)
				}
			}

			resultPaths := idx.getKeysSlice()
			if len(resultPaths) != len(tt.expectedPaths) {
				t.Fatalf("Expected %d entries but got %d", len(tt.expectedPaths), len(resultPaths))
			}

			slices.Sort(tt.expectedPaths)
			for i, expectedPath := range tt.expectedPaths {
				if expectedPath != resultPaths[i] {
					t.Errorf("Entry[%d]: expected %s path but got %s", i, expectedPath, resultPaths[i])
				}
			}
		})
	}
}
