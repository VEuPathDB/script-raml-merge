package script

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rbuild"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const generated = `
################################################################################
#                                                                              #
#  DO NOT EDIT THIS FILE; IT WAS GENERATED AUTOMATICALLY.                      #
#  CHANGES MADE HERE WILL BE LOST.                                             #
#                                                                              #
################################################################################

`

func ProcessRaml(path string, excluded []string) string {
	stat := MustStat(path)

	if !stat.IsDir() {
		logrus.Fatalf("Provided path is not a directory: %s", path)
	}

	normalizeExclusions(path, excluded)

	paths := getFiles(path, excluded)
	types := sortingHat(paths)

	out := strings.Builder{}
	out.WriteString("#%RAML 1.0 Library")
	out.WriteString(generated)

	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)

	if err := enc.Encode(merge(paths, types)); err != nil {
		logrus.Fatalf("Error while encoding the merged raml files: %s", err)
	}

	return out.String()
}

func normalizeExclusions(path string, exclusions []string) {
	for i, p := range exclusions {
		if filepath.IsAbs(p) {
			continue
		}

		if fileExists(p) {
			continue
		}

		relativeToSchemaDir := filepath.Join(path, p)
		if fileExists(relativeToSchemaDir) {
			logrus.Debugf("correcting path %s to %s", p, relativeToSchemaDir)
			exclusions[i] = relativeToSchemaDir
		}
	}
}

func sortingHat(files FileIndex) (out *RamlFiles) {
	out = &RamlFiles{
		Libs:  NewSortedLibraryBuilder(len(files)),
		Types: raml.NewDataTypeMap(32),
	}

	for path := range files {
		sortFile(path, out)
	}

	return
}

func sortFile(path string, out *RamlFiles) {
	logrus.Tracef("Sorting raml type for file %s", path)
	node := new(yaml.Node)

	file, err := os.Open(path)
	if err != nil {
		logrus.Fatalf("Error while attempting to open file %s: %s", path, err)
		panic(nil)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	dec := yaml.NewDecoder(file)
	err = dec.Decode(node)
	if err != nil {
		logrus.Fatalf("Error while attempting to decode file %s as YAML: %s", path, err)
		panic(nil)
	}

	if len(node.Content) == 0 {
		logrus.Debugf("Skipping empty fragment: %s", path)
		return
	}

	parts := strings.Fields(node.HeadComment)
	logrus.Debug(node)
	if len(parts) < 3 {
		logrus.Tracef("Header: %s", parts)
		logrus.Fatalf("Untyped RAML fragment: %s", path)
		panic(nil)
	}

	if parts[2] == "Library" {
		logrus.Debugf("Sorted as library")
		tmp := rbuild.NewLibrary()
		err = tmp.UnmarshalYAML(node.Content[0])
		if err != nil {
			logrus.Fatalf("Error while attempting to parse YAML file %s: %s", path, err)
			panic(nil)
		}
		out.Libs.Put(path, tmp)
	} else if parts[2] == "DataType" {
		logrus.Debugf("Sorted as typedef")
		tmp := rbuild.NewAnyDataType()
		err = tmp.UnmarshalRAML(node.Content[0])
		if err != nil {
			logrus.Fatalf("Error while attempting to parse RAML file %s: %s", path, err)
			panic(nil)
		}
		out.Types.Put(path, tmp)
	} else {
		logrus.Debugf("Failed to sort RAML file")
		logrus.Debugf("Header: %s", parts)
	}
}

type RamlFiles struct {
	Libs  *SortedLibraryBuilder
	Types raml.DataTypeMap
}
