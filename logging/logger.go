/*
DESCRIPTION
  logger.go provides a logging interface with different levels.

AUTHOR
  Saxon Nelson-Milton <saxon@ausocean.org>
  Jack Richardson <richardson.jack@outlook.com>

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
	"go.uber.org/zap"
)

// Used to define level of logging when making call to Logger.Log.
const (
	Fatal   = int8(zap.FatalLevel)
	Error   = int8(zap.ErrorLevel)
	Warning = int8(zap.WarnLevel)
	Info    = int8(zap.InfoLevel)
	Debug   = int8(zap.DebugLevel)
)

// Logger is a logger interface supporting multi level logging, and
// output level configuration.
type Logger interface {
	SetLevel(int8)
	Log(level int8, message string, params ...interface{})
	Debug(msg string, params ...interface{})
	Info(msg string, params ...interface{})
	Warning(msg string, params ...interface{})
	Error(msg string, params ...interface{})
	Fatal(msg string, params ...interface{})
}
