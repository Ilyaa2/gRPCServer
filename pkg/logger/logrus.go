package logger

import (
	"github.com/sirupsen/logrus"
	"io"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger(writer io.Writer, level string) (Logger, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(writer)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	log.SetLevel(lvl)
	return &LogrusLogger{logger: log}, nil
}

func (l *LogrusLogger) Debug(msg string, params map[string]interface{}) {
	l.logger.WithFields(params).Debug(msg)
}

func (l *LogrusLogger) Info(msg string, params map[string]interface{}) {
	l.logger.WithFields(params).Info(msg)
}

func (l *LogrusLogger) Warn(msg string, params map[string]interface{}) {
	l.logger.WithFields(params).Warn(msg)
}

func (l *LogrusLogger) Error(msg string, params map[string]interface{}) {
	l.logger.WithFields(params).Error(msg)
}
