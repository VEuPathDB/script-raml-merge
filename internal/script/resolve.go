package script

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/Foxcapades/Go-ChainRequest/simple"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rbuild"
)

const (
	headerPrefix = "#%RAML 1.0 "
)

const (
	errParseYamlFail = "Failed to parse contents of file as YAML: "
	errParseRamlFail = "Failed to parse contents of file as RAML: "
	errNoHeader      = "No RAML header found in file "
	errBadHeader     = "Invalid RAML header found in file "
	errBadType       = "Unrecognized or invalid RAML fragment type for file "
	errReadFail      = "Failed to read file contents from "
)

// TypeToFiles is a mapping of type names to files in which that type name
// appears.
//
// This is used for type conflict detection.
type TypeToFiles map[string]map[string]bool

func (t TypeToFiles) Append(name, file string) {
	if _, ok := t[name]; ok {
		t[name][file] = true
	} else {
		t[name] = map[string]bool{file: true}
	}
}

func (t TypeToFiles) Merge(o TypeToFiles) {
	for name, files := range o {
		for file := range files {
			t.Append(name, file)
		}
	}
}

func (t TypeToFiles) GetFilesFor(typeName string) (out []string) {
	if mp, ok := t[typeName]; ok {
		out = make([]string, 0, len(mp))
		for file := range mp {
			out = append(out, file)
		}
	}

	return
}

// ResolveUses recursively ensures that every entry in the given RAML uses map
// exists in the given caches.
//
// If the reference is to a file that did not previously exist in the given
// caches, the path will be resolve and added to the caches.
func ResolveUses(
	dir string,
	uses raml.StringMap,
	files map[string]bool,
	types *RamlFiles,
) TypeToFiles {
	out := make(TypeToFiles, uses.Len())

	uses.ForEach(func(key string, path string) {
		file := fixPath(dir, path)

		// If we already know about it, there's nothing to resolve.
		if files[file] {
			return
		}

		lib, dt := resolve(path, dir, types)

		// Record resolved file
		if path[:4] == "http" {
			files[path] = true
		} else {
			files[file] = true
		}

		// Record type
		if dt != nil {
			types.Types[file] = dt
			if mp := ParseDTUses(file, dt); mp != nil {
				out.Merge(ResolveUses(filepath.Dir(file), mp, files, types))
			}
		} else if lib != nil {
			types.Libs[file] = lib
			lib.Types().ForEach(func(name string, _ raml.DataType) {
				out.Append(name, file)
			})

			out.Merge(ResolveUses(filepath.Dir(file), lib.Uses(), files, types))
		}
	})

	return out
}

func resolve(file, dir string, files *RamlFiles) (raml.Library, raml.DataType) {
	if file[:4] == "http" {
		return resolveRemote(file)
	}

	file = fixPath(dir, file)

	if v, ok := files.Libs[file]; ok {
		return v, nil
	}

	if v, ok := files.Types[file]; ok {
		return nil, v
	}

	return resolveLocal(file)
}

func resolveLocal(file string) (raml.Library, raml.DataType) {
	tmp, err := ioutil.ReadFile(file)
	if err != nil {
		logrus.Error(err)
		logrus.Fatal(errReadFail)
		panic(nil)
	}

	return resolveParse(tmp, file)
}

func resolveRemote(url string) (raml.Library, raml.DataType) {
	res := simple.GetRequest(url).Submit()
	defer res.Close()

	tmp, err := res.GetBody()
	if err != nil {
		logrus.Error(err)
		logrus.Fatal(errReadFail + url)
		panic(nil)
	}

	return resolveParse(tmp, url)
}

func resolveParse(in []byte, ref string) (lib raml.Library, dt raml.DataType) {
	node := new(yaml.Node)
	if err := yaml.Unmarshal(in, node); err != nil {
		logrus.Error(err)
		logrus.Fatal(errParseYamlFail + ref)
		panic(nil)
	}

	if node.HeadComment == "" {
		logrus.Fatal(errNoHeader + ref)
		panic(nil)
	} else if !strings.HasPrefix(node.HeadComment, headerPrefix) {
		logrus.Fatal(errBadHeader + ref)
		panic(nil)
	}

	kind := node.HeadComment[len(headerPrefix):]

	if len(node.Content) == 0 {
		return nil, nil
	}

	switch kind {
	case "DataType":
		tmp := rbuild.NewAnyDataType()
		if err := tmp.UnmarshalRAML(node.Content[0]); err != nil {
			logrus.Error(err)
			logrus.Fatal(errParseRamlFail + ref)
			panic(nil)
		}
		dt = tmp

	case "Library":
		lib = rbuild.NewLibrary()
		if err := lib.UnmarshalYAML(node.Content[0]); err != nil {
			logrus.Error(err)
			logrus.Fatal(errParseRamlFail + ref)
			panic(nil)
		}

	default:
		logrus.Fatal(errBadType + kind)
		panic(nil)
	}

	return
}
