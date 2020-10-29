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

		logrus.Tracef("Examining path %s", path)

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".raml") {
			logrus.Debugf("Skipping non-file path %s", path)
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
