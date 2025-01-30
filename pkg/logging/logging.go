package logging

import (
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

var (
	logger     *logrus.Logger
	loggerOnce sync.Once
)

func Logger() *logrus.Logger {
	if logger == nil {
		logger = NewNoOpLogger()
	}
	return logger
}

func InitLogger(level *string) {
	loggerOnce.Do(func() {
		logger = logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
			ForceColors:      true,
			DisableColors:    false,
		})
		switch *level {
		case "trace":
			logger.SetLevel(logrus.TraceLevel)
		case "debug":
			logger.SetLevel(logrus.DebugLevel)
		case "info":
			logger.SetLevel(logrus.InfoLevel)
		case "warn":
			logger.SetLevel(logrus.WarnLevel)
		case "error":
			logger.SetLevel(logrus.ErrorLevel)
		case "fatal":
			logger.SetLevel(logrus.FatalLevel)
		default:
			logger.SetLevel(logrus.WarnLevel) // Default to the warn level
		}
	})
}

// SetLogger allows users to provide their own Logrus logging.
func SetLogger(userLogger *logrus.Logger) {
	loggerOnce.Do(func() {
		logger = userLogger
	})
}

func NewNoOpLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard all logs
	return logger
}
