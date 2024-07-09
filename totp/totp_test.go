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

func TestTOTP(t *testing.T) {
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
