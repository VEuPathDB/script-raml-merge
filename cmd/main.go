package main

import (
	"fmt"
	"os"

	"github.com/Foxcapades/Argonaut"
	"github.com/Foxcapades/Argonaut/pkg/argo"
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"

	"script-raml-merger/internal/script"
)

var version = "untagged-dev-version"

func main() {
	var verbose uint8
	var path string
	var exclusions []string

	logrus.SetFormatter(new(prefixed.TextFormatter))

	cli.Command().
		WithFlag(cli.ShortFlag('x').
			WithDescription("Exclude file(s).\nIf this flag is not used, the default behavior is to exclude 'library.raml'.\n\nMay be specified more than once.").
			WithArgument(cli.Argument().
				WithName("file").
				WithBinding(&exclusions).
				WithDefault("library.raml").
				Require())).
		WithFlag(cli.ShortFlag('v').
			WithDescription("Verbose process logging").
			WithCallback(func(flag argo.Flag) { verbose++ })).
		WithFlag(cli.ComboFlag('V', "version").
			WithDescription("Print tool version").
			WithCallback(func(argo.Flag) {
				fmt.Println(version)
				os.Exit(0)
			})).
		WithArgument(cli.Argument().
			WithName("RAML path").
			Require().
			WithBinding(&path)).
		MustParse(os.Args)

	if verbose == 1 {
		logrus.SetLevel(logrus.DebugLevel)
	} else if verbose > 1 {
		logrus.SetLevel(logrus.TraceLevel)
	}
	logrus.SetOutput(os.Stderr)

	fmt.Println(script.ProcessRaml(path, exclusions))
}
