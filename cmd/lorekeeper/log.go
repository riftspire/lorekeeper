package main

import (
	"os"

	"github.com/charmbracelet/log"
)

const (
	// Default log level - Warn.
	defaultLogLevel = log.WarnLevel
)

type _logLevels []log.Level

// charmbracelet/log doesn't have an AllLevels list.
var _AllLogLevels = _logLevels{
	log.DebugLevel,
	log.InfoLevel,
	log.WarnLevel,
	log.ErrorLevel,
	log.FatalLevel,
}

// initLogging handles the initialisation for the charmbracelet/log package.
func initLogging() {
	// Set the log level to the default log level.
	log.SetLevel(defaultLogLevel)

	// Set the log output to Stderr.
	log.SetOutput(os.Stderr)
}

// logLevelsAbove returns all log.Level that are higher than the provided
// log.Level.
func logLevelsAbove(level log.Level) []log.Level {
	var (
		levelIdx = logLevelIndex(level)
		levels   []log.Level
	)

	for idx, l := range _AllLogLevels {
		if idx > levelIdx {
			levels = append(levels, l)
		}
	}

	return levels
}

// logLevelIndex returns the index of the provided log.Level in the
// _AllLogLevels slice.
func logLevelIndex(level log.Level) int {
	for idx, l := range _AllLogLevels {
		if l == level {
			return idx
		}
	}
	return -1
}
