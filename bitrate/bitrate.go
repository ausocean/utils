/*
NAME
  bitrate

DESCRIPTION
  Utility for calculating the bitrate from various senders.

AUTHOR
  Scott Barnard <scott@ausocean.org>

LICENSE
  bitrate is Copyright (C) 2020 the Australian Ocean Lab (AusOcean).

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt.  If not, see [GNU licenses](http://www.gnu.org/licenses).
*/

// Package bitrate is a utility for calculating the bitrate from various senders.
package bitrate

import (
	"sync"
	"time"
)

// Calculator is used for calculating the bitrate from one or many reporting sources.
type Calculator struct {
	sent int
	time time.Time
	mu   sync.Mutex
}

// NewCalculator returns a new Calculator struct and starts its internal timer.
func NewCalculator() *Calculator {
	return &Calculator{time: time.Now()}
}

// Report is used for reporting the amount of data sent by a sender.
func (b *Calculator) Report(l int) {
	b.mu.Lock()
	b.sent += l
	b.mu.Unlock()
}

// Bitrate calculates the bitrate of all senders combined since the last time it was called.
func (b *Calculator) Bitrate() int {
	b.mu.Lock()
	dur := time.Now().Sub(b.time).Milliseconds()
	br := int(int64(b.sent) * 1000.0 / dur)

	b.time = time.Now()
	b.sent = 0
	b.mu.Unlock()

	return br * 8
}
