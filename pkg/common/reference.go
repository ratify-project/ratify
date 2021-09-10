package common

import (
	"github.com/opencontainers/go-digest"
)

type Reference struct {
	Path     string
	Digest   digest.Digest
	Tag      string
	Original string
}

func (ref Reference) String() string {
	return ref.Original
}
