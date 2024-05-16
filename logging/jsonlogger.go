/*
DESCRIPTION
  jsonlogger.go provides an implementation of the Logger interface providing
  JSON styled logs with message and key-value pairs for additional information.
  This logger also provides log suppression for repeated messages.

AUTHOR
  Jack Richardson <richardson.jack@outlook.com>
  Saxon Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2017-2021 the Australian Ocean Lab (AusOcean).

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  in gpl.txt. If not, see [GNU licenses](http://www.gnu.org/licenses).
*/

package logging

import (
	"io"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Caller skip i.e. at what level the caller is logged as.
const callerSkip = 2

// Sampler defaults.
const (
	defaultSamplerTick = time.Minute
	defaultLogFirst    = 3
	defaultThenEvery   = int(math.MaxInt32)
)

// Default zapcore encoder keys, used in default encoder configuration.
const (
	defaultMessageKey    = "message"
	defaultLevelKey      = "level"
	defaultTimeKey       = "time"
	defaultCallerKey     = "caller"
	defaultStackTraceKey = "stackTrace"
)

type JSONLogger struct {
	*zap.SugaredLogger
	level         zap.AtomicLevel
	samplerTick   time.Duration // The sampling interval.
	logFirst      int           // How many times we log of a particular message before supressing.
	thenEvery     int           // Then after logging logFirst we log every thenEvery if more messages come.
	verbosity     int8
	writer        io.Writer // The destination the logger is writing to.
	suppress      bool      // Indicates state of suppression on repetitive logging.
	callerFilters []string  // Filters to apply to caller.
	config        zapcore.EncoderConfig
	mu            sync.Mutex
}

// New generates and returns a new JSONLogger.
// Optionally, a zapcore encoder config can be passed, otherwise a default is used.
func New(verbosity int8, writer io.Writer, suppress bool, config ...zapcore.EncoderConfig) *JSONLogger {
	// set up a default zapcore encoder configuration.
	cfg := zapcore.EncoderConfig{
		MessageKey:    defaultMessageKey,
		LevelKey:      defaultLevelKey,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		TimeKey:       defaultTimeKey,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		CallerKey:     defaultCallerKey,
		EncodeCaller:  zapcore.ShortCallerEncoder,
		StacktraceKey: defaultStackTraceKey,
	}

	// If an optional encoder config is provided we overwrite the default above.
	if len(config) != 0 {
		cfg = config[0]
	}

	// Populate logger fields that will then be used for initialisation.
	l := &JSONLogger{
		level:       zap.NewAtomicLevel(),
		samplerTick: defaultSamplerTick,
		logFirst:    defaultLogFirst,
		thenEvery:   defaultThenEvery,
		verbosity:   verbosity,
		writer:      writer,
		suppress:    suppress,
		config:      cfg,
	}
	l.init()
	return l
}

// Logging methods for each level.
func (l *JSONLogger) Debug(msg string, args ...interface{})           { l.log(Debug, msg, args...) }
func (l *JSONLogger) Info(msg string, args ...interface{})            { l.log(Info, msg, args...) }
func (l *JSONLogger) Warning(msg string, args ...interface{})         { l.log(Warning, msg, args...) }
func (l *JSONLogger) Error(msg string, args ...interface{})           { l.log(Error, msg, args...) }
func (l *JSONLogger) Fatal(msg string, args ...interface{})           { l.log(Fatal, msg, args...) }
func (l *JSONLogger) Log(level int8, msg string, args ...interface{}) { l.log(level, msg, args...) }

// Log takes a log level, message and arbitrary number of key:value pairs and logs them using
// appropriate Zap call. Logs may be filtered based on the caller file name if SetCallerFilters
// is used.
func (l *JSONLogger) log(level int8, message string, args ...interface{}) {
	if !l.shouldLog() {
		return
	}

	// Lock so that we synchronise with any re-initialisation.
	l.mu.Lock()
	switch level {
	case Fatal:
		l.Fatalw(message, args...)
	case Error:
		l.Errorw(message, args...)
	case Warning:
		l.Warnw(message, args...)
	case Info:
		l.Infow(message, args...)
	case Debug:
		l.Debugw(message, args...)
	}

	if level >= Warning {
		l.Sync()
	}
	l.mu.Unlock()
}

// shouldLog returns true if the caller should be logged, and false otherwise
// based on whether the log caller file is in the callerFilters.
func (l *JSONLogger) shouldLog() bool {
	// Get the file name where the log was called from.
	const skip = 3
	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		return true
	}
	// Make sure we have the base.
	file = filepath.Base(file)
	for _, f := range l.callerFilters {
		if strings.Contains(file, f) {
			return false
		}
	}
	return true
}

// SetSamplerTick sets the global samplerTick that will apply for all loggers.
func (l *JSONLogger) SetSamplerTick(d time.Duration) {
	l.samplerTick = d
	l.init()
}

// SetLogFirst sets the global logFirst count that will apply for all loggers.
func (l *JSONLogger) SetLogFirst(n int) {
	l.logFirst = n
	l.init()
}

// SetThenEvery sets the global thenEvery count that will apply for all loggers.
func (l *JSONLogger) SetThenEvery(n int) {
	l.thenEvery = n
	l.init()
}

// SetLevel sets the maximum log level that will be written to file
func (l *JSONLogger) SetLevel(level int8) {
	l.level.SetLevel(zapcore.Level(level))
}

// SetSuppress will turn on log sampling if s is true, and false otherwise.
func (l *JSONLogger) SetSuppress(s bool) {
	l.suppress = s
	l.init()
}

// SetCallerFilters will set the caller filters.
// Therefore, if a caller file is in the callerFilters, it will not be logged.
func (l *JSONLogger) SetCallerFilters(filters ...string) {
	l.callerFilters = filters
}

// init will initialise the logger with a zap logger containing a core, which
// may also possess a sampler.
func (l *JSONLogger) init() {
	// Lock so that we synchronise with any logging currently happening.
	l.mu.Lock()
	defer l.mu.Unlock()

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(l.config),
		zapcore.AddSync(l.writer),
		l.level,
	)

	// If we're suppressing repetitive logs, we add a sampling layer to the core.
	if l.suppress {
		core = zapcore.NewSampler(core, l.samplerTick, l.logFirst, l.thenEvery)
	}

	l.SugaredLogger = zap.New(core).WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(callerSkip),
		zap.AddStacktrace(zap.ErrorLevel),
	).Sugar()

	l.SetLevel(l.verbosity)
}
