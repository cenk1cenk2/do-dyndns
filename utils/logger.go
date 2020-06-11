package utils

import (
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// LogLevelVerbose log level for the cli
var LogLevelVerbose bool

// Log Returns a new logrus logger instance.
var Log = logrus.New()

// InitiateLogger the default logger
func InitiateLogger() {
	Log.Out = os.Stdout

	Log.SetFormatter(&nested.Formatter{
		TimestampFormat: "20060102-15:04:05",
		TrimMessages:    true,
		HideKeys:        true,
		FieldsOrder:     []string{"component", "action", "category"},
		NoFieldsColors:  false,
		ShowFullLevel:   true,
	})

	// set log level
	if LogLevelVerbose {
		Log.SetLevel(logrus.DebugLevel)
	} else {
		Log.SetLevel(logrus.InfoLevel)
	}

}
