/*
NAME
  ring_test.go - a test suite adopting the golang testing library to test functionality of the
  RingBuffer structure

DESCRIPTION
  See README.md

AUTHOR
  Dan Kortschak <dan@ausocean.org>

LICENSE
  ring_test.go is Copyright (C) 2017 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
 for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt.  If not, see http://www.gnu.org/licenses.
*/

package ring

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

var roundTripTests = []struct {
	name string

	len         int
	size        int
	timeout     time.Duration
	nextTimeout time.Duration

	data       [][]string
	readDelay  time.Duration
	writeDelay time.Duration
}{
	{
		name: "happy",
		len:  2, size: 50,
		timeout:     100 * time.Millisecond,
		nextTimeout: 100 * time.Millisecond,

		data: [][]string{
			{"frame1", "frame2", "frame3", "frame4"},
			{"frame5", "frame6"},
			{"frame5", "frame6", "frame7"},
			{"frame8", "frame9", "frame10"},
			{"frame11"},
			{"frame12", "frame13"},
			{"frame14", "frame15", "frame16", "frame17"},
		},
	},
	{
		name: "slow write",
		len:  2, size: 50,
		timeout:     100 * time.Millisecond,
		nextTimeout: 100 * time.Millisecond,

		data: [][]string{
			{"frame1", "frame2", "frame3", "frame4"},
			{"frame5", "frame6"},
			{"frame5", "frame6", "frame7"},
			{"frame8", "frame9", "frame10"},
			{"frame11"},
			{"frame12", "frame13"},
			{"frame14", "frame15", "frame16", "frame17"},
		},
		writeDelay: 500 * time.Millisecond,
	},
	{
		name: "slow read",
		len:  2, size: 50,
		timeout:     100 * time.Millisecond,
		nextTimeout: 100 * time.Millisecond,

		data: [][]string{
			{"frame1", "frame2", "frame3", "frame4"},
			{"frame5", "frame6"},
			{"frame5", "frame6", "frame7"},
			{"frame8", "frame9", "frame10"},
			{"frame11"},
			{"frame12", "frame13"},
			{"frame14", "frame15", "frame16", "frame17"},
		},
		readDelay: 500 * time.Millisecond,
	},
}

func TestRoundTrip(t *testing.T) {
	const maxTimeouts = 100
	for _, test := range roundTripTests {
		b := NewBuffer(test.len, test.size, test.timeout)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for _, c := range test.data {
				var dropped int
				for _, f := range c {
					time.Sleep(test.writeDelay) // Simulate slow data capture.
					_, err := b.Write([]byte(f))
					switch err {
					case nil:
						dropped = 0
					case ErrDropped:
						if dropped > maxTimeouts {
							t.Errorf("too many write drops for %q", test.name)
							return
						}
						dropped++
					default:
						t.Errorf("unexpected write error for %q: %v", test.name, err)
						return
					}
				}
				b.Flush()
			}
			b.Close()
		}()
		go func() {
			buf := make([]byte, 1<<10)
			defer wg.Done()
			var got []string
			var timeouts int
		elements:
			for {
				_, err := b.Next(test.nextTimeout)
				switch err {
				case nil:
					timeouts = 0
				case ErrTimeout:
					if timeouts > maxTimeouts {
						t.Errorf("too many timeouts for %q", test.name)
						return
					}
					timeouts++
				case io.EOF:
					break elements
				default:
					t.Errorf("unexpected read error for %q: %v", test.name, err)
					return
				}
			reads:
				for {
					n, err := b.Read(buf)
					if n != 0 {
						time.Sleep(test.readDelay) // Simulate slow data processing.
						got = append(got, string(buf[:n]))
					}
					switch err {
					case nil:
					case io.EOF:
						break reads
					default:
						t.Errorf("unexpected read error for %q: %v", test.name, err)
						return
					}
				}
			}
			var want []string
			for _, c := range test.data {
				want = append(want, strings.Join(c, ""))
			}
			if test.readDelay == 0 {
				if !reflect.DeepEqual(got, want) {
					t.Errorf("unexpected round-trip result for %q:\ngot: %#v\nwant:%#v", test.name, got, want)
				}
			} else {
				// We may have dropped writes in this case.
				// So just check that we can consume every
				// received element with reference to what
				// was sent.
				// TODO(kortschak): Check that the number of
				// missing elements matches the number of
				// dropped writes.
				var sidx, ridx int
				var recd string
				for ridx, recd = range got {
					for ; sidx < len(want); sidx++ {
						if recd == want[sidx] {
							break
						}
					}
				}
				if ridx != len(got)-1 {
					t.Errorf("unexpected round-trip result for %q (unexplained element received):\ngot: %#v\nwant:%#v", test.name, got, want)
				}
			}
		}()
		wg.Wait()
	}
}

func TestRoundTripWriterTo(t *testing.T) {
	const maxTimeouts = 100
	for _, test := range roundTripTests {
		b := NewBuffer(test.len, test.size, test.timeout)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for _, c := range test.data {
				var dropped int
				for _, f := range c {
					time.Sleep(test.writeDelay) // Simulate slow data capture.
					_, err := b.Write([]byte(f))
					switch err {
					case nil:
						dropped = 0
					case ErrDropped:
						if dropped > maxTimeouts {
							t.Errorf("too many write drops for %q", test.name)
							return
						}
						dropped++
					default:
						t.Errorf("unexpected write error for %q: %v", test.name, err)
						return
					}
				}
				b.Flush()
			}
			b.Close()
		}()
		go func() {
			var buf bytes.Buffer
			defer wg.Done()
			var got []string
			var timeouts int
		elements:
			for {
				chunk, err := b.Next(test.nextTimeout)
				switch err {
				case nil:
					timeouts = 0
				case ErrTimeout:
					if timeouts > maxTimeouts {
						t.Errorf("too many timeouts for %q", test.name)
						return
					}
					timeouts++
					continue
				case io.EOF:
					break elements
				default:
					t.Errorf("unexpected read error for %q: %v", test.name, err)
					return
				}

				n, err := chunk.WriteTo(&buf)
				if n != 0 {
					time.Sleep(test.readDelay) // Simulate slow data processing.
					got = append(got, buf.String())
					buf.Reset()
				}
				if err != nil {
					t.Errorf("unexpected writeto error for %q: %v", test.name, err)
					return
				}
				err = chunk.Close()
				if err != nil {
					t.Errorf("unexpected close error for %q: %v", test.name, err)
					return
				}
			}
			var want []string
			for _, c := range test.data {
				want = append(want, strings.Join(c, ""))
			}
			if test.readDelay == 0 {
				if !reflect.DeepEqual(got, want) {
					t.Errorf("unexpected round-trip result for %q:\ngot: %#v\nwant:%#v", test.name, got, want)
				}
			} else {
				// We may have dropped writes in this case.
				// So just check that we can consume every
				// received element with reference to what
				// was sent.
				// TODO(kortschak): Check that the number of
				// missing elements matches the number of
				// dropped writes.
				var sidx, ridx int
				var recd string
				for ridx, recd = range got {
					for ; sidx < len(want); sidx++ {
						if recd == want[sidx] {
							break
						}
					}
				}
				if ridx != len(got)-1 {
					t.Errorf("unexpected round-trip result for %q (unexplained element received):\ngot: %#v\nwant:%#v", test.name, got, want)
				}
			}
		}()
		wg.Wait()
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	const (
		maxTimeouts = 100

		len     = 50
		size    = 150e3
		timeout = 10 * time.Millisecond

		frameLen = 30e3

		writeDelay = 20 * time.Millisecond
		readDelay  = 50 * time.Millisecond
	)

	// Allocated prior to timer reset since it is an
	// amortised cost.
	rb := NewBuffer(len, size, timeout)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var timeouts int
	elements:
		for {
			_, err := rb.Next(timeout)
			switch err {
			case nil:
				timeouts = 0
			case ErrTimeout:
				if timeouts > maxTimeouts {
					b.Error("too many timeouts")
					return
				}
				timeouts++
			case io.EOF:
				break elements
			default:
				b.Errorf("unexpected read error: %v", err)
				return
			}

			_, err = ioutil.ReadAll(rb)
			time.Sleep(readDelay) // Simulate slow data processing.
			if err != nil {
				b.Errorf("unexpected read error: %v", err)
				return
			}
		}
	}()

	data := make([]byte, frameLen)

	b.ResetTimer()
	b.SetBytes(frameLen)

	var dropped int
	for i := 0; i < b.N; i++ {
		time.Sleep(writeDelay) // Simulate slow data capture.
		_, err := rb.Write(data)
		switch err {
		case nil:
			dropped = 0
		case ErrDropped:
			if dropped > maxTimeouts {
				b.Error("too many write drops")
				return
			}
			dropped++
		default:
			b.Errorf("unexpected write error: %v", err)
			return
		}
	}

	rb.Close()

	wg.Wait()
}

func BenchmarkRoundTripWriterTo(b *testing.B) {
	const (
		maxTimeouts = 100

		len     = 50
		size    = 150e3
		timeout = 10 * time.Millisecond

		frameLen = 30e3

		writeDelay = 20 * time.Millisecond
		readDelay  = 50 * time.Millisecond
	)

	// Allocated prior to timer reset since it is an
	// amortised cost.
	rb := NewBuffer(len, size, timeout)

	// This is hoisted here to ensure the allocation
	// is not counted since this is outside the control
	// of the ring buffer.
	buf := bytes.NewBuffer(make([]byte, 0, size+1))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var timeouts int
	elements:
		for {
			chunk, err := rb.Next(timeout)
			switch err {
			case nil:
				timeouts = 0
			case ErrTimeout:
				if timeouts > maxTimeouts {
					b.Error("too many timeouts")
					return
				}
				timeouts++
				continue
			case io.EOF:
				break elements
			default:
				b.Errorf("unexpected read error: %v", err)
				return
			}

			n, err := chunk.WriteTo(buf)
			if n != 0 {
				time.Sleep(readDelay) // Simulate slow data processing.
				buf.Reset()
			}
			if err != nil {
				b.Errorf("unexpected writeto error: %v", err)
				return
			}
			err = chunk.Close()
			if err != nil {
				b.Errorf("unexpected close error: %v", err)
				return
			}
		}
	}()

	data := make([]byte, frameLen)

	b.ResetTimer()
	b.SetBytes(frameLen)

	var dropped int
	for i := 0; i < b.N; i++ {
		time.Sleep(writeDelay) // Simulate slow data capture.
		_, err := rb.Write(data)
		switch err {
		case nil:
			dropped = 0
		case ErrDropped:
			if dropped > maxTimeouts {
				b.Error("too many write drops")
				return
			}
			dropped++
		default:
			b.Errorf("unexpected write error: %v", err)
			return
		}
	}

	rb.Close()

	wg.Wait()
}
