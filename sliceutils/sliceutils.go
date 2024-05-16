/*
DESCRIPTION
  sliceutils.go provides general functionality for working with slices.

AUTHOR
  saxon Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2019 the Australian Ocean Lab (AusOcean)

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

// Package sliceutils provides general functionality for working with slices.
package sliceutils

import (
	"strings"
)

// ContainsString returns true if the wanted value w is in the slice s.
func ContainsString(s []string, w string) bool {
	for _, v := range s {
		if v == w {
			return true
		}
	}
	return false
}

// ContainsStringPrefix returns true and the index if one of the strings in the slice s starts with the wanted value w.
func ContainsStringPrefix(s []string, w string) (bool, int) {
	for i, v := range s {
		if strings.HasPrefix(v, w) {
			return true, i
		}
	}
	return false, -1
}

// StringPart returns the nth part of a string slice, or the empty string if n is out of range.
func StringPart(s []string, n int) string {
	if n >= 0 && n < len(s) {
		return s[n]
	}
	return ""
}

// ContainsUint8 returns true if the wanted value w is in the slice s.
func ContainsUint8(s []uint8, w uint8) bool {
	for _, v := range s {
		if v == w {
			return true
		}
	}
	return false
}
