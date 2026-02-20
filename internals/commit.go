package internals

import "fmt"

type Commit struct {
	parent_oid []byte
	tree_oid   []byte
	oid        []byte
	message    string
	author     Author
	committer  Author
}

func (c *Commit) New(parent_oid, tree_oid []byte, message string, author, committer Author) error {
	c.parent_oid = parent_oid
	c.tree_oid = tree_oid
	c.message = message
	c.author = author
	c.committer = committer

	return nil
}

func (c *Commit) GetOid() []byte {
	return c.oid
}

func (c *Commit) SetOid(oid []byte) {
	c.oid = oid
}

func (c *Commit) Type() string {
	return "commit"
}
func (c *Commit) GetMessage() string {
	return c.message
}
func (c *Commit) ToString() string {
	parent := func() string {
		if c.parent_oid == nil {
			return ""
		}
		return fmt.Sprintf("parent %x\n", c.parent_oid)
	}

	// Sample Output
	// tree 1298djdhahgs89172hdga
	// parent 1298djdhahgs89172hdga
	// author yassin mohamed <ym910402@gmail.com> 2026-06blablabla
	// committer author yassin mohamed <ym910402@gmail.com> 2026-06blablabla
	//
	// Initial commit

	return fmt.Sprintf("tree %x\n%s\nauthor %s\ncommitter %s\n\n%s", c.tree_oid, parent(), c.author.ToString(), c.committer.ToString(), c.message)
}
