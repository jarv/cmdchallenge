package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Logger *logrus.Logger

func NewLogger() *logrus.Logger {
	level := LogLevel("info")
	logger := &logrus.Logger{
		Out:   os.Stdout,
		Level: level,
		Formatter: &prefixed.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			FullTimestamp:   true,
			ForceColors:     true,
		},
	}
	Logger = logger
	return Logger
}

func LogLevel(lvl string) logrus.Level {
	switch lvl {
	case "info":
		return logrus.InfoLevel
	case "error":
		return logrus.ErrorLevel
	default:
		panic("Not supported")
	}
}
