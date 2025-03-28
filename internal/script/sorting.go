package script

import (
	"iter"
	"slices"

	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
)

func NewSortedLibraryBuilder(initialSize int) *SortedLibraryBuilder {
	return &SortedLibraryBuilder{
		m: make(map[string]raml.Library, initialSize),
	}
}

type SortedLibraryBuilder struct {
	m map[string]raml.Library
}

func (s *SortedLibraryBuilder) Put(key string, value raml.Library) {
	s.m[key] = value
}

func (s *SortedLibraryBuilder) Get(key string) (raml.Library, bool) {
	out, ok := s.m[key]
	return out, ok
}

func (s *SortedLibraryBuilder) Iterator() iter.Seq2[string, raml.Library] {
	ordered := make([]string, 0, len(s.m))
	for k := range s.m {
		ordered = append(ordered, k)
	}

	slices.Sort(ordered)

	return func(yield func(string, raml.Library) bool) {
		for _, k := range ordered {
			yield(k, s.m[k])
		}
	}
}
