package room

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

const idLen = 12

// NewRoomID returns a fresh room ID, e.g. "r_abc234def567".
func NewRoomID() string {
	return "r_" + randBase32(idLen)
}

// NewCommandID returns a fresh command ID, e.g. "c_abc234def567".
func NewCommandID() string {
	return "c_" + randBase32(idLen)
}

// NewParticipantID returns a fresh participant ID.
func NewParticipantID() string {
	return "p_" + randBase32(idLen)
}

func randBase32(n int) string {
	bytes := make([]byte, (n*5+7)/8)
	if _, err := rand.Read(bytes); err != nil {
		panic("kite: crypto/rand failed: " + err.Error())
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	enc = strings.ToLower(enc)
	if len(enc) > n {
		enc = enc[:n]
	}
	return enc
}
