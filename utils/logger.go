package utils

import (
	"os"

	"github.com/sirupsen/logrus"
	formatter "gitlab.kilic.dev/libraries/go-utils/logger/formatter"
)

// LogLevelVerbose log level for the cli
var LogLevelVerbose bool

// Log Returns a new logrus logger instance.
var Log = logrus.New()

// InitiateLogger the default logger
func InitiateLogger() {
	Log.Out = os.Stdout

	Log.SetFormatter(&formatter.Formatter{
		FieldsOrder: []string{"component", "action", "category"},
	})

	// set log level
	if LogLevelVerbose {
		Log.SetLevel(logrus.DebugLevel)
	} else {
		Log.SetLevel(logrus.InfoLevel)
	}

}
