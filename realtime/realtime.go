/*
DESCRIPTION
  realtime.go provides functionality for getting an accurate time if system
  time cannot be trusted.

AUTHOR
  Saxon Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2017-2018 the Australian Ocean Lab (AusOcean)

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

// Package realtime provides means to obtain a realtime when system time cannot be trusted.
package realtime

import (
	"sync"
	"time"
)

// RealTime provides means to obtain an accurate time if system time cannot be
// trusted. Set must be called and provided with an accurate reference time
// that will need to be obtained from an NTP server or locally networked machine.
// Calls to Get will use the reference time to calculate real time.
type RealTime struct {
	realRef time.Time  // Holds a reference real time given to SetTime.
	sysRef  time.Time  // Holds a system reference time set using time.Now() when realRefTime is obtained.
	isSet   bool       // Indicates if the time has been set.
	mu      sync.Mutex // Used when accessing/mutating above time vars.
}

// NewRealTime returns a new RealTime.
func NewRealTime() *RealTime {
	return &RealTime{mu: sync.Mutex{}}
}

// Set allows setting of current time. The user may wish to obtain an
// accurate time from an NTP server or local machine and pass to this function.
func (rt *RealTime) Set(t time.Time) {
	rt.mu.Lock()
	rt.realRef = t
	rt.sysRef = time.Now()
	rt.isSet = true
	rt.mu.Unlock()
}

// Get provides either a real time that has been calculated from a reference
// set by Set, or using the current system time if the real reference time has
// not been set.
func (rt *RealTime) Get() time.Time {
	rt.mu.Lock()
	t := rt.realRef.Add(time.Now().Sub(rt.sysRef))
	rt.mu.Unlock()
	return t
}

// IsSet returns true if Set has been used to set a real reference time.
func (rt *RealTime) IsSet() bool {
	rt.mu.Lock()
	b := rt.isSet
	rt.mu.Unlock()
	return b
}
