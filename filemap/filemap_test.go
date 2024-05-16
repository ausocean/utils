/*
NAME
  filemap.go - functions for working with maps stored in files.

AUTHOR
  Alan Noble <alann@ausocean.org>

LICENSE
  filemap.go is Copyright (C) 2019 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  in gpl.txt.  If not, see http://www.gnu.org/licenses.
*/

package filemap

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

// TestFilemap tests reading/writing maps from/to files.
func TestFilemap(t *testing.T) {
	// Create a temporary file
	f, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Errorf("iotutil.TempFile failed with error %v", err)
	}
	file := f.Name()
	defer os.Remove(file)

	// Test reading an empty file
	fm, err := ReadFrom(file, "\n", " ")
	if err != nil {
		t.Errorf("ReadFrom returned with error %v", err)
	}
	if len(fm) != 0 {
		t.Errorf("ReadFrom returned a non-empty map %v", err)
	}

	// Test writing in default (sorted) order.
	fm2 := map[string]string{
		"key1": "value1",
		"key2": "value2a,value2b,value2c",
		"key3": "",
	}
	err = WriteTo(file, "\n", " ", fm2, nil)
	if err != nil {
		t.Errorf("WriteTo returned with error %v", err)
	}
	fm, err = ReadFrom(file, "\n", " ")
	if err != nil {
		t.Errorf("ReadFrom returned with error %v", err)
	}
	if !reflect.DeepEqual(fm, fm2) {
		t.Errorf("ReadFrom did not return the correct map")
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		t.Errorf("ioutil.ReadFile returned with error %v", err)
	}
	if string(content) != "key1 value1\nkey2 value2a,value2b,value2c\nkey3 \n" {
		t.Errorf("Order is wrong")
	}

	// Test writing in specified (reverse) order.
	err = WriteTo(file, "\n", " ", fm2, []string{"key3", "key2", "key1"})
	if err != nil {
		t.Errorf("WriteTo returned with error %v", err)
	}
	content, err = ioutil.ReadFile(file)
	if err != nil {
		t.Errorf("ioutil.ReadFile returned with error %v", err)
	}
	if string(content) != "key3 \nkey2 value2a,value2b,value2c\nkey1 value1\n" {
		t.Errorf("Order is wrong")
	}

	// Test validity
	if !IsValid(fm2, "key1", []string{"value1"}, "") {
		t.Errorf("IsValid #1 failed")
	}
	if !IsValid(fm2, "key1", []string{"value1", "value2"}, "") {
		t.Errorf("IsValid #2 failed")
	}
	if IsValid(fm2, "key1", []string{"value2"}, "") {
		t.Errorf("IsValid #3 failed")
	}
	if !IsValid(fm2, "key2", []string{"value2a", "value2b", "value2c"}, ",") {
		t.Errorf("IsValid #4 failed")
	}
	if IsValid(fm2, "key2", []string{"value2a", "value2b", "value2d"}, ",") {
		t.Errorf("IsValid #5 failed")
	}
	if IsValid(fm2, "key3", []string{"value3"}, "") {
		t.Errorf("IsValid #6 failed")
	}
}
