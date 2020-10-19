package script

import (
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rbuild"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rmeta"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

const generated = `
################################################################################
#                                                                              #
#  DO NOT EDIT THIS FILE; IT WAS GENERATED AUTOMATICALLY.                      #
#  CHANGES MADE HERE WILL BE LOST.                                             #
#                                                                              #
################################################################################

`

func ProcessRaml(path string) string {
	stat, err := os.Stat(path)
	if err == os.ErrNotExist {
		logrus.Fatalf("Provided path does not exist: %s", path)
	} else if err != nil {
		logrus.Fatalf("Could not stat path %s: %s", path, err)
	}

	if !stat.IsDir() {
		logrus.Fatalf("Provided path is not a directory: %s", path)
	}

	paths := getFiles(path)
	types := sortingHat(paths)

	out := strings.Builder{}
	out.WriteString(rmeta.HeaderLibrary)
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
		Libs:  make(map[string]raml.Library),
		Types: make(map[string]raml.DataType),
	}

	for path := range files {
		node := new(yaml.Node)

		data, err := ioutil.ReadFile(path)
		if err != nil {
			logrus.Fatal(err)
			panic(nil)
		}

		err = yaml.Unmarshal(data, node)
		if err != nil {
			logrus.Fatal(err)
			panic(nil)
		}

		parts := strings.Split(node.HeadComment, " ")
		if len(parts) < 3 {
			logrus.Fatalf("Untyped RAML fragment: %s", path)
			panic(nil)
		}

		if parts[2] == "Library" {
			tmp := rbuild.NewLibrary()
			err = tmp.UnmarshalYAML(node)
			if err != nil {
				logrus.Fatal(err)
				panic(nil)
			}
			out.Libs[path] = tmp
		} else if parts[2] == "DataType" {
			tmp := rbuild.NewAnyDataType()
			err = tmp.UnmarshalRAML(node)
			if err != nil {
				logrus.Fatal(err)
				panic(nil)
			}
			out.Types[path] = tmp
		}
	}

	return
}

type RamlFiles struct {
	Libs  map[string]raml.Library
	Types map[string]raml.DataType
}
