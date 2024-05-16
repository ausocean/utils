/*
NAME
  ring.go - a structure that encapsulates a Buffer datastructure with conccurency
  functionality

DESCRIPTION
  See Readme.md

AUTHOR
  Dan Kortschak <dan@ausocean.org>

LICENSE
  ring.go is Copyright (C) 2017 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
 for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt.  If not, see http://www.gnu.org/licenses.
*/

// Package ring provides a ring buffer of io.ReadWriters.
package ring

import (
	"errors"
	"io"
	"time"
)

var (
	ErrTimeout = errors.New("ring: buffer cycle timeout")
	ErrDropped = errors.New("ring: dropped old write")
	ErrStall   = errors.New("ring: unable to dump old write")
	ErrTooLong = errors.New("ring: write too long for buffer element")
)

// Buffer implements a ring buffer.
//
// The buffer has a writable head and a readable tail with a queue from the head
// to the tail. Concurrent read a write operations are safe.
type Buffer struct {
	head, tail  *Chunk
	full, empty chan *Chunk
	timeout     time.Duration
}

// NewBuffer returns a Buffer with len elements of the given size. The timeout
// parameter specifies how long a write operation will wait before failing with
// a temporary timeout error.
func NewBuffer(len, size int, timeout time.Duration) *Buffer {
	if len <= 0 || size <= 0 {
		return nil
	}
	b := Buffer{
		full:    make(chan *Chunk, len),
		empty:   make(chan *Chunk, len),
		timeout: timeout,
	}
	for i := 0; i < len; i++ {
		b.empty <- newChunk(make([]byte, 0, size))
	}
	return &b
}

// Len returns the number of full buffer elements.
func (b *Buffer) Len() int {
	return len(b.full)
}

// Write writes the bytes in b to the next current or next available element of the ring buffer
// it returns the number of bytes written and any error.
// If no element can be gained within the timeout or stolen from the queue, ErrStall is returned
// and if the len(p) is greater than the buffer's element size, ErrTooLong is returned. If a
// write was successful but a previous write was dropped, ErrDropped is returned.
//
// Write is safe to use concurrently with Read, but may not be used concurrently with another
// write operation.
func (b *Buffer) Write(p []byte) (int, error) {
	var dropped bool
	if b.head == nil {
		timer := time.NewTimer(b.timeout)
		select {
		case <-timer.C:
			select {
			case b.head = <-b.full:
				b.head.reset()
				dropped = true
			default:
				// This should never happen.
				return 0, ErrStall
			}
		case b.head = <-b.empty:
			timer.Stop()
		}
	}
	if len(p) > b.head.cap() {
		return 0, ErrTooLong
	}
	if len(p) > b.head.cap()-b.head.Len() {
		b.full <- b.head
		b.head = nil
		return b.Write(p)
	}
	n, err := b.head.write(p)
	if b.head.cap()-b.head.Len() == 0 {
		b.full <- b.head
		b.head = nil
	}
	if dropped && err == nil {
		err = ErrDropped
	}
	return n, err
}

// Flush puts the currently writing element of the buffer into the queue for reading. Flush
// is idempotent.
//
// Flush is safe to use concurrently with Read, but may not be used concurrently with another
// another write operation.
func (b *Buffer) Flush() {
	if b.head == nil {
		return
	}
	b.full <- b.head
	b.head = nil
}

// Close closes the buffer. The buffer may not be written to after a call to close, but can
// be drained by calls to Read.
//
// Flush is safe to use concurrently with Read, but may not be used concurrently with another
// another write operation.
func (b *Buffer) Close() error {
	b.Flush()
	close(b.full)
	return nil
}

// Next gets the next element from the queue ready for reading, returning ErrTimeout if no
// element is available within the timeout. If the Buffer has been closed Next returns io.EOF.
//
// Is it the responsibility of the caller to close the returned Chunk unless the chunk is
// implicitly consumed by reading the Buffer until the io.EOF. A completely consuming read
// will close the chunk implicitly.
//
// Next is safe to use concurrently with write operations, but may not be used concurrently with
// another Read call or Next call. A goroutine calling Next must not call Flush or Close.
func (b *Buffer) Next(timeout time.Duration) (*Chunk, error) {
	if b.tail == nil {
		timer := time.NewTimer(timeout)
		var ok bool
		select {
		case <-timer.C:
			return nil, ErrTimeout
		case b.tail, ok = <-b.full:
			timer.Stop()
			if !ok {
				return nil, io.EOF
			}
		}
	}
	b.tail.owner = b
	return b.tail, nil
}

// Read reads bytes from the current tail of the ring buffer into p and returns the number of
// bytes read and any error.
//
// Read is safe to use concurrently with write operations, but may not be used concurrently with
// another Read call or Next call. A goroutine calling Read must not call Flush or Close.
func (b *Buffer) Read(p []byte) (int, error) {
	if b.tail == nil {
		return 0, io.EOF
	}
	n, err := b.tail.read(p)
	if b.tail.Len() == 0 {
		b.tail.reset()
		b.tail.owner = nil
		b.empty <- b.tail
		b.tail = nil
	}
	return n, err
}

// Chunk is a simplified version of byte buffer without the capacity to grow beyond the
// buffer's original cap, and a modified WriteTo method that allows multiple calls without
// consuming the buffered data.
type Chunk struct {
	buf   []byte
	off   int
	owner *Buffer
}

func newChunk(buf []byte) *Chunk {
	return &Chunk{buf: buf[:0]}
}

// Len returns the number of bytes held in the chunk.
func (b *Chunk) Len() int {
	return len(b.buf) - b.off
}

func (b *Chunk) cap() int {
	return cap(b.buf)
}

func (b *Chunk) reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

func (b *Chunk) write(p []byte) (n int, err error) {
	if len(p) > cap(b.buf)-len(b.buf) {
		err = ErrTooLong
	}
	l := len(b.buf)
	m := l + len(p)
	if m > cap(b.buf) {
		m = cap(b.buf)
	}
	b.buf = b.buf[:m]
	n = copy(b.buf[l:], p)
	return n, err
}

func (b *Chunk) read(p []byte) (n int, err error) {
	if b.Len() <= 0 {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}

// Bytes returns a slice of length b.Len() holding the unread portion of the Chunk.
// The slice is valid for use only until the next call to b.Close or Buffer.Read on
// the Buffer that returned b.
// The slice aliases the buffer content at least until the next buffer modification,
// so immediate changes to the slice will affect the result of future reads.
func (b *Chunk) Bytes() []byte {
	return b.buf[b.off:]
}

// WriteTo writes data to w until there's no more data to write or when an error occurs.
// The return value n is the number of bytes written. Any error encountered during the
// write is also returned. Repeated called to WriteTo will write the same data until
// the Chunk's Close method is called.
//
// WriteTo will panic if the Chunk has not been obtained through a call to Buffer.Next or
// has been closed. WriteTo must be used in the same goroutine as the call to Next.
func (b *Chunk) WriteTo(w io.Writer) (n int64, err error) {
	if b.owner == nil || b.owner.tail != b {
		panic("ring: invalid use of ring buffer chunk")
	}
	_n, err := w.Write(b.buf)
	if _n > len(b.buf) {
		panic("ring: invalid byte count")
	}
	if _n != len(b.buf) && err == nil {
		err = io.ErrShortWrite
	}
	return int64(_n), err
}

// Close closes the Chunk, reseting its data and releasing it back to the Buffer. A Chunk
// may not be used after it has been closed. Close must be used in the same goroutine as
// the call to Next. Close is idempotent for each Chunk returned by Next.
func (b *Chunk) Close() error {
	if b.owner == nil || b.owner.tail != b {
		return nil
	}
	b.reset()
	b.owner.tail = nil
	empty := b.owner.empty
	b.owner = nil
	empty <- b
	return nil
}
