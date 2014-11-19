package parser

import (
	"github.com/sylphon/build-runner/git"
)

// TODO: add template-based tagging, do away with the rest of this

/*
Tag is the interface for specifying tags for container builds.
*/
type Tag interface {
	Evaluate(top string) (result string)
}

// used for git-based tags
type tag struct {
	value string
}

/*
NewTag returns a Tag instance.  See function implementation for details on what
args to pass.
*/
func NewTag(value string) Tag {
	return &tag{value: value}
}

func (t *tag) Evaluate(top string) string {
	switch t.value {
	case "git:branch":
		return git.Branch(top)
	case "git:rev", "git:sha":
		return git.Sha(top)
	case "git:short", "git:tag":
		return git.Tag(top)
	default:
		return t.value
	}
}
