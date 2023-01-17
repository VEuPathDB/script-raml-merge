package script

import (
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-yaml/v1/pkg/xyml"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"sort"
)

func ParseDTUses(file string, dt raml.DataType) raml.StringMap {
	var uses *yaml.Node

	dt.ExtraFacets().ForEach(func(k interface{}, v interface{}) {
		if k.(string) == "uses" {
			uses = v.(*yaml.Node)
		}
	})

	if uses == nil || len(uses.Content) == 0 {
		logrus.Tracef("Skipping DataType file with no imports: %s", file)
		return nil
	} else {
		logrus.Tracef("Processing imports for DataType: %s", file)
	}

	mp := raml.NewStringMap(len(uses.Content) / 2)
	_ = xyml.MapForEach(uses, func(k, v *yaml.Node) error {
		mp.Put(k.Value, v.Value)
		return nil
	})

	mp.ForEach(func(name string, v string) {
		cleanRef(name+".")("", dt)
	})

	dt.ExtraFacets().Delete("uses")

	return mp
}

func NewSortedLibraryMap(initialSize int) *SortedLibraryMap {
	return &SortedLibraryMap{
		o: make([]string, 0, initialSize),
		m: make(map[string]raml.Library, initialSize),
	}
}

type SortedLibraryMap struct {
	o []string
	m map[string]raml.Library
}

func (s *SortedLibraryMap) Put(key string, value raml.Library) {
	if _, ok := s.m[key]; !ok {
		s.o = append(s.o, key)
		sort.Strings(s.o)
	}

	s.m[key] = value
}

func (s *SortedLibraryMap) Get(key string) (value raml.Library, found bool) {
	value, found = s.m[key]
	return
}

func (s *SortedLibraryMap) Has(key string) bool {
	_, found := s.m[key]
	return found
}

func (s *SortedLibraryMap) MustGet(key string) raml.Library {
	if value, found := s.m[key]; found {
		return value
	} else {
		panic("SortedLibraryMap.MustGet called requesting a value that was not present in the map")
	}
}

func (s *SortedLibraryMap) ForEach(fn func(key string, value raml.Library)) {
	for _, key := range s.o {
		fn(key, s.m[key])
	}
}
