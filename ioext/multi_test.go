/*
NAME
  multi_test.go

DESCRIPTION
  multi_test.go provides testing functionality for the multiWriteCloser found
  in multi.go.

AUTHORS
  Saxon A. Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2019 the Australian Ocean Lab (AusOcean)

  This is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
 	for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt.  If not, see http://www.gnu.org/licenses.

	Copyright 2010 The Go Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/

package ioext

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// testWriteCloser will implement io.WriteCloser and will allow control over
// whether calls to Write or Close will fail for testing purposes.
type testWriteCloser struct {
	buf         []byte
	closed      bool
	failOnWrite bool
	failOnClose bool
}

// Write implements io.Writer.
func (wc *testWriteCloser) Write(d []byte) (int, error) {
	if wc.failOnWrite {
		return 0, errors.New("failed to write")
	}
	wc.buf = append(wc.buf, d...)
	return len(d), nil
}

// Close implements io.Closer.
func (wc *testWriteCloser) Close() error {
	if wc.failOnClose {
		return errors.New("failed to close")
	}
	wc.closed = true
	return nil
}

// TestWriteSuccess checks the behaviour of the multiWriteCloser.Write when all
// Writes to the io.WriteClosers are successful.
func TestWriteSuccess(t *testing.T) {
	writeClosers := []io.WriteCloser{
		&testWriteCloser{},
		&testWriteCloser{},
		&testWriteCloser{},
	}

	mwc := MultiWriteCloser(writeClosers...)

	testData := []byte{0x01, 0x02, 0x03, 0x04}

	n, err := mwc.Write(testData)
	if n != len(testData) {
		t.Errorf("number of bytes written is not expected. Got: %v\n Want: %v\n", n, len(testData))
	}

	if err != nil {
		t.Errorf("unexpected error: %v\n", err)
	}

	for i, wc := range mwc.(*multiWriteCloser).writers {
		got := wc.(*testWriteCloser).buf
		if !bytes.Equal(got, testData) {
			t.Errorf("unexpected data in writeCloser buffer: %v\n Got: %v\n Want: %v\n", i, got, testData)
		}
	}
}

// TestWriteError checks the behaviour of multiWriteCloser.Write when an error is
// encountered.
func TestWriteError(t *testing.T) {
	writeClosers := []io.WriteCloser{
		&testWriteCloser{},
		&testWriteCloser{failOnWrite: true},
		&testWriteCloser{},
	}

	mwc := MultiWriteCloser(writeClosers...)

	testData := []byte{0x01, 0x02, 0x03, 0x04}

	n, err := mwc.Write(testData)
	if n == 0 {
		t.Errorf("number of bytes written is not expected. Got: %v\n Want: %v\n", n, len(testData))
	}
	if err == nil {
		t.Error("did not get error from write as expected")
	}

	for i, wc := range writeClosers {
		if !wc.(*testWriteCloser).failOnWrite {
			got := wc.(*testWriteCloser).buf
			if !bytes.Equal(testData, got) {
				t.Errorf("did not get data in testWriteCloser: %v as expected. Got: %v\n Want: %v\n", i, got, testData)
			}
		} else {
			if len(wc.(*testWriteCloser).buf) != 0 {
				t.Errorf("testWriteCloser: %v that failed did not have empty buf as expected", i)
			}
		}
	}
}

// TestCloseSuccess checks the behaviour of multiWriteCloser.Close when all
// calls of Close on the io.WriteClosers are successful.
func TestCloseSuccess(t *testing.T) {
	writeClosers := []io.WriteCloser{
		&testWriteCloser{},
		&testWriteCloser{},
		&testWriteCloser{},
	}

	mwc := MultiWriteCloser(writeClosers...)

	err := mwc.(*multiWriteCloser).Close()
	if err != nil {
		t.Errorf("unexpected error from Close: %v", err)
	}

	for i, wc := range mwc.(*multiWriteCloser).writers {
		if !wc.(*testWriteCloser).closed {
			t.Errorf("testWriteCloser: %v not closed as expected", i)
		}
	}
}

// TestCloseError checks behaviour of multiWriteCloser.Close when
// one of the io.WriteCloser's return an error on their call to Close.
func TestCloseError(t *testing.T) {
	writeClosers := []io.WriteCloser{
		&testWriteCloser{},
		&testWriteCloser{failOnClose: true},
		&testWriteCloser{},
	}

	mwc := MultiWriteCloser(writeClosers...)

	err := mwc.(*multiWriteCloser).Close()
	if err == nil {
		t.Error("did not get expected error on close")
	}

	for i, wc := range writeClosers {
		if !wc.(*testWriteCloser).failOnClose {
			if !wc.(*testWriteCloser).closed {
				t.Errorf("testWriteCloser: %v was not closed as expected", i)
			}
		} else {
			if wc.(*testWriteCloser).closed {
				t.Errorf("testWriteCloser: %v should not have been closed", i)
			}
		}
	}
}
