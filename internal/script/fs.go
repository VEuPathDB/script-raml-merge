package script

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type FileIndex map[string] bool

func getFiles(path string) FileIndex {
	files := make(FileIndex)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".raml") {
			logrus.Trace("Skipping non-file path", path)
			return nil
		}

		if filepath.Base(path) == "library.raml" {
			logrus.Debugf("Skipping %s", path)
			return nil
		}

		files[path] = true
		return nil
	})

	if err != nil {
		logrus.Fatalf("Could not process path %s: %s", path, err)
	}

	return files
}
