package internals

import (
	"JIT/utils"
	"encoding/hex"
	"fmt"
	"strings"
)

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

func ParseCommit(scanner *utils.SmartScanner) Object {
	headers := make(map[string]string)
	scanner.SplitByDelim(' ')
	for scanner.Scan() {
		headerType := scanner.Text()

		if headerType == "" || headerType == "\n" {
			break
		}
		scanner.SplitByDelim('\n')
		scanner.Scan()
		headerContent := scanner.Text()

		headers[headerType] = headerContent

		scanner.SplitByDelim(' ')
	}
	scanner.ScanRest()
	message := scanner.Text()

	scanner.NewReader(strings.NewReader(headers["author"]))
	author := ParseAuthor(scanner)

	scanner.NewReader(strings.NewReader(headers["committer"]))
	committer := ParseAuthor(scanner)

	parentOid, _ := hex.DecodeString(headers["parent"])
	treeOid, _ := hex.DecodeString(headers["tree"])
	return &Commit{
		parent_oid: parentOid,
		tree_oid:   treeOid,
		author:     *author,
		committer:  *committer,
		message:    message,
	}

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
func (c *Commit) GetTreeOid() []byte {
	return c.tree_oid
}
func (c *Commit) GetParentOid() []byte {
	return c.parent_oid
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
