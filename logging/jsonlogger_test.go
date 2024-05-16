/*
DESCRIPTION
  jsonlogger_test.go provides testing for functionality found in jsonlogger.go.

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
  in gpl.txt.  If not, see [GNU licenses](http://www.gnu.org/licenses).
*/

package logging

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

var (
	logger *JSONLogger
	Types  = []zapcore.Level{zapcore.Level(Error), zapcore.Level(Warning), zapcore.Level(Info)}
)

func printMessages() {
	for i := 0; i < 100; i++ {
		logger.Log(int8(Types[i%3]), "this is a number", "Number", i)
	}
}

func TestFatal(t *testing.T) {
	t.Skip("Won't work with network logging")
	if os.Getenv("BE_CRASHER") == "1" {
		//logger = New(Info, "../")
		logger.Log(Info, "Testing Fatal Logging")
		logger.Log(Fatal, "dying")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatal")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestBasicLogging(t *testing.T) {
	t.Skip("Won't work with network logging")

	//outdated API
	//logger = New(Info, "../")
	logger.Log(Info, "Testing Basic Logging")
	printMessages()
}

func TestLevelLogging(t *testing.T) {
	t.Skip("Won't work with network logging")
	//logger = New(Info, "../")
	logger.Log(Info, "Testing Leveled Logging")
	printMessages()
	logger.SetLevel(Error)
	printMessages()
	logger.SetLevel(Fatal)
	printMessages()
}

func TestSampler(t *testing.T) {
	t.Skip("skipping in CI environment")

	tests := []struct {
		tick       time.Duration
		first      int
		thereafter int
		nLogs      int
	}{
		{
			tick:       time.Second,
			first:      10,
			thereafter: 1,
			nLogs:      30,
		},
		{
			tick:       time.Second,
			first:      10,
			thereafter: 10,
			nLogs:      30,
		},
		{
			tick:       time.Second,
			first:      5,
			thereafter: 1000,
			nLogs:      1000000,
		},
		{
			tick:       time.Second,
			first:      5,
			thereafter: int(math.MaxInt64),
			nLogs:      10000,
		},
	}

	for i, test := range tests {
		fmt.Printf(
			"test no.: %d tick: %d first: %d thereafter: %d nLogs: %d\n",
			i,
			test.tick,
			test.first,
			test.thereafter,
			test.nLogs,
		)
		l := New(Info, os.Stdout, false)
		l.SetSamplerTick(test.tick)
		l.SetLogFirst(test.first)
		l.SetThenEvery(test.thereafter)
		for i := 0; i < test.nLogs; i++ {
			l.Log(Info, "some message")
		}
	}
}

// TestCallerFilter tests that we can apply log caller filters
// using JSONLogger.SetCallerFilters and that the filters are applied.
func TestCallerFilter(t *testing.T) {
	var (
		logger Logger
		buf    bytes.Buffer
	)
	logger = New(Debug, &buf, false)

	// Get the current filename and add it to the caller filters.
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not get current filename")
	}
	logger.(*JSONLogger).SetCallerFilters(file)

	// Write a random log (from this file, which is in the caller filters).
	// This should be filtered out.
	logger.Info("testing filter")
	if strings.Contains(buf.String(), file) {
		t.Error("did not expect log to contain filename")
	}
}

// TestCallerSkip tests that the caller is correctly logged for
// each logging function.
func TestCallerSkip(t *testing.T) {
	// Get the current file name.
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not get current file name")
	}

	// Occasionally file is returned as an absolute path, so we need to get the
	// base name if we don't have it already.
	file = filepath.Base(file)

	var (
		logger Logger
		buf    bytes.Buffer
	)
	logger = New(Debug, &buf, false)

	// Test that each log function has the expected caller.
	for _, f := range []func(string, ...interface{}){
		logger.Debug,
		logger.Info,
		logger.Warning,
		logger.Error,
		// Can't test logger.Fatal as it calls os.Exit.
	} {
		f("test")
		if !strings.Contains(buf.String(), file) {
			t.Errorf("expected log caller to be %s for log function %s\nbut instead got %s", file, functionName(f), buf.String())
		}
		buf.Reset()
	}

	// Log is special because it needs a log level :/
	logger.Log(Debug, "test")
	if !strings.Contains(buf.String(), file) {
		t.Errorf("expected log caller to be %s for log function %s\nbut instead got %s", file, functionName(logger.Log), buf.String())
	}
	buf.Reset()
}

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
