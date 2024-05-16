/*
DESCRIPTION
  nmea.go provides functionality for getting GPS data from a NMEA serial device.

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

// Package location provides functionality for determining location using
// external serial GPS devices.
package location

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/kortschak/nmea"
)

// NMEAGPS represents an external NMEA device connected via serial port,
// providing methods for getting the latest GPS data.
type NMEAGPS struct {
	alt  float64
	port io.Reader
	errc chan error
}

// Loc represents a location, with latitude, longitude and altitude fields.
type Loc struct{ Lat, Lng, Alt float64 }

// NewNMEAGPS returns a new NMEAGPS.
func NewNMEAGPS(name string, baud int, alt float64) (*NMEAGPS, error) {
	return newNMEAGPS(name, baud, alt)
}

// Location returns the most recently updated location data.
func (g *NMEAGPS) Location() (Loc, error) {
	s := bufio.NewScanner(g.port)
	for s.Scan() {
		str := s.Text()
		if len(str) == 0 {
			return Loc{}, errors.New("no sentence")
		}

		var gga nmea.GGA
		err := nmea.ParseTo(&gga, str)
		switch err {
		case nil:
		case nmea.ErrNMEAType:
			continue
		default:
			return Loc{}, fmt.Errorf("could not parse gga sentence: %w", err)
		}

		if gga.NorthSouth == "S" {
			gga.Latitude *= -1
		}

		if gga.EastWest == "W" {
			gga.Longitude *= -1
		}
		return Loc{Lat: gga.Latitude, Lng: gga.Longitude}, nil
	}
	return Loc{}, io.ErrUnexpectedEOF
}
