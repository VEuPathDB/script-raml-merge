package script

import (
	"github.com/Foxcapades/lib-go-raml-types/v0/pkg/raml"
	"github.com/sirupsen/logrus"
	"path"
	"path/filepath"
	"strings"
)

const (
	logScanFile    = "Scanning File"
	logScanType    = "Scanning Type Definition"
	logRefs        = `Cleaning up refs to import alias "%s"`
	logErrConflict = "Type \"%s\" is defined in more than one file:\n  %s"
	logErrFatal    = "Cannot merge RAML files"
	logRefPath     = "Changed import path:\n  From: %s\n  To:   %s"
)

func merge(files map[string]bool, libs map[string]*raml.Library) *raml.Library {
	log := logrus.WithField("method", "merge")
	typeToFile := make(map[string][]string)

	out := new(raml.Library)
	out.Uses = make(map[string]string)
	out.Types = make(map[string]*raml.TypeDef)

	for file, lib := range libs {
		log = log.WithField("file", filepath.Base(file))
		log.Debug(logScanFile)

		for name, ref := range lib.Uses {
			upd := path.Join(filepath.Dir(file), ref)

			log.Tracef(logRefPath, ref, upd)

			if _, ok := files[upd]; ok {
				log.Debugf(logRefs, name)

				cleanup(name, file, lib.Types)
			} else {
				out.Uses[name] = upd
			}
		}

		for name, def := range lib.Types {
			l2 := log.WithField("type", name)
			l2.Debug(logScanType)

			if _, ok := typeToFile[name]; ok {
				typeToFile[name] = append(typeToFile[name], file)
			} else {
				typeToFile[name] = []string{file}
			}

			out.Types[name] = def
		}
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

func cleanup(key, file string, types map[string]*raml.TypeDef) {
	log := logrus.WithFields(logrus.Fields{
		"file":   path.Base(file),
		"method": "cleanup",
		"key":    key,
	})

	full := key + "."
	for _, kind := range types {
		log.Trace("Processing type ", kind.GetType())

		if strings.HasPrefix(kind.GetType(), full) {
			log.Debug("Correcting ", kind.GetType())
			kind.SetType(kind.GetType()[len(full):])
		}

		raw := kind.GetRawObject()

		if tmp, ok := raw.(*raml.Object); ok {
			cleanup(key, file, tmp.Properties)
		}
		if tmp, ok := raw.(*raml.Array); ok {
			cleanup(key, file, map[string]*raml.TypeDef{"tmp": tmp.Items})
		}
	}

}
