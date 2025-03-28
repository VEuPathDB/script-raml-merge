package script

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type FileIndex map[string]bool

// getFiles returns all the raml files present in the given directory, and its
// subdirectories.
func getFiles(path string, exclusions []string) FileIndex {
	files := make(FileIndex)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		logrus.Tracef("Examining path %s", path)

		if isExcluded(path, exclusions) {
			logrus.Debugf("Skipping %s because it is marked as excluded", path)

			// If the path matched a directory, skip it entirely.
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// If we are currently on a directory, bail here, the next element will be
		// the first file in the directory.
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".raml") {
			logrus.Tracef("Skipping non-raml file %s", path)
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

func isExcluded(name string, exclusions []string) bool {
	for _, x := range exclusions {
		if name == x {
			return true
		}
	}

	return false
}
