/*
LICENSE
  Copyright (C) 2024 the Australian Ocean Lab (AusOcean)

  This is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  This is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  in gpl.txt. If not, see http://www.gnu.org/licenses.
*/

// Package nmea defines some common NMEA functions.
package nmea

// Quantity describes a NMEA quantity code.
// For example: Quantity{Name: "Apparent Wind Speed", Code: "AWS", Type: "speed"}.
type Quantity struct {
	Name, Code, Type string
}

// DefaultQuantities provides a list of common NMEA quantities we might measure.
func DefaultQuantities() []Quantity {
	return []Quantity{
		{Code: "AWA", Name: "Apparent Wind Angle", Type: "angle"},
		{Code: "AWS", Name: "Apparent Wind Speed", Type: "speed"},
		{Code: "AUD", Name: "Audio", Type: "audio"},
		{Code: "BIN", Name: "Boolean", Type: "bool"},
		{Code: "DIS", Name: "Distance", Type: "length"},
		{Code: "DPT", Name: "Depth", Type: "length"},
		{Code: "GGA", Name: "GPS Fix", Type: "position"},
		{Code: "DCV", Name: "DC Voltage", Type: "voltage"},
		{Code: "ACV", Name: "AC Voltage", Type: "voltage"},
		{Code: "HDM", Name: "Heading (Magnetic)", Type: "angle"},
		{Code: "HDT", Name: "Heading (True)", Type: "angle"},
		{Code: "MMB", Name: "Humidity", Type: "percent"},
		{Code: "MTA", Name: "Air Pressure", Type: "pressure"},
		{Code: "MWH", Name: "Air Temperature", Type: "temperature"},
		{Code: "MTW", Name: "Water Temperature", Type: "temperature"},
		{Code: "PPT", Name: "Precipitation", Type: "length"},
		{Code: "SOG", Name: "Speed Over Ground", Type: "speed"},
		{Code: "STW", Name: "Speed Thru Water", Type: "speed"},
		{Code: "TBD", Name: "Turbidity", Type: "turbidity"},
		{Code: "TWA", Name: "True Wind Angle", Type: "angle"},
		{Code: "TWG", Name: "True Wind Gust", Type: "speed"},
		{Code: "TWS", Name: "True Wind Speed", Type: "speed"},
		{Code: "MWS", Name: "Wave Height", Type: "distance"},
		{Code: "VID", Name: "Video", Type: "video"},
		{Code: "OTH", Name: "Other", Type: "unknown"},
	}
}
