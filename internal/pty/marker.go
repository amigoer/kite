// Package pty owns the persistent bash process backing a room and the
// marker-based command-boundary protocol described in SPEC §4.
package pty

import "regexp"

// MarkerEnd is the literal marker template. The protocol writes
//
//	printf '\n__KITE_END_%d_<command_id>__\n' $?
//
// to the bash stdin so it emits this string once the user command exits, and
// the read loop scans bytes for the regex below.
const MarkerEnd = "__KITE_END_"

// markerRe captures (exit_code, command_id).
var markerRe = regexp.MustCompile(`__KITE_END_(-?\d+)_(c_[a-z2-7]{12})__`)
