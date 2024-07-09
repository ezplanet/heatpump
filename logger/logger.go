package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"

	"heatpump/base"
)

var Log zerolog.Logger

func Init(level string, output string) error {
	log.Printf("Log level: %s - output: %s", level, output)
	// Set the log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}
	// Set the log output
	var err error

	// to optimise disable console writer on prod JSON could be much easier to parse
	if output == "stdout" && base.Environment != "e3" {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, NoColor: false, TimeFormat: time.RFC3339}
		Log = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()
	} else {
		// this combination logs plain text to stderr like the standard go log
		if output == "stderr" && base.Environment == "e3" {
			consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true, TimeFormat: time.RFC3339}
			Log = zerolog.New(consoleWriter).With().Timestamp().Logger()
		} else {
			// PRODUCTION
			var logfile *os.File
			if output == "stdout" {
				logfile = os.Stdout
			} else if output == "stderr" {
				logfile = os.Stderr
			} else {
				logfile, err = os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println(err)
					return err
				}
			}
			Log = zerolog.New(logfile).With().Timestamp().Logger()
		}
	}
	return err
}
