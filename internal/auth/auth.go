package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// randomHex returns n random bytes as a hex string (2 characters per byte).
// crypto/rand is used, not math/rand, because these values are secrets
// or must be unguessable: math/rand is predictable, crypto/rand is not.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateStreamKey creates the SECRET key a broadcaster puts in OBS.
// Anyone who has it can stream as that user, so it is never shown publicly.
func GenerateStreamKey() (string, error) {
	h, err := randomHex(16) // 16 bytes = 32 hex characters
	if err != nil {
		return "", err
	}
	return "live_" + h, nil
}

// GeneratePlaybackID creates the PUBLIC id viewers use in the watch URL.
// It is separate from the stream key so the key never has to be exposed.
func GeneratePlaybackID() (string, error) {
	h, err := randomHex(8)
	if err != nil {
		return "", err
	}
	return "pb_" + h, nil
}
