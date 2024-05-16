//go:build !nopool && !managed
// +build !nopool,!managed

/*
NAME
  pool.go - a structure that encapsulates a Buffer data structure with concurrency
  functionality

DESCRIPTION
  See Readme.md

AUTHOR
  Dan Kortschak <dan@ausocean.org>

LICENSE
  pool.go is Copyright (C) 2020 the Australian Ocean Lab (AusOcean)

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
	c := pool[bits(uint64(l))].Get().(*Chunk)
	n := allocated + cap(c.buf)
	if n < 0 {
		panic("pool: negative allocation")
	}
	allocated = n
	mu.Unlock()
	return c
}

func putChunk(b *Chunk) {
	b.buf = b.buf[:0]
	b.off = 0
	mu.Lock()
	n := allocated - cap(b.buf)
	if n < 0 {
		panic("pool: negative allocation")
	}
	allocated = n
	mu.Unlock()
	pool[bits(uint64(cap(b.buf)))].Put(b)
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
			b.buf = b.buf[:0]
			b.off = 0
			n := allocated - cap(b.buf)
			if n < 0 {
				panic("pool: allocation underflow")
			}
			allocated = n
			pool[bits(uint64(cap(b.buf)))].Put(b)
			dropped = true
		default:
			// This should never happen.
			return false, ErrStall
		}
	}

	return dropped, nil
}

var (
	// pool contains size stratified buffer chunk pools.
	// Each pool element i returns sized Chunks with a buf
	// slice capped at 1<<i.
	pool [63]sync.Pool
)

func init() {
	for i := range pool {
		l := 1 << uint(i)
		pool[i].New = func() interface{} {
			return &Chunk{buf: make([]byte, 0, l)}
		}
	}
}

// bits returns the ceiling of base 2 log of v.
// Approach based on http://stackoverflow.com/a/11398748.
func bits(v uint64) byte {
	if v == 0 {
		return 0
	}
	v <<= 2
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	return tab64[((v-(v>>1))*0x07EDD5E59A4E28C2)>>58] - 1
}

var tab64 = [64]byte{
	0x3f, 0x00, 0x3a, 0x01, 0x3b, 0x2f, 0x35, 0x02,
	0x3c, 0x27, 0x30, 0x1b, 0x36, 0x21, 0x2a, 0x03,
	0x3d, 0x33, 0x25, 0x28, 0x31, 0x12, 0x1c, 0x14,
	0x37, 0x1e, 0x22, 0x0b, 0x2b, 0x0e, 0x16, 0x04,
	0x3e, 0x39, 0x2e, 0x34, 0x26, 0x1a, 0x20, 0x29,
	0x32, 0x24, 0x11, 0x13, 0x1d, 0x0a, 0x0d, 0x15,
	0x38, 0x2d, 0x19, 0x1f, 0x23, 0x10, 0x09, 0x0c,
	0x2c, 0x18, 0x0f, 0x08, 0x17, 0x07, 0x06, 0x05,
}
