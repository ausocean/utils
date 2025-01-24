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

// Code represents an NMEA code.
type Code string

// Type represents the type of value in an NMEA Quantity.
type Type string

// Quantity describes a NMEA quantity code.
// For example: Quantity{Name: "Apparent Wind Speed", Code: "AWS", Type: "speed"}.
type Quantity struct {
	Name string
	Code Code
	Type Type
}

// Constant exports of NMEA codes.
const (
	ApparentWindAngle Code = "AWA"
	ApparentWindSpeed Code = "AWS"
	Audio             Code = "AUD"
	Boolean           Code = "BIN"
	Distance          Code = "DIS"
	Depth             Code = "DPT"
	GPSFix            Code = "GGA"
	DCVoltage         Code = "DCV"
	DCCurrent         Code = "DCI"
	ACVoltage         Code = "ACV"
	HeadingMagnetic   Code = "HDM"
	HeadingTrue       Code = "HDT"
	Humidity          Code = "MMB"
	AirPressure       Code = "MTA"
	AirTemperature    Code = "MWH"
	WaterTemperature  Code = "MTW"
	Precipitation     Code = "PPT"
	SpeedOverGround   Code = "SOG"
	SpeedThruWater    Code = "STW"
	Turbidity         Code = "TBD"
	TrueWindAngle     Code = "TWA"
	TrueWindGust      Code = "TWG"
	TrueWindSpeed     Code = "TWS"
	WaveHeight        Code = "MWS"
	Video             Code = "VID"
	Other             Code = "OTH"
)

// Constant exports of NMEA types.
const (
	TypeAngle       Type = "angle"
	TypeSpeed       Type = "Speed"
	TypeAudio       Type = "audio"
	TypeBool        Type = "bool"
	TypeLength      Type = "length"
	TypePosition    Type = "position"
	TypeVoltage     Type = "voltage"
	TypeCurrent     Type = "current"
	TypePercent     Type = "percent"
	TypePressure    Type = "pressure"
	TypeTemperature Type = "temperature"
	TypeDistance    Type = "distance"
	TypeVideo       Type = "video"
	TypeUnknown     Type = "unknown"
)

// DefaultQuantities provides a list of common NMEA quantities we might measure.
func DefaultQuantities() []Quantity {
	return []Quantity{
		{Code: ApparentWindAngle, Name: "Apparent Wind Angle", Type: TypeAngle},
		{Code: ApparentWindSpeed, Name: "Apparent Wind Speed", Type: TypeSpeed},
		{Code: Audio, Name: "Audio", Type: TypeAudio},
		{Code: Boolean, Name: "Boolean", Type: TypeBool},
		{Code: Distance, Name: "Distance", Type: TypeLength},
		{Code: Depth, Name: "Depth", Type: TypeLength},
		{Code: GPSFix, Name: "GPS Fix", Type: TypePosition},
		{Code: DCVoltage, Name: "DC Voltage", Type: TypeVoltage},
		{Code: DCCurrent, Name: "DC Current", Type: TypeCurrent},
		{Code: ACVoltage, Name: "AC Voltage", Type: TypeVoltage},
		{Code: HeadingMagnetic, Name: "Heading (Magnetic)", Type: TypeAngle},
		{Code: HeadingTrue, Name: "Heading (True)", Type: TypeAngle},
		{Code: Humidity, Name: "Humidity", Type: TypePercent},
		{Code: AirPressure, Name: "Air Pressure", Type: TypePressure},
		{Code: AirTemperature, Name: "Air Temperature", Type: TypeTemperature},
		{Code: WaterTemperature, Name: "Water Temperature", Type: TypeTemperature},
		{Code: Precipitation, Name: "Precipitation", Type: TypeLength},
		{Code: SpeedOverGround, Name: "Speed Over Ground", Type: TypeSpeed},
		{Code: SpeedThruWater, Name: "Speed Thru Water", Type: TypeSpeed},
		{Code: Turbidity, Name: "Turbidity", Type: "turbidity"},
		{Code: TrueWindAngle, Name: "True Wind Angle", Type: TypeAngle},
		{Code: TrueWindGust, Name: "True Wind Gust", Type: TypeSpeed},
		{Code: TrueWindSpeed, Name: "True Wind Speed", Type: TypeSpeed},
		{Code: WaveHeight, Name: "Wave Height", Type: TypeDistance},
		{Code: Video, Name: "Video", Type: TypeVideo},
		{Code: Other, Name: "Other", Type: TypeUnknown},
	}
}
