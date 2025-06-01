// Utility to read a NetSender SD card data file.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	version       = 1
	versionMarker = 0x7ffffffe
	timeMarker    = 0x7fffffff
)

// SmallScalar defines the SD card data file record format
// that NetSender uses in offline mode.
type SmallScalar struct {
	Value     int32
	Timestamp uint32
}

func main() {
	var path string

	flag.StringVar(&path, "f", "", "Path to the binary data file")
	flag.Parse()

	if path == "" {
		fmt.Printf("File not specified (-f option).")
		return
	}

	fmt.Printf("* Reading data from '%s'\n", path)
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a record to read raw bytes into
	sz := binary.Size(SmallScalar{})
	if sz != 8 {
		panic("SmallScalar should be 8 bytes in size")
	}

	buf := make([]byte, sz)
	count := 0
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
			fmt.Printf("Error unpacking record #%d: %v\n", count+1, err)
			return
		}

		switch record.Value {
		case versionMarker:
			v := (int)(record.Timestamp)
			fmt.Printf("* Version=%d\n", v)
			if v != version {
				fmt.Printf("Error: data file has wrong version %d", v)
				return
			}

		case timeMarker:
			refTs = record.Timestamp
			prevTs = refTs
			fmt.Printf("* Ref timestamp=%d\n", refTs)

		default:
			fmt.Printf("%10d: %4d (+%d)\n", record.Timestamp, record.Value, record.Timestamp-prevTs)
			prevTs = record.Timestamp
		}

		count++
	}
}
