/*
AUTHORS

	Trek Hopton <trek@ausocean.org>

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

package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/gorilla/securecookie"
)

func main() {
	key := securecookie.GenerateRandomKey(32) // 256-bit key
	if key == nil {
		log.Fatal("Failed to generate session key")
	}
	fmt.Println("sessionKey:", base64.StdEncoding.EncodeToString(key))
}
