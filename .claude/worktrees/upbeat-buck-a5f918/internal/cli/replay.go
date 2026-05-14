package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/client"
	"github.com/amigoer/kite/internal/room"
)

func newReplayCmd() *cobra.Command {
	var (
		speed    float64
		noTiming bool
		search   string
	)
	cmd := &cobra.Command{
		Use:   "replay <room_id>",
		Short: "Replay the event timeline of a room in your terminal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roomID := args[0]
			c := clientFromFlags(cmd)

			events, _, err := c.GetEvents(cmd.Context(), roomID, client.GetEventsOptions{Limit: 100000})
			if err != nil {
				return hintIfUnreachable(err)
			}
			return playEvents(os.Stdout, events, speed, noTiming, search)
		},
	}
	cmd.Flags().Float64Var(&speed, "speed", 1.0, "playback speed multiplier (0 = instant)")
	cmd.Flags().BoolVar(&noTiming, "no-timing", false, "skip the inter-event delay")
	cmd.Flags().StringVar(&search, "search", "", "only replay commands matching this substring")
	return cmd
}

func playEvents(w io.Writer, events []*room.Event, speed float64, noTiming bool, search string) error {
	if len(events) == 0 {
		fmt.Fprintln(os.Stderr, "no events")
		return nil
	}
	matchID := map[string]bool{}
	var lastTS time.Time
	for _, ev := range events {
		switch ev.Type {
		case room.EvtCommandStarted:
			var p room.CommandStartedPayload
			_ = json.Unmarshal(ev.Payload, &p)
			if search != "" && !strings.Contains(p.Cmd, search) {
				continue
			}
			matchID[p.CommandID] = true
			fmt.Fprintf(w, "\n$ %s\n", p.Cmd)
		case room.EvtCommandOutput:
			var p room.CommandOutputPayload
			_ = json.Unmarshal(ev.Payload, &p)
			if !matchID[p.CommandID] {
				continue
			}
			if !noTiming && !lastTS.IsZero() && speed > 0 {
				delay := ev.Timestamp.Sub(lastTS)
				if delay > 0 {
					time.Sleep(time.Duration(float64(delay) / speed))
				}
			}
			_, _ = w.Write(p.Data)
		case room.EvtCommandFinished:
			var p room.CommandFinishedPayload
			_ = json.Unmarshal(ev.Payload, &p)
			if !matchID[p.CommandID] {
				continue
			}
			fmt.Fprintf(w, "[exit %d, %dms]\n", p.ExitCode, p.DurationMs)
		}
		lastTS = ev.Timestamp
	}
	return nil
}
