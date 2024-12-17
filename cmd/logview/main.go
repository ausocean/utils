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

// SSH Configuration
const (
	username = "pi"                               // Replace with your username
	logFile  = "/var/log/netsender/netsender.log" // Replace with the log file you want to stream
)

var raspberryPiAddr string
var password string

func init() {
	// Command-line flags for IP and password
	flag.StringVar(&raspberryPiAddr, "ip", "192.168.1.117:22", "Raspberry Pi IP address and port")
	flag.StringVar(&password, "password", "raspberry", "SSH password for the Raspberry Pi")
	flag.Parse() // Parse command-line arguments
}

// StreamLogs establishes an SSH connection and streams the log file
func StreamLogs(w io.Writer) {

	log.Println("Starting log streaming")

	// Ensure that the writer implements http.Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Fatal("Writer does not support flushing")
	}

	// SSH client configuration
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Use proper host key checks in production
		Timeout:         5 * time.Second,
	}

	// Connect to the Raspberry Pi
	client, err := ssh.Dial("tcp", raspberryPiAddr, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Start a session
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Run 'tail -f' to follow the log file
	output, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get output pipe: %v", err)
	}

	err = session.Start(fmt.Sprintf("tail -f %s", logFile))
	if err != nil {
		log.Fatalf("Failed to run tail command: %v", err)
	}

	// Stream the output to the provided writer
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)

		// Immediately flush the data to the client
		flusher.Flush()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading output: %v", err)
	}
}

// HTTP Handler to serve the log stream
func logHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for streaming (but not specific to EventSource)
	w.Header().Set("Content-Type", "text/plain") // Set to plain text since we stream raw logs
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Stream the logs
	StreamLogs(w)
}

func main() {
	// HTTP server to serve logs
	http.HandleFunc("/logs", logHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Log stream server started at http://localhost:8080/logs")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
