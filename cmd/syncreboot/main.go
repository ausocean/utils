/*
AUTHORS
  Alan Noble <alan@ausocean.org>
  Trek Hopton <trek@ausocean.org>

LICENSE
  Copyright (C) 2020 - 2024 the Australian Ocean Lab (AusOcean)

  This file is part of VidGrind. VidGrind is free software: you can
  redistribute it and/or modify it under the terms of the GNU
  General Public License as published by the Free Software
  Foundation, either version 3 of the License, or (at your option)
  any later version.

  VidGrind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with NetReceiver in gpl.txt.  If not, see
  <http://www.gnu.org/licenses/>.
*/

// syncreboot is a simple program to sync and reboot.
// It requires reboot capability in order to execute.
// See the accompanying Makefile.
// If run with the -s flag, it will simply shutdown.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
)

// Current software version.
const version = "v1.0.0"

func main() {
	showVersion := flag.Bool("version", false, "show version")
	shutdownPtr := flag.Bool("s", false, "shutdown system")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	syscall.Sync()
	if *shutdownPtr {
		err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
		if err != nil {
			log.Fatalf("shutdown error: %v", err)
			os.Exit(1)
		}
		return
	}

	err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	if err != nil {
		log.Fatalf("reboot error: %v", err)
		os.Exit(1)
	}
}
