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

// Package link provides an abstraction of a wireless link.
package link

import (
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// Link represents a single link/connection from one remote wireless device to another.
type Link struct {
	conn *ssh.Client
	dev  string
	info map[string]interface{}
}

// New creates a Link and opens the required SSH connection to the remote device.
func New(device, ip, port, user, pass string) (*Link, error) {
	l := Link{dev: device}
	var err error
	l.conn, err = openSSHConnection(ip+":"+port, user, pass)
	if err != nil {
		return nil, fmt.Errorf("could not open SSH connection for link: %w", err)
	}
	return &l, nil
}

// Update gets the latests link information from the remote device via a call over SSH.
func (l *Link) Update() error {
	cmd := "ubus call iwinfo info '{ \"device\": \"" + l.dev + "\" }'\n"

	// Start a session to execute a command.
	session, err := l.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to begin session: %w", err)
	}

	// Execute command.
	bytes, err := session.Output(cmd)
	if err != nil {
		return fmt.Errorf("could not get command output: %w", err)
	}

	session.Close()

	var v interface{}
	err = json.Unmarshal(bytes, &v)
	if err != nil {
		return fmt.Errorf("could not unmarshal json: %w", err)
	}
	data := v.(map[string]interface{})

	l.info = data

	return nil
}

// SSID returns the Link's SSID.
func (l *Link) SSID() string {
	return l.info["ssid"].(string)
}

// MAC returns the Link's MAC.
func (l *Link) MAC() string {
	return l.info["bssid"].(string)
}

// Mode returns the Link's Mode.
func (l *Link) Mode() string {
	return l.info["mode"].(string)
}

// Signal returns the Link's Signal.
func (l *Link) Signal() int {
	return int(l.info["signal"].(float64))
}

// Quality returns the Link's Quality.
func (l *Link) Quality() int {
	return int(l.info["quality"].(float64))
}

// Noise returns the Link's Noise.
func (l *Link) Noise() int {
	return int(l.info["noise"].(float64))
}

// Bitrate returns the Link's Bitrate.
func (l *Link) Bitrate() int {
	return int(l.info["bitrate"].(float64))
}

// Channel returns the Link's Channel.
func (l *Link) Channel() int {
	return int(l.info["channel"].(float64))
}

func openSSHConnection(addr, user, pass string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial target: %w", err)
	}
	return conn, nil
}
