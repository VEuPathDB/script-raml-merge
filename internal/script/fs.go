package script

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

func getFiles(path string) map[string]bool {
	files := make(map[string]bool)

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
		logrus.Trace(path)
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