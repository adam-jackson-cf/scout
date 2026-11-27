package tooling

import (
	"strings"

	"scout/tools/gate/check"
)

type revisionPair struct {
	base, head string
}

func revisions(r *check.Run) (revisionPair, bool, error) {
	if r.Mode == check.Commit {
		return revisionPair{base: "HEAD", head: ""}, true, nil
	}
	if r.Base == "" {
		return revisionPair{}, false, nil
	}
	mb, err := r.Git("merge-base", r.Base, "HEAD")
	if err != nil {
		return revisionPair{}, false, err
	}
	return revisionPair{base: strings.TrimSpace(mb), head: "HEAD"}, true, nil
}

func showAt(r *check.Run, rev, rel string) (string, bool, error) {
	spec := rev + ":" + rel
	if _, err := r.Git("cat-file", "-e", spec); err != nil {
		return "", false, nil
	}
	out, err := r.Git("show", spec)
	return out, err == nil, err
}
