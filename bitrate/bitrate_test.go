/*
NAME
  bitrate_test.go

DESCRIPTION
  Tests for the bitrate package.

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

package bitrate

import (
	"testing"
	"time"
)

const tolerance = 22

var testData = []struct {
	data []int // Bits.
	want int   // Bits per second.
}{
	{
		data: []int{10},
		want: 320,
	},
	{
		data: []int{50, 50, 50},
		want: 4800,
	},
	{
		data: []int{120, 302, 152},
		want: 18368,
	},
	{
		data: []int{1340, 750, 830, 720, 1120, 960, 950, 940},
		want: 243520,
	},
}

func TestBitrate(t *testing.T) {
	bc := NewCalculator()
	for _, test := range testData {
		go func() {
			for _, len := range test.data {
				bc.Report(len)
			}
		}()

		time.Sleep(250 * time.Millisecond)

		got := bc.Bitrate()

		if test.want < (100-tolerance)*got/100 || test.want > (100+tolerance)*got/100 {
			t.Errorf("incorrect bitrate: got %d but want %d", got, test.want)
		}
	}
}
