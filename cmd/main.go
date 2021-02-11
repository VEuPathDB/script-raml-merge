package main

import (
	"fmt"
	"os"

	"github.com/Foxcapades/Argonaut/v0"
	"github.com/Foxcapades/Argonaut/v0/pkg/argo"
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"

	"script-raml-merger/internal/script"
)

var version = "untagged-dev-version"

func main() {
	var verbose uint8
	var path string

	logrus.SetFormatter(new(prefixed.TextFormatter))

	cli.NewCommand().
		Flag(cli.SFlag('v', "Verbose process logging").
			OnHit(func(flag argo.Flag) { verbose++ })).
		Flag(cli.SlFlag('V', "version", "Print tool version").
			OnHit(func(argo.Flag) {
				fmt.Println(version)
				os.Exit(0)
			})).
		Arg(cli.NewArg().Name("RAML path").Require().Bind(&path)).
		MustParse()

	if verbose == 1 {
		logrus.SetLevel(logrus.DebugLevel)
	} else if verbose > 1 {
		logrus.SetLevel(logrus.TraceLevel)
	}
	logrus.SetOutput(os.Stderr)

	fmt.Println(script.ProcessRaml(path))
}
