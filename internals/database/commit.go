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
	parent := (func() string {
		if len(c.parent_oid) == 0 {
			return ""
		}
		return fmt.Sprintf("parent %x\n", c.parent_oid)
	})()
	return fmt.Sprintf("tree %x\n%sauthor %s\ncommitter %s\n\n%s\n", c.tree_oid, parent, c.author.ToString(), c.committer.ToString(), c.message)
}
