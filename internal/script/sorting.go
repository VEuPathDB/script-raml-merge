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

func (s *SortedLibraryBuilder) Build() SortedLibrary {
	ordered := make([]string, 0, len(s.m))
	for k := range s.m {
		ordered = append(ordered, k)
	}
	slices.Sort(ordered)

	return SortedLibrary{ordered, s.m}
}

type SortedLibrary struct {
	o []string
	m map[string]raml.Library
}

func (s *SortedLibrary) Get(key string) (value raml.Library, found bool) {
	value, found = s.m[key]
	return
}

func (s *SortedLibrary) Has(key string) bool {
	_, found := s.m[key]
	return found
}

func (s *SortedLibrary) MustGet(key string) raml.Library {
	if value, found := s.m[key]; found {
		return value
	} else {
		panic("SortedLibraryBuilder.MustGet called requesting a value that was not present in the map")
	}
}

func (s *SortedLibrary) Iterator() iter.Seq2[string, raml.Library] {
	return func(yield func(string, raml.Library) bool) {
		for _, k := range s.o {
			yield(k, s.m[k])
		}
	}
}
