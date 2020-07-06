package script

import (
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml/rbuild"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	logErrConflict = "Type \"%s\" is defined in more than one file:\n  %s"
	logErrFatal    = "Cannot merge RAML files"
)

func merge(files map[string]bool, libs map[string]raml.Library) raml.Library {
	typeToFile := make(map[string][]string)

	out := rbuild.NewLibrary()

	for file, lib := range libs {
		dir := filepath.Dir(file)

		lib.Uses().ForEach(func(name, ref string) {
			ref = fixPath(dir, ref)

			if _, ok := files[ref]; ok {
				cleanupRefs(name, lib.Types())
			} else {
				out.Uses().Put(name, ref)
			}
		})

		lib.Types().ForEach(func(name string, def raml.DataType) {
			if _, ok := typeToFile[name]; ok {
				typeToFile[name] = append(typeToFile[name], file)
			} else {
				typeToFile[name] = []string{file}
			}

			out.Types().Put(name, def)
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

func cleanRef(full string) func(string, raml.DataType) {
	return func(_ string, kind raml.DataType) {
		if strings.HasPrefix(kind.Type(), full) {
			kind.OverrideType(kind.Type()[len(full):])
		}

		if tmp, ok := kind.(raml.ObjectType); ok {
			cleanupProps(full, tmp.Properties())
		}

		if tmp, ok := kind.(raml.ArrayType); ok {
			cleanRef(full)(full, tmp.Items())
		}
	}
}

func fixPath(dir, file string) string {
	path := path.Join(dir, file)

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
