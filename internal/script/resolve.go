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

// TypeNameToParentFileMap is a mapping of type names to parent files in which
// those type names appear.
//
// This is used for type definition conflict detection, i.e. detecting when more
// than one type is defined with the same name in the scope of the RAML files.
//
// Example:
//   {
//     "MyRamlType": {
//       "file1.raml": true,
//     },
//     "MyOtherRamlType": {
//       "file1.raml": true,
//       "file2.raml": true,
//     }
//   }
type TypeNameToParentFileMap map[string]map[string]bool

// Append takes the given RAML type name and parent file and appends the two to
// the mapping of type names to source files.
//
// Example:
//   Original State:
//   {
//     "Type1": {
//       "file1.raml": true
//     }
//   }
//
//   Input 1:
//   t.Append("Type1", "file2.raml")
//
//   Resulting State 1:
//   {
//     "Type1": {
//       "file1.raml": true,
//       "file2.raml": true,
//     }
//   }
//
//   Input 2:
//   t.Append("Type2", "file1.raml")
//
//   Resulting State 2:
//   {
//     "Type1": {
//       "file1.raml": true,
//       "file2.raml": true,
//     },
//     "Type2": {
//       "file1.raml": true,
//     }
//   }
func (t TypeNameToParentFileMap) Append(ramlTypeName, parentFile string) {
	if _, ok := t[ramlTypeName]; ok {
		t[ramlTypeName][parentFile] = true
	} else {
		t[ramlTypeName] = map[string]bool{parentFile: true}
	}
}

// Merge merges the contents of the given other TypeNameToParentFileMap into
// the current TypeNameToParentFileMap.
//
// Example:
//   Input 1:
//   {
//     "Type1": {
//       "file1.raml": true,
//     }
//   }
//
//   Input 2:
//   {
//     "Type1": {
//       "file1.raml": true,
//       "file2.raml": true,
//     }
//     "Type2": {
//       "file1.raml": true
//     }
//   }
//
//   Resulting State
//   {
//     "Type1": {
//       "file1.raml": true,
//       "file2.raml": true,
//     },
//     "Type2": {
//       "file1.raml": true,
//     }
//   }
func (t TypeNameToParentFileMap) Merge(o TypeNameToParentFileMap) {
	for name, files := range o {
		for file := range files {
			t.Append(name, file)
		}
	}
}

// GetFilesFor returns a slice of filenames for files that contain a RAML type
// with the given name.
func (t TypeNameToParentFileMap) GetFilesFor(typeName string) (out []string) {
	if mp, ok := t[typeName]; ok {
		out = make([]string, 0, len(mp))
		for file := range mp {
			out = append(out, file)
		}
	}

	return
}

// ResolveUsesFiles recursively ensures that every entry in the given RAML
// 'uses' map is a known file that exists in the given RAML caches.
//
// If the reference is to a file that did not previously exist in the given
// caches, the path will be resolved and added to the caches.
func ResolveUsesFiles(
	dir string,
	uses raml.StringMap,
	files map[string]bool,
	types *RamlFiles,
) TypeNameToParentFileMap {
	out := make(TypeNameToParentFileMap, uses.Len())

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
			types.Types.Put(file, dt)
			if mp := ParseDTUses(file, dt); mp != nil {
				out.Merge(ResolveUsesFiles(filepath.Dir(file), mp, files, types))
			}
		} else if lib != nil {
			types.Libs.Put(file, lib)
			lib.Types().ForEach(func(name string, _ raml.DataType) {
				out.Append(name, file)
			})

			out.Merge(ResolveUsesFiles(filepath.Dir(file), lib.Uses(), files, types))
		}
	})

	return out
}

func resolve(file, dir string, files *RamlFiles) (raml.Library, raml.DataType) {
	if file[:4] == "http" {
		return resolveRemote(file)
	}

	file = fixPath(dir, file)

	if v, ok := files.Libs.Get(file); ok {
		return v, nil
	}

	if v, ok := files.Types.Get(file); ok {
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
