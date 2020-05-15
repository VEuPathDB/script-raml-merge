package main

import (
	"fmt"
	"github.com/Foxcapades/Argonaut/v0"
	"github.com/Foxcapades/Argonaut/v0/pkg/argo"
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"os"

	"script-raml-merger/internal/script"
)

var version = "untagged-dev-version"

func main() {
	var verbose bool
	var path    string

	logrus.SetFormatter(new(prefixed.TextFormatter))

	cli.NewCommand().
		Flag(cli.SFlag('v', "Verbose process logging").Bind(&verbose, false)).
		Flag(cli.SlFlag('V', "version", "Print tool version").
			OnHit(func(argo.Flag) {
				fmt.Println(version)
				os.Exit(0)
			})).
		Arg(cli.NewArg().Name("RAML path").Require().Bind(&path)).
		MustParse()

	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.SetOutput(os.Stderr)

	fmt.Println(script.ProcessRaml(path))
}
