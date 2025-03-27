package script

import (
	"errors"
	"os"
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
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		logrus.Fatalf("Provided path does not exist: %s", path)
	} else if err != nil {
		logrus.Fatalf("Could not stat path %s: %s", path, err)
	}

	if !stat.IsDir() {
		logrus.Fatalf("Provided path is not a directory: %s", path)
	}

	paths := getFiles(path, excluded)
	types := sortingHat(paths)

	out := strings.Builder{}
	out.WriteString("#%RAML 1.0 Library")
	out.WriteString(generated)

	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)

	if err := enc.Encode(merge(paths, types)); err != nil {
		logrus.Fatal(err)
	}

	return out.String()
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
		logrus.Fatal(err)
		panic(nil)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	dec := yaml.NewDecoder(file)
	err = dec.Decode(node)
	if err != nil {
		logrus.Fatal(err)
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
			logrus.Fatal(err)
			panic(nil)
		}
		out.Libs.Put(path, tmp)
	} else if parts[2] == "DataType" {
		logrus.Debugf("Sorted as typedef")
		tmp := rbuild.NewAnyDataType()
		err = tmp.UnmarshalRAML(node.Content[0])
		if err != nil {
			logrus.Fatal(err)
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
