/*
AUTHOR
  Trek Hopton <trek@ausocean.org>
LICENSE
  Copyright (C) 2026 the Australian Ocean Lab (AusOcean)

  This is free software: you can redistribute it and/or modify it
  under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  It is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  in gpl.txt. If not, see http://www.gnu.org/licenses/.
*/

package logging

import (
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TerminalLogger provides an implementation of the Logger interface providing
// terminal styled logs (human readable and colour coded).
type TerminalLogger struct {
	*zap.SugaredLogger
	level         zap.AtomicLevel
	samplerTick   time.Duration
	logFirst      int
	thenEvery     int
	verbosity     int8
	writer        io.Writer
	suppress      bool
	callerFilters []string
	config        zapcore.EncoderConfig
	mu            sync.Mutex
}

// NewTerminal generates and returns a new TerminalLogger.
func NewTerminal(verbosity int8, writer io.Writer, suppress bool, config ...zapcore.EncoderConfig) *TerminalLogger {
	cfg := zapcore.EncoderConfig{
		MessageKey:    defaultMessageKey,
		LevelKey:      defaultLevelKey,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		TimeKey:       defaultTimeKey,
		EncodeTime:    zapcore.TimeEncoderOfLayout("15:04:05.000"),
		CallerKey:     defaultCallerKey,
		EncodeCaller:  zapcore.ShortCallerEncoder,
		StacktraceKey: defaultStackTraceKey,
	}

	if len(config) != 0 {
		cfg = config[0]
	}

	l := &TerminalLogger{
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

func (l *TerminalLogger) Debug(msg string, args ...interface{})           { l.log(Debug, msg, args...) }
func (l *TerminalLogger) Info(msg string, args ...interface{})            { l.log(Info, msg, args...) }
func (l *TerminalLogger) Warning(msg string, args ...interface{})         { l.log(Warning, msg, args...) }
func (l *TerminalLogger) Error(msg string, args ...interface{})           { l.log(Error, msg, args...) }
func (l *TerminalLogger) Fatal(msg string, args ...interface{})           { l.log(Fatal, msg, args...) }
func (l *TerminalLogger) Log(level int8, msg string, args ...interface{}) { l.log(level, msg, args...) }

func (l *TerminalLogger) log(level int8, message string, args ...interface{}) {
	if !l.shouldLog() {
		return
	}

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

func (l *TerminalLogger) shouldLog() bool {
	const skip = 3
	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		return true
	}
	file = filepath.Base(file)
	for _, f := range l.callerFilters {
		if strings.Contains(file, f) {
			return false
		}
	}
	return true
}

func (l *TerminalLogger) SetLevel(level int8) {
	l.level.SetLevel(zapcore.Level(level))
}

func (l *TerminalLogger) init() {
	l.mu.Lock()
	defer l.mu.Unlock()

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(l.config),
		zapcore.AddSync(l.writer),
		l.level,
	)

	if l.suppress {
		core = zapcore.NewSamplerWithOptions(core, l.samplerTick, l.logFirst, l.thenEvery)
	}

	l.SugaredLogger = zap.New(core).WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(callerSkip),
		zap.AddStacktrace(zap.ErrorLevel),
	).Sugar()

	l.SetLevel(l.verbosity)
}
