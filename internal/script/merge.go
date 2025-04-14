package script

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/Foxcapades/goop/v1/pkg/option"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rbuild"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rmeta"
	"github.com/Foxcapades/lib-go-yaml/v1/pkg/xyml"
)

const (
	logErrConflict = "Type \"%s\" is defined in more than one file:\n  %s"
	logErrFatal    = "Cannot merge RAML files"
)

func merge(files map[string]bool, types *RamlFiles) raml.Library {
	conflicts := make(TypeNameToParentFileMap, 100)

	builder := rbuild.NewLibrary()

	// Recursively resolve all imports
	for file, lib := range types.Libs.Iterator() {
		conflicts.Merge(ResolveUsesFiles(file, lib.Uses(), files, types))
	}

	types.Types.ForEach(func(file string, dt raml.DataType) {
		if mp := ParseDTUses(file, dt); mp != nil {
			conflicts.Merge(ResolveUsesFiles(file, mp, files, types))
		}
	})

	for file, lib := range types.Libs.Iterator() {
		logrus.Debugf("Processing library file: %s", file)

		dir := filepath.Dir(file)

		lib.Uses().ForEach(func(name, _ string) {
			cleanupRefs(name, lib.Types())
		})

		// Index type names to library file paths
		lib.Types().ForEach(func(name string, def raml.DataType) {
			conflicts.Append(name, file)

			if def.Kind() == rmeta.TypeInclude {
				path := fixPath(dir, strings.TrimSpace(
					strings.ReplaceAll(def.Type(), "!include", ""),
				))

				if dt, ok := types.Types.Get(path); ok {
					builder.Types().Put(name, dt)
				} else {
					builder.Types().Put(name, def)
				}
			} else {
				builder.Types().Put(name, def)
			}
		})
	}

	err := false

	for key := range conflicts {
		val := conflicts.GetFilesFor(key)

		if len(val) > 1 {
			logrus.Errorf(logErrConflict, key, strings.Join(val, "\n  "))
			err = true
		}
	}

	if err {
		logrus.Fatalf(logErrFatal)
	}

	keys := make([]string, 0, builder.Types().Len())
	builder.Types().ForEach(func(k string, _ raml.DataType) { keys = append(keys, k) })

	sort.Strings(keys)

	out := rbuild.NewLibrary()
	for _, key := range keys {
		val, _ := builder.Types().Get(key)
		out.Types().Put(key, val)
	}

	return out
}

func cleanupRefs(key string, types raml.DataTypeMap) {
	refPrefix := key + "."

	logrus.Debugf("Cleaning references for prefix \"%s\"", refPrefix)

	types.ForEach(cleanRef(refPrefix))
}

func cleanRef(prefix string) func(string, raml.DataType) {
	return func(s string, kind raml.DataType) {
		if len(s) == 0 {
			s = "undefined"
		}

		logrus.Tracef("Cleaning type: %s", s)

		if kind == nil {
			logrus.Traceln("given DataType instance was nil")
			return
		}

		if strings.HasPrefix(kind.Type(), prefix) {
			kind.OverrideType(kind.Type()[len(prefix):])
		}

		if tmp, ok := kind.(raml.ObjectType); ok {
			cleanupProps(prefix, tmp.Properties())
		}

		if tmp, ok := kind.(raml.CustomType); ok {
			cleanupProps1(prefix, tmp.ExtraFacets().GetOpt("properties"))
		}

		if tmp, ok := kind.(raml.AnyType); ok {
			cleanupProps1(prefix, tmp.ExtraFacets().GetOpt("properties"))
		}

		if tmp, ok := kind.(raml.ArrayType); ok {
			cleanRef(prefix)(prefix, tmp.Items())
		}
	}
}

func cleanupProps(key string, types raml.PropertyMap) {
	fn := cleanRef(key)
	types.ForEach(func(k string, v raml.Property) { fn(k, v) })
}

func cleanupProps1(prefix string, types option.Untyped) {
	if types.IsNil() {
		return
	}

	_ = xyml.MapForEach(types.Get().(*yaml.Node), cleanupProps2(prefix))
}

func cleanupProps2(prefix string) func(k, v *yaml.Node) error {
	return func(k, v *yaml.Node) error {
		if xyml.IsString(v) {
			if strings.HasPrefix(v.Value, prefix) {
				v.Value = v.Value[len(prefix):]
			}
			return nil
		}

		if xyml.IsMap(v) {
			return xyml.MapForEach(v, cleanupProps2(prefix))
		}

		return nil
	}
}

// fixPath attempts to correct the given possibly relative path
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
	} else if err != nil {
		logrus.Fatalf("unexpected error when attempting to stat file %s: %s", ref, err)
		panic("unreachable")
	}
	return true
}
