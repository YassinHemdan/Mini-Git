package internals

import (
	internals "JIT/internals/database"
	index "JIT/internals/index"
	"os"
	"syscall"
)


/*
	The inspector here is responsible for comparing index entry against the workspace and the head tree
	It takes indexEntry, headEntry, and fileInfo as inputs, so it does not care how the entries are
	loaded or something, it just deals with the comparison and the user is responsible for providing
	these data

	we used to implement this logic in the status but we will need it in the migration, so we 
	encapsulated this logic in a separate class (Inspector) to answer for us the type of the change
	if an indexEntry
*/
type Inspector struct {
	repo *Repository
}

func newInspector(repo *Repository) *Inspector {
	return &Inspector{repo: repo}
}

// to check if the a directory is empty or not
// if there are only nested directories with no files, we consider the higher directory is empty
// if there is at least one file -> not empty
func (ins *Inspector) isTrackableFile(entryName string, entryInfo os.FileInfo) (bool, error) {
	if !entryInfo.IsDir() {
		return true, nil
	}
	queue := []string{entryName}
	for len(queue) != 0 {
		dirName := queue[0]
		queue = queue[1:]

		childEntries, err := ins.repo.Workspace().ListDir(dirName)
		if err != nil {
			return false, err
		}

		for childName, childInfo := range childEntries {
			if !childInfo.IsDir() {
				return true, nil
			}
			queue = append(queue, childName)
		}

	}

	return false, nil
}

func (ins *Inspector) checkIndexAgainstWorkspace(entry *index.IndexEntry, info os.FileInfo) (string, error) {
	if entry == nil {
		return UNTRACKED, nil
	}
	if info == nil {
		return DELETED, nil
	}

	stat := info.Sys().(*syscall.Stat_t)
	if !entry.IsMatchedStat(stat) {
		return MODIFIED, nil
	}
	if entry.IsMatchedTime(stat) {
		return "", nil
	}

	blob, err := ins.createBlob(entry.GetPathname())
	if err != nil {
		return "", err
	}
	oid, err := ins.repo.Database().HashObject(blob)
	if err != nil {
		return "", err
	}

	if string(oid) != string(entry.GetOid()) {
		return MODIFIED, nil
	}

	return "", nil
}

func (ins *Inspector) checkIndexAgainstHeadTree(indexEntry *index.IndexEntry, headEntry internals.Entry) string {
	if headEntry == nil {
		return ADDED
	}
	if indexEntry == nil {
		return DELETED
	}
	if headEntry.GetMode() != indexEntry.GetMode() || string(headEntry.GetOid()) != string(indexEntry.GetOid()) {
		return MODIFIED
	}
	return ""
}

func (ins *Inspector) createBlob(pathname string) (*internals.Blob, error) {
	data, err := ins.repo.Workspace().ReadFile(pathname)
	if err != nil {
		return nil, err
	}
	blob := internals.NewBlob(data)
	return blob, nil
}
