package nmea

import (
	"errors"
	"fmt"
	"time"

	gonmea "github.com/adrianmo/go-nmea"
)

// GPSData stores data collected from NMEA sentences
type GPSData struct {
	Latitude         *float64   // Latitude
	Longitude        *float64   // Longitude
	LastFixTime      *time.Time // Time of last GPS fix
	Speed            *float64   // Speed in knots
	Course           *float64   // True course
	Variation        *float64   // Magnetic variation
	FixQuality       *string    // Quality of fix.
	NumSatellites    *int64     // Number of satellites in use.
	Altitude         *float64   // Altitude.
	TrueTrack        *float64   // True track made good (degrees)
	MagneticTrack    *float64   // Magnetic track made good
	GroundSpeedKnots *float64   // Ground speed in Knots
	GroundSpeedKPH   *float64   // Ground speed in km/hr
	Heading          *float64   // Heading in degrees
	True             *bool      // Heading is relative to true north
}

func ProcessSentence(dst GPSData, raw string) (GPSData, error) {
	s, err := gonmea.Parse(raw)
	if err != nil {
		return dst, fmt.Errorf("failed to process sentence: %w", err)
	}

	switch s := s.(type) {
	case gonmea.RMC:
		if s.Validity != "A" {
			// Not valid data.
			return dst, errors.New("invalid data in RMC sentence")
		}

		newFix := time.Date(
			s.Date.YY,
			time.Month(s.Date.MM),
			s.Date.DD,
			s.Time.Hour,
			s.Time.Minute,
			s.Time.Second,
			s.Time.Millisecond*1000000,
			time.UTC,
		)

		// Valid, update
		dst.Latitude = &s.Latitude
		dst.Longitude = &s.Longitude
		dst.Speed = &s.Speed
		dst.Course = &s.Course
		dst.Variation = &s.Variation
		dst.LastFixTime = &newFix
	case gonmea.GGA:
		currentTime := time.Now()
		newFix := time.Date(
			currentTime.Year(),
			currentTime.Month(),
			currentTime.Day(),
			s.Time.Hour,
			s.Time.Minute,
			s.Time.Second,
			s.Time.Millisecond*1000000,
			time.UTC,
		)
		dst.Latitude = &s.Latitude
		dst.Longitude = &s.Longitude
		dst.FixQuality = &s.FixQuality
		dst.NumSatellites = &s.NumSatellites
		dst.Altitude = &s.Altitude
		dst.LastFixTime = &newFix
	case gonmea.GLL:
		if s.Validity != "A" {
			// Not valid data.
			return dst, errors.New("invalid data in GLL sentence")
		}

		currentTime := time.Now()
		newFix := time.Date(
			currentTime.Year(),
			currentTime.Month(),
			currentTime.Day(),
			s.Time.Hour,
			s.Time.Minute,
			s.Time.Second,
			s.Time.Millisecond*1000000,
			time.UTC,
		)
		dst.Latitude = &s.Latitude
		dst.Longitude = &s.Longitude
		dst.LastFixTime = &newFix
	case gonmea.VTG:
		dst.TrueTrack = &s.TrueTrack
		dst.MagneticTrack = &s.MagneticTrack
		dst.GroundSpeedKnots = &s.GroundSpeedKnots
		dst.GroundSpeedKPH = &s.GroundSpeedKPH
	case gonmea.HDT:
		dst.Heading = &s.Heading
		dst.True = &s.True
	case gonmea.ZDA, gonmea.PGRME, gonmea.GSV, gonmea.GSA:
		return dst, fmt.Errorf("sentence type %s not implemented", s.Prefix())
	default:
		return dst, fmt.Errorf("failure to read parsed sentence: %s", s.String())
	}

	return dst, nil
}
