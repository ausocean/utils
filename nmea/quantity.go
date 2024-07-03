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
// For example: Quantity{Name: "Apparent Wind Speed", Code: AWS, Type: "speed"}.
type Quantity struct {
	Name, Code, Type string
}

// Constant exports of NMEA codes.
const (
	ApparentWindAngle string = "AWA"
	ApparentWindSpeed string = "AWS"
	Audio             string = "AUD"
	Boolean           string = "BIN"
	Distance          string = "DIS"
	Depth             string = "DPT"
	GPSFix            string = "GGA"
	DCVoltage         string = "DCV"
	ACVoltage         string = "ACV"
	HeadingMagnetic   string = "HDM"
	HeadingTrue       string = "HDT"
	Humidity          string = "MMB"
	AirPressure       string = "MTA"
	AirTemperature    string = "MWH"
	WaterTemperature  string = "MTW"
	Precipitation     string = "PPT"
	SpeedOverGround   string = "SOG"
	SpeedThruWater    string = "STW"
	Turbidity         string = "TBD"
	TrueWindAngle     string = "TWA"
	TrueWindGust      string = "TWG"
	TrueWindSpeed     string = "TWS"
	WaveHeight        string = "MWS"
	Video             string = "VID"
	Other             string = "OTH"
)

// DefaultQuantities provides a list of common NMEA quantities we might measure.
func DefaultQuantities() []Quantity {
	return []Quantity{
		{Code: ApparentWindAngle, Name: "Apparent Wind Angle", Type: "angle"},
		{Code: ApparentWindSpeed, Name: "Apparent Wind Speed", Type: "speed"},
		{Code: Audio, Name: "Audio", Type: "audio"},
		{Code: Boolean, Name: "Boolean", Type: "bool"},
		{Code: Distance, Name: "Distance", Type: "length"},
		{Code: Depth, Name: "Depth", Type: "length"},
		{Code: GPSFix, Name: "GPS Fix", Type: "position"},
		{Code: DCVoltage, Name: "DC Voltage", Type: "voltage"},
		{Code: ACVoltage, Name: "AC Voltage", Type: "voltage"},
		{Code: HeadingMagnetic, Name: "Heading (Magnetic)", Type: "angle"},
		{Code: HeadingTrue, Name: "Heading (True)", Type: "angle"},
		{Code: Humidity, Name: "Humidity", Type: "percent"},
		{Code: AirPressure, Name: "Air Pressure", Type: "pressure"},
		{Code: AirTemperature, Name: "Air Temperature", Type: "temperature"},
		{Code: WaterTemperature, Name: "Water Temperature", Type: "temperature"},
		{Code: Precipitation, Name: "Precipitation", Type: "length"},
		{Code: SpeedOverGround, Name: "Speed Over Ground", Type: "speed"},
		{Code: SpeedThruWater, Name: "Speed Thru Water", Type: "speed"},
		{Code: Turbidity, Name: "Turbidity", Type: "turbidity"},
		{Code: TrueWindAngle, Name: "True Wind Angle", Type: "angle"},
		{Code: TrueWindGust, Name: "True Wind Gust", Type: "speed"},
		{Code: TrueWindSpeed, Name: "True Wind Speed", Type: "speed"},
		{Code: WaveHeight, Name: "Wave Height", Type: "distance"},
		{Code: Video, Name: "Video", Type: "video"},
		{Code: Other, Name: "Other", Type: "unknown"},
	}
}
