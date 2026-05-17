package internals

import (
	internals "JIT/internals/database"
	"JIT/internals/utils"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// const (
// 	COMMIT = "commit"
// )

type HintedError struct {
	Message string
	Hints   []string
}

type InvalidObject struct {
	message string
}

func (e *InvalidObject) Error() string {
	return fmt.Sprintf("%s", e.message)
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

const (
	REVISION_COMMIT = "commit"
)

type Revision struct {
	repo   *Repository
	exp    string
	query  node
	errors []*HintedError
}

func NewRevision(repo *Repository, exp string) *Revision {
	r := &Revision{repo: repo, exp: exp, query: nil, errors: make([]*HintedError, 0)}
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

func (r *Revision) Resolve(objType string) ([]byte, error) {

	/*
		The Resolve here responsible for:
		1- running query.resolve(r {the current revision}) as this should return for us the matching
			oid for the current node (a ref node)
			note that we already called r.Parese in the initialization of the revision,
				so the query here is just a ref type node returned for us from the r.Parse(exp string)
		2- returning an InvalidObject type error if the following happened:
			- the query is nil and that means the r.Parse() was not successfull
			- an error or an empty oid returned after calling r.query.resolve(r)
	*/

	if r.query == nil {
		return nil, &InvalidObject{message: fmt.Sprintf("Not a valid object name: '%s'.", r.exp)}
	}

	oid, err := r.query.resolve(r)
	if err != nil || oid == nil {
		return nil, &InvalidObject{message: fmt.Sprintf("Not a valid object name: '%s'.", r.exp)}
	}

	if objType != "" && !r.checkObjectType(oid, objType) {
		return nil, &InvalidObject{message: fmt.Sprintf("Not a valid branch point: '%s'.", r.exp)}
	}
	return oid, nil
}

func (r *Revision) IsValidRef(name string) bool {
	return utils.IsValidName(name)
}

func (r *Revision) checkObjectType(oid []byte, objType string) bool {
	obj, _ := r.repo.Database().Load(oid)
	if obj.Type() != objType {
		message := fmt.Sprintf("error: object %x is a %s, not a commit", oid, obj.Type())
		r.errors = append(r.errors, &HintedError{Message: message, Hints: make([]string, 0)})

		return false
	}
	return true
}
func (r *Revision) commitParent(oid []byte) ([]byte, error) {
	// if an object does not exist, the oid will be nil, so we wanna make sure we don't load it
	if oid == nil {
		return nil, nil
	}
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
	oid, err := r.repo.Refs().ReadRef(name)
	if err != nil {
		return nil, err
	}

	// we found a matched oid by a reference name
	if oid != nil {
		return oid, nil
	}

	// if we did not find an oid, try to search on the prefixs of object ids
	candidates, err := r.repo.Database().PrefixMatch(name)

	// if size of matched > 1, there is an ambiguity
	if len(candidates) > 1 {
		r.logAmbiguousSha1(name, candidates)
	} else if len(candidates) == 1 {
		return hex.DecodeString(strings.TrimSpace(candidates[0]))
	}

	return nil, nil
}


/*
	We can in the future consider the following:
		what if only one of the candidates is a commit object?
		we can consider this a valid object name and resolve using this commitObject
		
		we can only consider an ambiguity if more than one candidate is a commit object
*/
func (r *Revision) logAmbiguousSha1(name string, candidates []string) {
	// main error message
	message := fmt.Sprintf("error: short object ID %s is ambiguous", name)

	hintMessages := []string{"The candidates are:"}

	objectMessages := (func(oids []string) []string {
		hints := make([]string, 0)

		for _, oid := range oids {
			oidBytes, _ := hex.DecodeString(oid)
			object, _ := r.repo.Database().Load(oidBytes)

			info := fmt.Sprintf("  %s %s", r.repo.Database().ShortId(object.GetOid()), object.Type())

			if object.Type() == "commit" {
				commit := object.(*internals.Commit)
				info += fmt.Sprintf("  %s - %s", commit.Author().ShortDate(), commit.GetMessage())
			}

			hints = append(hints, info)
		}

		return hints
	})(candidates)

	hintMessages = append(hintMessages, objectMessages...)

	r.errors = append(r.errors, &HintedError{Message: message, Hints: hintMessages})
}

func (r *Revision) HintedErrors() []*HintedError {
	return r.errors
}
