package server

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"net/http"
	"strings"
)

// randomShortID returns a lowercase base32 token suitable for tagging
// transient daemon-side handles (e.g. attach holder IDs).
func randomShortID() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("kite/server: rand failed: " + err.Error())
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b[:]))
}

// APIError is the shape every JSON error response uses.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	Error APIError `json:"error"`
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: APIError{Code: code, Message: message}})
}

// writeJSON sends a JSON success response.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
