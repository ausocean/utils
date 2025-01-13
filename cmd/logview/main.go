/*
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

// logview is a simple web server that streams netsender logs from a Raspberry Pi via SSH.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	username = "pi"
	logFile  = "/var/log/netsender/netsender.log"
)

var raspberryPiAddr string
var password string

func init() {
	flag.StringVar(&raspberryPiAddr, "ip", "192.168.1.2:22", "Raspberry Pi IP address and port")
	flag.StringVar(&password, "password", "raspberry", "SSH password for the Raspberry Pi")
	flag.Parse()
}

// StreamLogs establishes an SSH connection and streams the log file
func StreamLogs(w io.Writer) {
	log.Println("starting log streaming")

	// Ensure that the writer implements http.Flusher.
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Fatal("writer does not support flushing")
	}

	// SSH client configuration.
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use proper host key checks in production.
		Timeout:         5 * time.Second,
	}

	// Connect to the Raspberry Pi.
	client, err := ssh.Dial("tcp", raspberryPiAddr, config)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// Start a session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("failed to get output pipe: %v", err)
	}

	// Run 'tail -f' to follow the log file.
	err = session.Start(fmt.Sprintf("tail -f %s", logFile))
	if err != nil {
		log.Fatalf("failed to run tail command: %v", err)
	}

	// Stream the output to the provided writer.
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)

		// Immediately flush the data to the client.
		flusher.Flush()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading output: %v", err)
	}
}

// HTTP Handler to serve the log stream. Once this handler is called, the connection will be kept open so the logs can be streamed.
func logHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain") // Set to plain text since we stream raw logs.
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	StreamLogs(w)
}

func main() {
	http.HandleFunc("/logs", logHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Log stream server started at http://localhost:8080/logs, go to http://localhost:8080 to view the logs.")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
