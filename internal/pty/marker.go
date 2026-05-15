// Package pty owns the persistent bash process backing a room and the
// marker-based command-boundary protocol described in SPEC §4.
package pty

import "regexp"

// MarkerEnd is the literal marker template emitted by per-Exec calls. The
// protocol writes
//
//	printf '\n__KITE_END_%d_<command_id>__\n' $?
//
// to the bash stdin so it emits this string once the user command exits, and
// the read loop scans bytes for the regex below.
const MarkerEnd = "__KITE_END_"

// MarkerPrompt is the literal sentinel emitted by PROMPT_COMMAND. Bash fires
// PROMPT_COMMAND right before redrawing the prompt, so this sentinel marks
// "shell is idle, last command (whatever it was) has returned". Used by
// callers that need to know when a human-typed command finished, even though
// no Exec wrapped it.
const MarkerPrompt = "__KITE_PROMPT__"

// markerRe matches either an end-of-exec marker (with exit code + command
// id) or a prompt sentinel. Match groups:
//
//	[1] non-empty when matching MarkerEnd
//	[2] exit code (only when [1] matched)
//	[3] command id (only when [1] matched)
//	[4] non-empty when matching MarkerPrompt
var markerRe = regexp.MustCompile(`(__KITE_END_(-?\d+)_(c_[a-z2-7]{12})__)|(__KITE_PROMPT__)`)
