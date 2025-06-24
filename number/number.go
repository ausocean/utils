/*
AUTHOR
  Trek Hopton <trek@ausocean.org>
LICENSE
  Copyright (C) 2025 the Australian Ocean Lab (AusOcean)

  This is free software: you can redistribute it and/or modify it
  under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  It is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  in gpl.txt. If not, see http://www.gnu.org/licenses/.
*/

// package number is a simple package for generating types of numbers and IDs.
package number

import (
	"math/rand"
	"time"
)

// GenerateInt64ID generates a 10 digit int64 value which can be used as
// an ID in many datastore types.
//
// NOTE: the generated ID can be cast to an int32 if required.
func GenerateInt64ID() int64 {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	// This function generates a random number between 0, and
	// the largest number which can be expressed as a signed int32.
	// Subtracting 1000000000 from the range allows 1000000000 to be
	// added back to the number after generation to ensure that the
	// value is at least 10 digits long.
	return r.Int63n((1<<31)-1000000000) + 1000000000
}
