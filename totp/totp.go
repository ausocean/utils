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
	"crypto/hmac"
	"crypto/sha256"
	"strconv"
	"strings"
	"time"
)

// GenerateTOTP generates a time-based one-time (numeric) password as
// a string with the requested number of digits using the shared
// secret (i.e., similar to Google Authenticator). The maximum number
// of digits is 16, since we use SHA 256 for hashing, which returns 32
// bytes. The supplied time is truncated (floored) to the nearest
// minute. A common usage is to pass time.Now() as the time. However,
// if a minute boundary is crossed between password generation and
// password verification, the verifier may need to call GenerateTOTP
// again using an earlier time.
func GenerateTOTP(t time.Time, digits int, secret []byte) (string, error) {
	const maxDigits = 16
	if digits > maxDigits {
		digits = maxDigits
	}
	t = t.Truncate(time.Minute)
	tsBytes := []byte(strconv.Itoa(int(t.Unix())))
	hasher := hmac.New(sha256.New, secret)
	_, err := hasher.Write(append(secret, tsBytes...))
	if err != nil {
		return "", err
	}
	hashed := hasher.Sum(nil)
	nonce := make([]string, digits)
	// Emit one digit per two bytes.
	for i := 0; i < digits; i++ {
		nonce[i] = strconv.Itoa((int(hashed[2*i]) + int(hashed[2*i+1])) % 10)
	}
	return strings.Join(nonce, ""), nil
}
