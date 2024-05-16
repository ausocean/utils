/*
NAME
  filemap.go - manipulate maps stored in files.

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
  in gpl.txt. If not, see http://www.gnu.org/licenses.
*/

// Package filemap provides functions for manipulating maps stored in
// files. The major delimiter separates key-value pairs and the minor
// delimiter separates keys from corresponding values. For example,
// using newline as the major delimiter and space as the minor
// delimiter, the file containing the following:
//
//   key1 value1\nkey2 value2a,value2b,value2c\nkey3 \n"
//
// is represented as the following map:
//
//   {
//	"key1": "value1",
//	"key2": "value2a,value2b,value2c",
//	"key3": "",
//   }
//
// NB: Keys and values must not contain strings used as delimiters.
package filemap

import (
	"io/ioutil"
	"sort"
	"strings"
)

// KeyValue represents a key value pair.
type KeyValue struct {
	Key, Value string
}

// ReadFrom reads a map[string]string from a file. The major delimiter
// separates key-value pairs and the minor delimiter separates keys
// from values.
func ReadFrom(file, major, minor string) (map[string]string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return Split(string(content), major, minor), nil
}

// WriteTo writes a map[string]string to a file. The major delimiter
// separates key-value pairs and the minor delimiter separates keys
// from values. See Sort() for a description of the order.
func WriteTo(file, major, minor string, fm map[string]string, order []string) error {
	var content string

	for _, kv := range Sort(fm, order) {
		content += kv.Key + minor + kv.Value + major
	}
	return ioutil.WriteFile(file, []byte(content), 0666)
}

// Split splits a string twice into a map. The major delimiter
// separates key-value pairs and the minor delimiter separates keys
// from values.
func Split(str string, major string, minor string) map[string]string {
	fm := map[string]string{}
	if str == "" {
		return fm
	}
	for _, item := range strings.Split(str, major) {
		if len(item) == 0 {
			continue
		}
		ss := strings.Split(item, minor)
		if len(ss) == 0 {
			continue
		}
		if len(ss) == 1 {
			fm[ss[0]] = ""
		} else {
			fm[ss[0]] = ss[1]
		}
	}
	return fm
}

// Sort produces a sorted slice of the map's keys and values. If
// order is non-nil, it specifies which keys to write and their
// order. Writing a subset is therefore possible. If the order is nil,
// all keys are written in string sort order.
func Sort(fm map[string]string, order []string) []KeyValue {
	if order == nil {
		order = []string{}
		for k := range fm {
			order = append(order, k)
		}
		sort.Strings(order)
	}

	list := []KeyValue{}
	for _, k := range order {
		list = append(list, KeyValue{Key: k, Value: fm[k]})
	}
	return list
}

// IsValid checks that the value for the given key is valid.
// If delim is empty, the value must exactly match one of the supplied values.
// If delim is non-empty, the value is split by delim and each part must match one of the values.
// For example, consider the following map, fm:
//
//   {
//	"key1": "value1",
//	"key2": "value2a,value2b,value2c",
//	"key3": "",
//   }
//
// The following all return true:
//
//   IsValid(fm, "key1", []string{"value1"}, "")                         // Exact match.
//   IsValid(fm, "key1", []string{"value1", "value1b"}, ",")             // Matches one of the delimited values.
//   IsValid(fm, "key2", []string{"value2a", "value2b", "value2c"}, ",") // Matches all of the delimited values.
//
// The following all return false:
//
//   IsValid(fm, "key1", []string{"value1a"}, "")                        // Exact match fails.
//   IsValid(fm, "key1", []string{"value1a", "value1b"}, ",")            // Matches none of delimited values.
//
func IsValid(fm map[string]string, key string, values []string, delim string) bool {
	value := fm[key]
	if delim == "" {
		for _, v := range values {
			if v == value {
				return true
			}
		}
		return false
	}

	parts := strings.Split(value, delim)
	for _, part := range parts {
		matches := false
		for _, v := range values {
			if v == part {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}
	return true
}
