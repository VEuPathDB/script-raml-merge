package script

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Foxcapades/lib-go-raml/v0/pkg/raml"
	"github.com/Foxcapades/lib-go-yaml/v1/pkg/xyml"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func ParseDTUses(file string, dt raml.DataType) raml.StringMap {
	var uses *yaml.Node

	dt.ExtraFacets().ForEach(func(k interface{}, v interface{}) {
		if k.(string) == "uses" {
			uses = v.(*yaml.Node)
		}
	})

	if uses == nil || len(uses.Content) == 0 {
		logrus.Tracef("Skipping DataType file with no imports: %s", file)
		return nil
	} else {
		logrus.Tracef("Processing imports for DataType: %s", file)
	}

	mp := raml.NewStringMap(len(uses.Content) / 2)
	_ = xyml.MapForEach(uses, func(k, v *yaml.Node) error {
		mp.Put(k.Value, v.Value)
		return nil
	})

	mp.ForEach(func(name string, v string) {
		cleanRef(name+".")("", dt)
	})

	dt.ExtraFacets().Delete("uses")

	return mp
}

func RelativePath(path, relativeTo string) string {
	if out, err := filepath.Rel(relativeTo, path); err != nil {
		logrus.Fatalf("failed to relativise path %s against path %s: %s", path, relativeTo, err)
		panic("unreachable")
	} else {
		return out
	}
}

func CWD() string {
	if dir, err := os.Getwd(); err != nil {
		logrus.Fatalf("failed to get current working directory: %s", err)
		panic("unreachable")
	} else {
		return dir
	}
}

func MustStat(path string) os.FileInfo {
	stat, err := os.Stat(path)

	if errors.Is(err, os.ErrNotExist) {
		logrus.Fatalf("Provided path does not exist: %s", path)
		panic("unreachable")
	} else if err != nil {
		logrus.Fatalf("Could not stat path %s: %s", path, err)
		panic("unreachable")
	}

	return stat
}
