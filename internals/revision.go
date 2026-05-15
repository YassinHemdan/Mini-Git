package internals

import (
	internals "JIT/internals/database"
	"JIT/internals/utils"
	"fmt"
	"regexp"
	"strconv"
)

type InvalidObject struct {
	message string
}

func (e *InvalidObject) Error() string {
	return fmt.Sprintf("fatal: %s", e.message)
}

type node interface {
	resolve(*Revision) ([]byte, error)
}

type ref struct {
	name string
}

func newRef(name string) *ref {
	return &ref{name: name}
}
func (n *ref) resolve(context *Revision) ([]byte, error) { // the terminating experssion
	return context.readRef(n.name)
}

type parent struct {
	revNode node
}

func newParent(rev node) *parent {
	return &parent{revNode: rev}
}
func (n *parent) resolve(context *Revision) ([]byte, error) {
	oid, err := n.revNode.resolve(context)
	if err != nil {
		return oid, err
	}
	return context.commitParent(oid)
}

type ancestor struct {
	revNode node
	n       int
}

func newAncestor(rev node, n int) *ancestor {
	return &ancestor{revNode: rev, n: n}
}
func (an *ancestor) resolve(context *Revision) ([]byte, error) {
	oid, err := an.revNode.resolve(context)
	if err != nil {
		return oid, err
	}

	for i := 0; i < an.n; i++ {
		oid, err = context.commitParent(oid)
		if err != nil {
			return oid, err
		}
	}

	return oid, nil
}

var regexParent *regexp.Regexp = regexp.MustCompile(`^(.+)\^$`)
var regexAncestor *regexp.Regexp = regexp.MustCompile(`^(.+)~(\d+)$`)
var ref_aliases = map[string]string{
	"@": "HEAD",
}

type Revision struct {
	repo  *Repository
	exp   string
	query node
}

func NewRevision(repo *Repository, exp string) *Revision {
	r := &Revision{repo: repo, exp: exp, query: nil}
	r.query = r.parse(exp)

	return r
}
func (r *Revision) parse(revision string) node {
	if matched := regexParent.FindStringSubmatch(revision); matched != nil {
		rev := r.parse(matched[1])
		if rev != nil {
			return newParent(rev)
		}
		return nil
	} else if matched := regexAncestor.FindStringSubmatch(revision); matched != nil {
		rev := r.parse(matched[1])
		if rev != nil {
			n, _ := strconv.Atoi(matched[2])
			return newAncestor(rev, int(n))
		}
		return nil
	} else if r.IsValidRef(revision) {
		val, ok := ref_aliases[revision]
		if ok {
			revision = val
		}
		return newRef(revision)
	}
	return nil
}

func (r *Revision) Resolve() ([]byte, error) {
	if r.query == nil {
		return nil, &InvalidObject{message: fmt.Sprintf("Not a valid object name: '%s'", r.exp)}
	}

	oid, err := r.query.resolve(r)
	if err != nil {
		return nil, &InvalidObject{message: fmt.Sprintf("Not a valid object name: '%s'", r.exp)}
	}
	return oid, nil
}

func (r *Revision) IsValidRef(name string) bool {
	return utils.IsValidName(name)
}

func (r *Revision) commitParent(oid []byte) ([]byte, error) {
	commit, err := r.repo.Database().Load(oid)
	if err != nil {
		return nil, err
	}

	parentOid := commit.(*internals.Commit).GetParentOid()
	if len(parentOid) == 0 {
		return nil, &InvalidObject{message: fmt.Sprintf("Not a valid object name: '%s'", r.exp)}
	}
	return parentOid, nil
}

func (r *Revision) readRef(name string) ([]byte, error) {
	return r.repo.Refs().ReadRef(name)
}
