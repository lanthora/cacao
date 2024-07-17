package logger

import (
	"bytes"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

type logFormatter struct{}

func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] [%s] %s\n", timestamp, entry.Level, entry.Message)
	b.WriteString(msg)
	return b.Bytes(), nil
}

func init() {
	logger = logrus.New()
	logger.SetReportCaller(true)
	logger.SetFormatter(&logFormatter{})
	if value := os.Getenv("CUCURBITA_LOGLEVEL"); len(value) != 0 {
		setLogLevel(value)
	}
}

func setLogLevel(level string) {
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	}
}

func Fatal(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Info(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Debug(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
