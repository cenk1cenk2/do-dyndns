package utils

import (
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// Log Returns a new logrus logger instance.
var Log = logrus.New()

func init() {
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	Log.Out = os.Stdout

	// You could set this to any `io.Writer` such as a file
	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err == nil {
	//  log.Out = file
	// } else {
	//  log.Info("Failed to log to file, using default stderr")
	// }

	Log.SetFormatter(&nested.Formatter{
		TimestampFormat: "20060102-15:04:05",
		TrimMessages:    true,
		HideKeys:        true,
		FieldsOrder:     []string{"component", "category"},
		NoFieldsColors:  true,
		ShowFullLevel:   true,
	})
}
