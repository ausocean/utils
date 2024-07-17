/*
AUTHORS
  Alan Noble <alan@ausocean.org>

LICENSE
  Copyright (C) 2024 the Australian Ocean Lab (AusOcean). All Rights Reserved.

  The Software and all intellectual property rights associated
  therewith, including but not limited to copyrights, trademarks,
  patents, and trade secrets, are and will remain the exclusive
  property of the Australian Ocean Lab (AusOcean).
*/

package totp

import (
	"testing"
	"time"
)

const secret = "not-so-secret"

func TestGenerateTOTP(t *testing.T) {
	nye2018 := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)

	var tests = []struct {
		t      time.Time
		digits int
		want   string
	}{
		{
			digits: 10,
			want:   "9593791354",
		},
		{
			digits: 16,
			want:   "9593791354356707",
		},
		{
			t:      nye2018,
			digits: 16,
			want:   "6646979953861765",
		},
		{
			t:      nye2018.Add(time.Duration(59) * time.Second),
			digits: 16,
			want:   "6646979953861765",
		},
	}

	for i, test := range tests {
		totp, err := GenerateTOTP(test.t, test.digits, []byte(secret))
		if err != nil {
			t.Fatalf("GenerateTOTP returned unexpected error: %v", err)
		}
		if len(totp) != test.digits {
			t.Errorf("%d: expected %d digits, got %d", i, test.digits, len(totp))
		}
		if totp != test.want {
			t.Errorf("%d: expected %s TOTP, got %s", i, test.want, totp)
		}
	}
}

func TestCheckTOTP(t *testing.T) {
	const digits = 16

	now := time.Now()
	var tests = []struct {
		tm     time.Time
		period time.Duration
		want   bool
	}{
		{
			tm:     now,
			period: time.Duration(2 * time.Minute),
			want:   true,
		},
		{
			tm:     now,
			period: time.Duration(time.Minute),
			want:   true,
		},
		{
			tm:     now,
			period: time.Duration(0),
			want:   true,
		},
		{
			tm:     now.Add(2 * time.Minute),
			period: time.Duration(0),
			want:   false,
		},
	}

	for i, test := range tests {
		totp, err := GenerateTOTP(now, digits, []byte(secret))
		if err != nil {
			t.Fatalf("%d: GenerateTOTP returned unexpected error: %v", i, err)
		}
		ok, err := CheckTOTP(totp, test.tm, test.period, digits, []byte(secret))
		if err != nil {
			t.Fatalf("%d: CheckTOTP returned unexpected error: %v", i, err)
		}
		if ok != test.want {
			t.Errorf("%d: expected %t got %t", i, test.want, ok)
		}
	}
}
