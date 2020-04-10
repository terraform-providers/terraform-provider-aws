package main

import (
	"fmt"
	"os"

	"github.com/bflad/tfproviderdocs/command"
	"github.com/bflad/tfproviderdocs/version"
	"github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"
)

const (
	Name = `tfproviderdocs`
)

func main() {
	ui := &cli.ColoredUi{
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		InfoColor:  cli.UiColorGreen,
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      colorable.NewColorableStdout(),
			ErrorWriter: colorable.NewColorableStderr(),
		},
	}

	c := &cli.CLI{
		Name:     Name,
		Version:  version.GetVersion().FullVersionNumber(true),
		Args:     os.Args[1:],
		Commands: command.Commands(ui),
	}

	exitStatus, err := c.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(exitStatus)
}
