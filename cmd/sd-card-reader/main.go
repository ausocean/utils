/*
AUTHORS
	Alan Noble <alan@ausocean.org>
LICENSE
	Copyright (C) 2025 the Australian Ocean Lab (AusOcean).
	This is free software: you can redistribute it and/or modify it
	under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	This is distributed in the hope that it will be useful, but WITHOUT
	ANY WARRANTY; without even the implied warranty of MERCHANTABILITY
	or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public
	License for more details.
	You should have received a copy of the GNU General Public License in
	gpl.txt. If not, see http://www.gnu.org/licenses/.
*/

// Utility to read a NetSender SD card data file.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	version       = 1
	versionMarker = 0x7ffffffe
	timeMarker    = 0x7fffffff
)

// SmallScalar defines the SD card data file record format.
// that NetSender uses in offline mode.
type SmallScalar struct {
	Value     int32
	Timestamp uint32
}

type FileInfo struct {
	Version, Count                  int
	Start, Finish, Duration, MaxGap uint32
	Sequential                      bool
}

func main() {
	var path string
	var info bool

	flag.StringVar(&path, "f", "", "Path to the binary data file")
	flag.BoolVar(&info, "i", false, "Print file info only")
	flag.Parse()

	if path == "" {
		fmt.Printf("File not specified (-f option).")
		return
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a record to read raw bytes into.
	sz := binary.Size(SmallScalar{})
	if sz != 8 {
		panic("SmallScalar should be 8 bytes in size")
	}

	buf := make([]byte, sz)
	var fileInfo FileInfo = FileInfo{Sequential: true}
	var refTs, prevTs uint32
	for {
		n, err := file.Read(buf)

		if err != nil {
			if err == io.EOF {
				break // We're finished.
			}
			fmt.Printf("Error reading from file: %v\n", err)
			return
		}

		// Check if we read a full buffer.
		if n != len(buf) {
			fmt.Printf("Error: Partial read (read %d bytes, expected %d)\n", n, len(buf))
			return
		}

		// Create a SmallScalar struct to unpack data into
		var record SmallScalar

		// NB: Arduino/ESP32 stores numbers in little-endian format.
		err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &record)
		if err != nil {
			fmt.Printf("Error unpacking record #%d: %v\n", fileInfo.Count+1, err)
			return
		}

		switch record.Value {
		case versionMarker:
			v := (int)(record.Timestamp)
			if v != version {
				fmt.Printf("Error: data file has wrong version %d", v)
				return
			}
			fileInfo.Version = v

		case timeMarker:
			refTs = record.Timestamp
			prevTs = refTs

		default:
			if !info {
				fmt.Printf("%d,%d\n", record.Timestamp, record.Value)
			}
			if fileInfo.Start == 0 {
				fileInfo.Start = record.Timestamp
			}
			fileInfo.Finish = record.Timestamp
			if record.Timestamp < prevTs {
				fileInfo.Sequential = false
			}
			if prevTs != 0 && record.Timestamp-prevTs > fileInfo.MaxGap {
				fileInfo.MaxGap = record.Timestamp - prevTs

			}
			prevTs = record.Timestamp
		}

		fileInfo.Count++
	}

	if info {
		const secondsPerDay = 60 * 60 * 24
		fileInfo.Duration = fileInfo.Finish - fileInfo.Start
		fmt.Printf("Version:    %10d\n", fileInfo.Version)
		fmt.Printf("Count:      %10d\n", fileInfo.Count)
		fmt.Printf("Start:      %10d %s\n", fileInfo.Start,
			time.Unix(int64(fileInfo.Start), 0).UTC().Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("Finish:     %10d %s\n", fileInfo.Finish,
			time.Unix(int64(fileInfo.Finish), 0).UTC().Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("Duration:   %10ds %5.1fd\n", fileInfo.Duration, (float64)(fileInfo.Duration)/secondsPerDay)
		fmt.Printf("Max Gap:    %10ds\n", fileInfo.MaxGap)
		fmt.Printf("Sequential: %10t\n", fileInfo.Sequential)
	}
}
