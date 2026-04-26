// Package api hosts the typed, OpenAPI-described HTTP layer. Operations are
// registered with huma so the spec, runtime validation and Go SDK stay in
// sync — see [Register] for the entry point that the router wires up.
//
// The package coexists with the legacy gin handlers in `internal/handler` and
// `internal/router`. As endpoints are migrated they're moved over here; the
// legacy versions are removed only after the typed counterpart ships.
package api

import (
	"encoding/json"

	"github.com/kite-plus/kite/internal/errcodes"
)

// Envelope is the shape every JSON response carries. Generics let each
// endpoint specialise the `data` payload while the wire format stays
// identical to what gin handlers have always returned, so existing clients
// (web admin, PicGo plugins, scripts) keep working.
type Envelope[T any] struct {
	Code    int    `json:"code" doc:"Business error code; 0 indicates success. See errcodes catalog."`
	Message string `json:"message" doc:"Human-readable description; localised in future revisions."`
	Data    T      `json:"data" doc:"Endpoint-specific payload. Null on error responses."`
}

// Ok wraps payload in a success envelope.
func Ok[T any](data T) Envelope[T] {
	return Envelope[T]{Code: int(errcodes.Success), Message: "success", Data: data}
}

// Page is the standard pagination wrapper for list endpoints.
type Page[T any] struct {
	Items []T   `json:"items" doc:"Items on this page."`
	Total int64 `json:"total" doc:"Total item count across all pages."`
	Page  int   `json:"page" doc:"Current page number, 1-indexed."`
	Size  int   `json:"size" doc:"Page size used for this query."`
}

// APIError is the wire-shape error that huma serialises for non-success
// outcomes. It satisfies huma.StatusError (via GetStatus) and
// json.Marshaler (so the body comes out matching the success envelope).
type APIError struct {
	Status int           `json:"-"` // HTTP status, used by huma to set the response code
	Code   errcodes.Code `json:"-"` // business error code, written into the JSON body
	Msg    string        `json:"-"` // human-readable message, written into the JSON body
}

// Error satisfies the standard error interface so APIError can flow through
// `if err != nil` like any other error value.
func (e *APIError) Error() string { return e.Msg }

// GetStatus is what huma calls to decide which HTTP status code to send.
func (e *APIError) GetStatus() int { return e.Status }

// MarshalJSON emits the {code, message, data: null} envelope. Huma's default
// would emit RFC 7807 Problem-Details JSON; we override so the wire shape
// matches the existing gin response format every client already handles.
func (e *APIError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}{
		Code:    int(e.Code),
		Message: e.Msg,
		Data:    nil,
	})
}

// Errf builds an APIError from an errcodes constant. The HTTP status is
// resolved from the catalog so callers don't have to remember the mapping.
func Errf(code errcodes.Code, format string, args ...any) *APIError {
	return &APIError{
		Status: errcodes.HTTPStatus(code),
		Code:   code,
		Msg:    sprintf(format, args...),
	}
}

// sprintf is a tiny indirection so we don't import fmt in the hot path of
// Errf when callers already pass a literal string.
func sprintf(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmtSprintf(format, args...)
}
