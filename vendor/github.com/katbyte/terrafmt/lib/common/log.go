package common

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = createLogger()

func createLogger() *logrus.Logger {
	l := logrus.New()

	l.SetOutput(os.Stderr)

	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	l.SetFormatter(customFormatter)

	lls := os.Getenv("TERRAFMT_LOG")
	if lls == "" {
		lls = "WARN"
	}

	ll, err := logrus.ParseLevel(lls)
	if err != nil {
		l.SetLevel(logrus.TraceLevel)
		l.Errorf("defaulting to TRACE: unable to parse `TERRAFMT_LOG` into a valid log level %v", err)
	} else {
		l.SetLevel(ll)
	}

	return l
}
