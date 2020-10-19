package script

import (
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rbuild"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rmeta"
	"github.com/Foxcapades/lib-go-yaml/v1/pkg/xyml"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

const (
	logErrConflict = "Type \"%s\" is defined in more than one file:\n  %s"
	logErrFatal    = "Cannot merge RAML files"
)

func merge(files map[string]bool, types *RamlFiles) raml.Library {
	typeToFile := make(map[string][]string)

	out := rbuild.NewLibrary()

	for file, lib := range types.Libs {
		dir := filepath.Dir(file)

		lib.Uses().ForEach(func(name, ref string) {
			ref = fixPath(dir, ref)

			if _, ok := files[ref]; ok {
				cleanupRefs(name, lib.Types())
			} else {
				out.Uses().Put(name, ref)
			}
		})

		// Index type names to library file paths
		lib.Types().ForEach(func(name string, def raml.DataType) {
			if _, ok := typeToFile[name]; ok {
				typeToFile[name] = append(typeToFile[name], file)
			} else {
				typeToFile[name] = []string{file}
			}

			if def.Kind() == rmeta.TypeInclude {
				path := fixPath(dir, strings.TrimSpace(
					strings.ReplaceAll(def.Type(), "!include", ""),
				))

				if dt, ok := types.Types[path]; ok {
					out.Types().Put(name, dt)
				} else {
					out.Types().Put(name, def)
				}
			} else {
				out.Types().Put(name, def)
			}
		})
	}

	for _, dt := range types.Types {
		// check if the datatype definition is doing it's own imports
		if !dt.ExtraFacets().Has("uses") {
			continue
		}

		rawRef, _ := dt.ExtraFacets().Get("uses")
		node := rawRef.(*yaml.Node)
		_ = xyml.MapForEach(node, func(k, v *yaml.Node) error {
			name := k.Value
			ref  := v.Value

			if _, ok := files[ref]; ok {
				cleanRef(name + ".")("", dt)
			} else {
				out.Uses().Put(name, ref)
			}

			return nil
		})
	}

	err := false

	for key, val := range typeToFile {
		if len(val) > 1 {
			logrus.Errorf(logErrConflict, key, strings.Join(val, "\n  "))
			err = true
		}
	}

	if err {
		logrus.Fatalf(logErrFatal)
	}

	return out
}

func cleanupRefs(key string, types raml.DataTypeMap) {
	types.ForEach(cleanRef(key + "."))
}

func cleanupProps(key string, types raml.PropertyMap) {
	fn := cleanRef(key)
	types.ForEach(func(k string, v raml.Property) { fn(k, v) })
}

func cleanRef(prefix string) func(string, raml.DataType) {
	return func(_ string, kind raml.DataType) {
		if strings.HasPrefix(kind.Type(), prefix) {
			kind.OverrideType(kind.Type()[len(prefix):])
		}

		if tmp, ok := kind.(raml.ObjectType); ok {
			cleanupProps(prefix, tmp.Properties())
		}

		if tmp, ok := kind.(raml.ArrayType); ok {
			cleanRef(prefix)(prefix, tmp.Items())
		}
	}
}

func fixPath(dir, file string) string {
	path := filepath.Clean(filepath.Join(dir, file))

	if fileExists(path) {
		return path
	}

	return file
}

func fileExists(ref string) bool {
	_, err := os.Stat(ref)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
