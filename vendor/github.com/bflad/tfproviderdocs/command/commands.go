package command

import (
	"github.com/bflad/tfproviderdocs/version"
	"github.com/mitchellh/cli"
)

const CommandHelpOptionFormat = "  %s\t%s\t\n"

func Commands(ui cli.Ui) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"check": func() (cli.Command, error) {
			return &CheckCommand{
				Ui: ui,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				Version: version.GetVersion(),
				Ui:      ui,
			}, nil
		},
	}
}
