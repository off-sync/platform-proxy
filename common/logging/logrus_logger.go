package logging

import (
	"github.com/sirupsen/logrus"
	"github.com/off-sync/platform-proxy/app/interfaces"
)

// NewFromLogrus creates a Logger from the provided logrus logger.
func NewFromLogrus(log *logrus.Logger) interfaces.Logger {
	return &logrusLogger{
		log: log,
	}
}

type logrusLogger struct {
	log *logrus.Logger
}

func (l *logrusLogger) Debug(msg string) { l.log.Debug(msg) }
func (l *logrusLogger) Info(msg string)  { l.log.Info(msg) }
func (l *logrusLogger) Warn(msg string)  { l.log.Warn(msg) }
func (l *logrusLogger) Error(msg string) { l.log.Error(msg) }
func (l *logrusLogger) Fatal(msg string) { l.log.Fatal(msg) }

func (l *logrusLogger) WithField(key string, value interface{}) interfaces.Logger {
	return &logrusEntryLogger{entry: l.log.WithField(key, value)}
}

func (l *logrusLogger) WithError(err error) interfaces.Logger {
	return &logrusEntryLogger{entry: l.log.WithError(err)}
}

type logrusEntryLogger struct {
	entry *logrus.Entry
}

func (l *logrusEntryLogger) Debug(msg string) { l.entry.Debug(msg) }
func (l *logrusEntryLogger) Info(msg string)  { l.entry.Info(msg) }
func (l *logrusEntryLogger) Warn(msg string)  { l.entry.Warn(msg) }
func (l *logrusEntryLogger) Error(msg string) { l.entry.Error(msg) }
func (l *logrusEntryLogger) Fatal(msg string) { l.entry.Fatal(msg) }

func (l *logrusEntryLogger) WithField(key string, value interface{}) interfaces.Logger {
	return &logrusEntryLogger{entry: l.entry.WithField(key, value)}
}

func (l *logrusEntryLogger) WithError(err error) interfaces.Logger {
	return &logrusEntryLogger{entry: l.entry.WithError(err)}
}
