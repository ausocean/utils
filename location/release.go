// +build !test

/*
DESCRIPTION
  test.go provides a release constructor for the NMEAGPS type, which opens a
  serial port to which an external NMEA device should be connected.


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

	"github.com/tarm/serial"
)

func newNMEAGPS(name string, baud int, alt float64) (*NMEAGPS, error) {
	port, err := serial.OpenPort(&serial.Config{Name: name, Baud: baud})
	if err != nil {
		return nil, fmt.Errorf("could not open GPS serial port: %w", err)
	}
	return &NMEAGPS{port: port, alt: alt, errc: make(chan error)}, nil
}
