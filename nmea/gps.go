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

package nmea

import (
	"fmt"
	"time"

	gonmea "github.com/adrianmo/go-nmea"
)

// GPSData is a parsed GPS fix from a GGA sentence.
type GPSData struct {
	Lat   float64   // Latitude in decimal degrees.
	Lon   float64   // Longitude in decimal degrees.
	Alt   float64   // Altitude in meters.
	Time  time.Time // Full UTC timestamp using base date + GGA time-of-day.
	Valid bool      // True if the fix is valid (FixQuality != 0).
}

// Parse parses a GGA sentence using baseDate for the date and the GGA time for the time-of-day.
// baseDate should be in UTC (e.g., Text.Date.UTC()).
// Returns Valid=false (no error) if FixQuality is invalid.
func Parse(raw string, baseDate time.Time) (GPSData, error) {
	s, err := gonmea.Parse(raw)
	if err != nil {
		return GPSData{}, fmt.Errorf("go-nmea parse failed: %w", err)
	}

	gga, ok := s.(gonmea.GGA)
	if !ok {
		return GPSData{}, fmt.Errorf("unsupported NMEA type: %T", s)
	}

	if gga.FixQuality == gonmea.Invalid {
		return GPSData{Valid: false}, nil
	}

	// Build full timestamp from base date + GGA time-of-day in UTC.
	t := time.Date(
		baseDate.Year(), baseDate.Month(), baseDate.Day(),
		gga.Time.Hour, gga.Time.Minute, gga.Time.Second,
		gga.Time.Millisecond*1_000_000,
		time.UTC,
	)

	return GPSData{
		Lat:   gga.Latitude,
		Lon:   gga.Longitude,
		Alt:   gga.Altitude,
		Time:  t,
		Valid: true,
	}, nil
}
