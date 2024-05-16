/*
AUTHORS
  Trek Hopton <trek@ausocean.org>

LICENSE
  Copyright (C) 2020-2021 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt. If not, see http://www.gnu.org/licenses.
*/

// link-logger is a program that monitors and logs the information for a specific link on a wireless device.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ausocean/utils/link"
)

func main() {
	user := flag.String("user", "root", "Username for remote machine.")
	pass := flag.String("password", "admin", "Password for given user on remote machine.")
	device := flag.String("device", "wlan0", "Wireless interface / device name of link to test.")
	port := flag.String("port", "22", "Port number for SSH connection.")
	outFile := flag.String("output", "signal-strength", "Prefix of the output file to be created and written to.")
	callInterval := flag.Duration("interval", 0, "How long to wait between command calls. An interval of 0 means it will call as fast as it can without waiting.")
	ip := flag.String("ip", "192.168.1.1", "Host's IP address.")
	flag.Parse()

	l, err := link.New(*device, *ip, *port, *user, *pass)
	if err != nil {
		log.Fatalf("could not create link %v", err)
	}

	// Create a file for writing output.
	f, err := os.Create(*outFile + time.Now().Format(time.RFC3339) + ".csv")
	if err != nil {
		log.Fatalf("could not create output file: %v", err)
	}

	for {
		err := l.Update()
		if err != nil {
			log.Fatalf("could not update link information: %v", err)
		}

		// Print relevant info to console and write to CSV file.
		log.Printf("quality: %v, signal: %v, noise: %v, bitrate: %v", l.Quality(), l.Signal(), l.Noise(), l.Bitrate())
		_, err = fmt.Fprintf(f, "%s,%v,%v,%v,%v\n", time.Now().String(), l.Quality(), l.Signal(), l.Noise(), l.Bitrate())
		if err != nil {
			log.Fatalf("could not write to output file: %v", err)
		}

		time.Sleep((time.Duration(*callInterval) * time.Millisecond))
	}
}
