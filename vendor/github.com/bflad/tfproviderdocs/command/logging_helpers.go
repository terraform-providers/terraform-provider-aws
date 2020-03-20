package command

import (
	"flag"
	"fmt"
	"log"
	"text/tabwriter"

	"github.com/hashicorp/go-hclog"
)

const (
	LogLevelFlagHelpDefinition  = `-log-level=[TRACE|DEBUG|INFO|WARN|ERROR]`
	LogLevelFlagHelpDescription = `Log output level.`
)

func LogLevelFlag(flagSet *flag.FlagSet, varToSave *string) {
	flagSet.StringVar(varToSave, "log-level", "INFO", "")
}

func LogLevelFlagHelp(w *tabwriter.Writer) {
	fmt.Fprintf(w, CommandHelpOptionFormat, LogLevelFlagHelpDefinition, LogLevelFlagHelpDescription)
}

func ConfigureLogging(loggerName string, logLevel string) {
	loggerOptions := &hclog.LoggerOptions{
		Name:  loggerName,
		Level: hclog.LevelFromString(logLevel),
	}
	logger := hclog.New(loggerOptions)

	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}
