package interfaces

// Logger defines the methods required for a logger.
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}
