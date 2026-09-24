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
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MultiLogger is an implementation of the Logger interface that multiplexes
// log entries to multiple underlying Loggers.
type MultiLogger struct {
	*zap.SugaredLogger
	level         zap.AtomicLevel
	callerFilters []string
	mu            sync.Mutex
	loggers       []Logger
}

// NewMultiLogger combines multiple Logger instances into one using a zapcore.Tee.
func NewMultiLogger(loggers ...Logger) Logger {
	var cores []zapcore.Core
	var lowestLevel zapcore.Level = zapcore.FatalLevel

	for _, l := range loggers {
		if jl, ok := l.(*JSONLogger); ok {
			cores = append(cores, jl.Desugar().Core())
			if jl.level.Level() < lowestLevel {
				lowestLevel = jl.level.Level()
			}
		} else if cl, ok := l.(*TerminalLogger); ok {
			cores = append(cores, cl.Desugar().Core())
			if cl.level.Level() < lowestLevel {
				lowestLevel = cl.level.Level()
			}
		}
	}

	teeCore := zapcore.NewTee(cores...)
	level := zap.NewAtomicLevelAt(lowestLevel)

	sl := zap.New(teeCore).WithOptions(
		zap.AddCaller(),
		zap.AddCallerSkip(callerSkip), // from jsonlogger.go
		zap.AddStacktrace(zap.ErrorLevel),
	).Sugar()

	return &MultiLogger{
		SugaredLogger: sl,
		level:         level,
		loggers:       loggers,
	}
}

func (m *MultiLogger) Debug(msg string, args ...interface{})           { m.log(Debug, msg, args...) }
func (m *MultiLogger) Info(msg string, args ...interface{})            { m.log(Info, msg, args...) }
func (m *MultiLogger) Warning(msg string, args ...interface{})         { m.log(Warning, msg, args...) }
func (m *MultiLogger) Error(msg string, args ...interface{})           { m.log(Error, msg, args...) }
func (m *MultiLogger) Fatal(msg string, args ...interface{})           { m.log(Fatal, msg, args...) }
func (m *MultiLogger) Log(level int8, msg string, args ...interface{}) { m.log(level, msg, args...) }

func (m *MultiLogger) log(level int8, message string, args ...interface{}) {
	if !m.shouldLog() {
		return
	}

	m.mu.Lock()
	switch level {
	case Fatal:
		m.Fatalw(message, args...)
	case Error:
		m.Errorw(message, args...)
	case Warning:
		m.Warnw(message, args...)
	case Info:
		m.Infow(message, args...)
	case Debug:
		m.Debugw(message, args...)
	}

	if level >= Warning {
		m.Sync()
	}
	m.mu.Unlock()
}

func (m *MultiLogger) shouldLog() bool {
	const skip = 3
	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		return true
	}
	file = filepath.Base(file)
	for _, f := range m.callerFilters {
		if strings.Contains(file, f) {
			return false
		}
	}
	return true
}

func (m *MultiLogger) SetLevel(level int8) {
	m.level.SetLevel(zapcore.Level(level))
	for _, l := range m.loggers {
		l.SetLevel(level)
	}
}
