// +build nopool
// +build !managed

/*
NAME
  nopool.go - a structure that encapsulates a Buffer data structure with concurrency
  functionality

DESCRIPTION
  See Readme.md

AUTHOR
  Dan Kortschak <dan@ausocean.org>

LICENSE
  nopool.go is Copyright (C) 2020 the Australian Ocean Lab (AusOcean)

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

// MaxAlloc is a no-op.
func MaxAlloc(n int) {}

// Allocated returns -1 indicating an unknown allocation of buffers.
func Allocated() int { return -1 }

func getChunk(l int) *Chunk {
	return &Chunk{buf: make([]byte, 0, l)}
}

func putChunk(b *Chunk) {}

func stealFrom(chunks <-chan *Chunk, want int) (dropped bool, err error) { return }
