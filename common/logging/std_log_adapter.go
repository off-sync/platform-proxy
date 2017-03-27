package logging

import (
	stdLog "log"

	"strings"

	"github.com/off-sync/platform-proxy/app/interfaces"
)

// NewStdLogAdapter creates a standard library logger based
// on the provided logger.
func NewStdLogAdapter(log interfaces.Logger) *stdLog.Logger {
	std := stdLog.New(&stdLogAdapterWriter{log: log}, "", 0)

	return std
}

type stdLogAdapterWriter struct {
	log interfaces.Logger
}

func (w *stdLogAdapterWriter) Write(p []byte) (n int, err error) {
	for _, s := range strings.Split(string(p), "\n") {
		if s == "" {
			continue
		}

		w.log.Info(s)
	}

	return len(p), nil
}
