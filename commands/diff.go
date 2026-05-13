package commands

import (
	"JIT/commands/utils"
	"JIT/diff"
	"JIT/internals"
	database "JIT/internals/database"
	utilsPkg "JIT/utils"
	"fmt"
	"path/filepath"
)

type diffInfo struct {
	oid  []byte
	path string
	mode string
	data string
}

func newDiffInfo(oid []byte, path, mode, data string) *diffInfo {
	return &diffInfo{
		oid:  oid,
		path: path,
		mode: mode,
		data: data,
	}
}

func (di *diffInfo) diffPath() string {
	if string(di.oid) == string(make([]byte, 40)) {
		return "/dev/null"
	}
	return di.path
}

type diffCommandHandler struct {
	ctx    *CommandContext
	repo   *internals.Repository
	status *internals.Status
}

func DiffCommand(ctx *CommandContext) {
	handler := &diffCommandHandler{ctx: ctx, status: nil}
	handler.run()
}

func (h *diffCommandHandler) run() {
	repo, err := utils.Repo(h.ctx.Dir)
	h.repo = repo
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Can't initialize repository: %v\n", err)
		h.ctx.Status = 128
		return
	}
	if _, err := h.repo.Index().Load(); err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't load index - %v\n", err)
		h.ctx.Status = 128
		return
	}
	h.status, err = h.repo.Status()
	if err != nil {
		fmt.Fprintf(h.ctx.Stderr, "Couldn't get status - %v\n", err)
		h.ctx.Status = 128
		return
	}
	if len(h.ctx.Args) == 0 {
		if err := h.diffIndexWorkspace(); err != nil {
			fmt.Fprintf(h.ctx.Stderr, "Error: Couldn't run diff command - %v\n", err)
			h.ctx.Status = 128
			return
		}
	} else if h.ctx.Args[0] == "--staged" || h.ctx.Args[0] == "--cached" {
		if err := h.diffHeadIndex(); err != nil {
			fmt.Fprintf(h.ctx.Stderr, "Error: Couldn't run diff --cached command - %v\n", err)
			h.ctx.Status = 128
			return
		}
	}
}

func (h *diffCommandHandler) diffIndexWorkspace() error {

	workspaceChanges := h.status.GetWorkspaceChanges()
	for path, state := range workspaceChanges {
		switch state {
		case internals.MODIFIED:
			a, err := h.fromIndex(path)
			if err != nil {
				return err
			}

			b, err := h.fromFile(path)
			if err != nil {
				return err
			}

			h.printDiff(a, b)
		case internals.DELETED:
			a, err := h.fromIndex(path)
			if err != nil {
				return err
			}
			b := h.fromNothing(path)

			h.printDiff(a, b)
		}
	}

	return nil
}
func (h *diffCommandHandler) diffHeadIndex() error {
	indexChanges := h.status.GetIndexChanges()
	for path, state := range indexChanges {
		switch state {
		case internals.ADDED:
			a := h.fromNothing(path)
			b, err := h.fromIndex(path)
			if err != nil {
				return err
			}
			h.printDiff(a, b)
		case internals.MODIFIED:
			a, err := h.fromHead(path)
			if err != nil {
				return err
			}

			b, err := h.fromIndex(path)
			if err != nil {
				return err
			}

			h.printDiff(a, b)
		case internals.DELETED:
			a, err := h.fromHead(path)
			if err != nil {
				return nil
			}

			b := h.fromNothing(path)

			h.printDiff(a, b)
		}

	}

	return nil
}
func (h *diffCommandHandler) printDiff(a, b *diffInfo) {
	a.path = filepath.Join("a", a.path)
	b.path = filepath.Join("b", b.path)

	message := fmt.Sprintf("diff --git %s %s\n", a.path, b.path)
	
	fmt.Fprintf(h.ctx.Stdout, "%s", utilsPkg.Format("bold", message))

	h.printDiffMode(a, b)
	h.printDiffContent(a, b)
}

func (h *diffCommandHandler) printDiffMode(a, b *diffInfo) {
	message := ""
	if a.mode == b.mode {
		return
	} else if b.mode == "" {
		message = fmt.Sprintf("deleted file mode %s\n", a.mode)
	} else if a.mode == "" {
		message = fmt.Sprintf("new file mode %s\n", b.mode)
	} else {
		message = fmt.Sprintf("old mode %s\nnew mode %s\n", a.mode, b.mode)
	}

	fmt.Fprintf(h.ctx.Stdout, "%s", utilsPkg.Format("bold", message))
}
func (h *diffCommandHandler) printDiffContent(a, b *diffInfo) {
	if string(a.oid) == string(b.oid) {
		return
	}

	message := fmt.Sprintf("index %s..%s", h.short(a.oid), h.short(b.oid))
	if a.mode == b.mode && b.mode != "" {
		message += fmt.Sprintf(" %s", a.mode)
	}

	message += "\n"
	message += fmt.Sprintf("--- %s\n+++ %s\n", a.diffPath(), b.diffPath())

	fmt.Fprintf(h.ctx.Stdout, "%s", utilsPkg.Format("bold", message))

	h.printDiffHunks(a, b)
}

func (h *diffCommandHandler) printDiffHunks(a, b *diffInfo) {
	hunks := diff.DiffHunks(a.data, b.data)
	for _, hunk := range hunks {
		fmt.Fprintf(h.ctx.Stdout, "%s", hunk.ToString())
	}
}
func (h *diffCommandHandler) fromHead(path string) (*diffInfo, error) {
	entry := h.status.GetHeadTree()[path]
	return h.fromEntry(path, entry)
}
func (h *diffCommandHandler) fromIndex(path string) (*diffInfo, error) {
	entry := h.repo.Index().GetEntry(path)
	return h.fromEntry(path, entry)
}
func (h *diffCommandHandler) fromFile(path string) (*diffInfo, error) {
	content, err := h.repo.Workspace().ReadFile(path)
	if err != nil {
		return nil, err
	}

	blob := database.NewBlob(content)
	oid, err := h.repo.Database().HashObject(blob)
	if err != nil {
		return nil, err
	}

	mode := "100644"
	ok, err := h.repo.Workspace().IsExecutable(path)
	if err != nil {
		return nil, err
	}
	if ok {
		mode = "100755"
	}

	return newDiffInfo(oid, path, mode, string(content)), nil

}
func (h *diffCommandHandler) fromNothing(path string) *diffInfo {
	return newDiffInfo(make([]byte, 40), path, "", "")
}
func (h *diffCommandHandler) fromEntry(path string, entry database.Entry) (*diffInfo, error) {
	obj, err := h.repo.Database().Load(entry.GetOid())
	blob := obj.(*database.Blob)
	if err != nil {
		return nil, err
	}
	return newDiffInfo(entry.GetOid(), path, entry.GetMode(), string(blob.GetData())), nil
}
func (h *diffCommandHandler) short(oid []byte) string {
	return h.repo.Database().ShortId(oid)
}
