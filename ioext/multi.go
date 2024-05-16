/*
NAME
  multi.go

DESCRIPTION
  multi.go provides a multiWriteCloser that can perform Write and Close on
  multiple io.WriteClosers.

AUTHORS
  Saxon A. Nelson-Milton <saxon@ausocean.org>
  Dan Kortschak <dan@ausocean.org>

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
	"fmt"
	"io"
)

// Errors is a collection of errors.
type multiError []error

func (e multiError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if len(e) == 0 {
		return "<empty>"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%q", []error(e))
}

// MultiWriteCloser creates an io.WriteCloser that duplicates its
// writes to all the provided io.WriteClosers, similar to the Unix
// tee(1) command. Similarly, a close of the io.WriteCloser is
// passed on to all of the provided io.WriteClosers.
//
// Each write is written to each listed writer, one at a time.
// If a listed writer returns an error, that overall write operation
// continues but the error is retained and returned in a []error.
// Failures during close calls are treated the same way.
type multiWriteCloser struct {
	writers []io.WriteCloser
}

// Write implements io.Writer.
func (t *multiWriteCloser) Write(p []byte) (int, error) {
	var err multiError
	for _, w := range t.writers {
		_, e := w.Write(p)
		if e != nil {
			err = append(err, e)
		}
	}
	if len(err) == 0 {
		return len(p), nil
	}
	return len(p), err
}

// Close calls Close on all it's io.CloseWriters in closeWriters.
func (t *multiWriteCloser) Close() error {
	var err multiError
	for _, wc := range t.writers {
		e := wc.Close()
		if e != nil {
			err = append(err, e)
		}
	}
	if len(err) == 0 {
		return nil
	}
	return err
}

// MultiWriteCloser returns a pointer as io.Writer to a new multiWriteCloser.
func MultiWriteCloser(writers ...io.WriteCloser) io.WriteCloser {
	allWriters := make([]io.WriteCloser, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriteCloser); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	return &multiWriteCloser{allWriters}
}
