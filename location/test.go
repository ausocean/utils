// +build test

/*
DESCRIPTION
  test.go provides a test constructor function for the NMEAGPS type which creates
  a dummy port for which random pseudo GPS data is generated.

AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2021 the Australian Ocean Lab (AusOcean)

  This file is part of VidGrind. VidGrind is free software: you can
  redistribute it and/or modify it under the terms of the GNU
  General Public License as published by the Free Software
  Foundation, either version 3 of the License, or (at your option)
  any later version.

  VidGrind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with NetReceiver in gpl.txt.  If not, see
  <http://www.gnu.org/licenses/>.
*/

package location

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	randMin = 0.0
	randMax = 500000.0
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// dummyPort implements io.Reader, providing pseudo GPS data in NMEA sentences.
type dummyPort struct {
	data []byte
}

func (p *dummyPort) Read(b []byte) (int, error) {
	for len(p.data) < len(b) {
		sentence := fmt.Sprintf("$GPGGA,181908.00,%f,N,%f,W,4,13,1.00,495.144,M,29.200,M,0.10,0000*40\n", randFloat64(), randFloat64())
		p.data = append(p.data, []byte(sentence)...)
	}
	n := copy(b, p.data)
	p.data = p.data[n:]
	return n, nil
}

func newNMEAGPS(name string, port int, alt float64) (*NMEAGPS, error) {
	return &NMEAGPS{port: &dummyPort{data: make([]byte, 0)}, alt: alt}, nil
}

func randFloat64() float64 { return randMin + rand.Float64()*(randMax-randMin) }
