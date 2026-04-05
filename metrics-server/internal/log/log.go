// Package log is used to implement server logger.
package log

import "go.uber.org/zap"

// NewLogger creates a new logger.
func NewLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	return logger.Sugar()
}
