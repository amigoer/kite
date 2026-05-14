package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/amigoer/kite/internal/client"
)

func newRoomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "room",
		Short: "Manage rooms",
	}
	cmd.AddCommand(newRoomCreateCmd())
	cmd.AddCommand(newRoomListCmd())
	cmd.AddCommand(newRoomShowCmd())
	cmd.AddCommand(newRoomCloseCmd())
	return cmd
}

func newRoomCreateCmd() *cobra.Command {
	var (
		name  string
		cwd   string
		shell string
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new room",
		Example: `  kite room create
  kite room create --name release-test --cwd /tmp`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c := clientFromFlags(cmd)
			r, err := c.CreateRoom(cmd.Context(), client.CreateRoomRequest{
				Name:  name,
				Cwd:   cwd,
				Shell: shell,
			})
			if err != nil {
				return hintIfUnreachable(err)
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(r)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "room created: %s\n", r.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "  url: %s%s\n", clientFromFlags(cmd).BaseURL, r.URL)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "human-readable room name")
	cmd.Flags().StringVar(&cwd, "cwd", "", "initial working directory")
	cmd.Flags().StringVar(&shell, "shell", "", "shell binary (default /bin/bash)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return cmd
}

func newRoomListCmd() *cobra.Command {
	var (
		status string
		limit  int
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rooms",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c := clientFromFlags(cmd)
			rooms, err := c.ListRooms(cmd.Context(), client.ListRoomsOptions{Status: status, Limit: limit})
			if err != nil {
				return hintIfUnreachable(err)
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(rooms)
			}
			if len(rooms) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no rooms")
				return nil
			}
			rows := make([][]string, 0, len(rooms))
			for _, r := range rooms {
				rows = append(rows, []string{
					r.ID, r.Name, r.Status,
					fmt.Sprintf("%d", r.CommandCount),
					r.CreatedAt.Local().Format(time.RFC3339),
				})
			}
			printTable(cmd.OutOrStdout(), []string{"ID", "NAME", "STATUS", "CMDS", "CREATED"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "filter by status (active | closed)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max rooms to return")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return cmd
}

func newRoomShowCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one room and its command history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromFlags(cmd)
			r, err := c.GetRoom(cmd.Context(), args[0])
			if err != nil {
				return hintIfUnreachable(err)
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(r)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ID:        %s\n", r.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Name:      %s\n", r.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Status:    %s\n", r.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Cwd:       %s\n", r.Cwd)
			fmt.Fprintf(cmd.OutOrStdout(), "Shell:     %s\n", r.Shell)
			fmt.Fprintf(cmd.OutOrStdout(), "Created:   %s\n", r.CreatedAt.Local().Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "Commands:  %d\n", r.CommandCount)
			fmt.Fprintf(cmd.OutOrStdout(), "URL:       %s%s\n", clientFromFlags(cmd).BaseURL, r.URL)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return cmd
}

func newRoomCloseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close <id>",
		Short: "Close a room and terminate its shell",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromFlags(cmd)
			if err := c.CloseRoom(cmd.Context(), args[0]); err != nil {
				return hintIfUnreachable(err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "room closed: %s\n", args[0])
			return nil
		},
	}
}
