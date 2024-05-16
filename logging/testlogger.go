/*
DESCRIPTION
  testlogger.go provides an implementation of the Logger interface that wraps
  the testing.T struct for use where a Logger is expected in a test. I.e. the
  implementation uses the testing.T.Log functionality.

AUTHOR
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
  in gpl.txt.  If not, see [GNU licenses](http://www.gnu.org/licenses).
*/

package logging

import "testing"

// TestLogger provides an implementation of Logger. It uses the testing.T
// struct for logging, i.e. this logger is useful for code requiring an
// implementation of Logger that is being run in a test where we wish to
// capture logs to the testing output.
type TestLogger testing.T

func (tl *TestLogger) Debug(msg string, args ...interface{})   { tl.Log(Debug, msg, args...) }
func (tl *TestLogger) Info(msg string, args ...interface{})    { tl.Log(Info, msg, args...) }
func (tl *TestLogger) Warning(msg string, args ...interface{}) { tl.Log(Warning, msg, args...) }
func (tl *TestLogger) Error(msg string, args ...interface{})   { tl.Log(Error, msg, args...) }
func (tl *TestLogger) Fatal(msg string, args ...interface{})   { tl.Log(Fatal, msg, args...) }
func (tl *TestLogger) SetLevel(lvl int8)                       {}
func (dl *TestLogger) Log(lvl int8, msg string, args ...interface{}) {
	var l string
	switch lvl {
	case Warning:
		l = "warning"
	case Debug:
		l = "debug"
	case Info:
		l = "info"
	case Error:
		l = "error"
	case Fatal:
		l = "fatal"
	}
	msg = l + ": " + msg

	// Just use test.T.Log if no formatting required.
	if len(args) == 0 {
		((*testing.T)(dl)).Log(msg)
		return
	}

	// Add braces with args inside to message.
	msg += " ("
	for i := 0; i < len(args); i += 2 {
		msg += " %v:\"%v\""
	}
	msg += " )"

	if lvl == Fatal {
		dl.Fatalf(msg+"\n", args...)
	}

	dl.Logf(msg+"\n", args...)
}
