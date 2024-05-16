//go:build !nopool && managed
// +build !nopool,managed

/*
NAME
  managed.go - a structure that encapsulates a Buffer data structure with concurrency
  functionality

DESCRIPTION
  See Readme.md

AUTHOR
  Dan Kortschak <dan@ausocean.org>

LICENSE
  managed.go is Copyright (C) 2020 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt. If not, see http://www.gnu.org/licenses.
*/

package pool

import "sync"

// MaxAlloc sets the maximum total allocation allowed by pool buffers.
// The default is 1MiB.
func MaxAlloc(n int) {
	mu.Lock()
	maxAlloc = n
	mu.Unlock()
}

// Allocated returns the size of allocated buffers.
func Allocated() int {
	mu.Lock()
	n := allocated
	mu.Unlock()
	return n
}

// allocated is the amount of currently allocated buffer space.
// It does not include Chunk value allocations or slice headers.
var (
	mu        sync.Mutex
	maxAlloc  int = 1 << 20
	allocated int
)

func getChunk(l int) *Chunk {
	mu.Lock()
	n := allocated + l
	if n < 0 {
		panic("pool: negative allocation")
	}
	allocated = n
	mu.Unlock()
	return &Chunk{buf: make([]byte, 0, l)}
}

func putChunk(b *Chunk) {
	mu.Lock()
	n := allocated - cap(b.buf)
	if n < 0 {
		panic("pool: negative allocation")
	}
	allocated = n
	mu.Unlock()
	b.buf = nil
}

func stealFrom(chunks <-chan *Chunk, want int) (dropped bool, err error) {
	defer mu.Unlock()
	mu.Lock()

	if want > maxAlloc {
		return false, ErrTooLongForPool
	}

	for allocated+want > maxAlloc {
		select {
		case b, ok := <-chunks:
			if !ok {
				return false, ErrClosed
			}
			n := allocated - cap(b.buf)
			if n < 0 {
				panic("pool: allocation underflow")
			}
			allocated = n
			dropped = true
		default:
			// This should never happen.
			return false, ErrStall
		}
	}

	return dropped, nil
}
