package cmd

import (
	"fmt"

	"github.com/hyperlab-be/ralph/internal/loop"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all loops",
	Long:    `List all registered loops with their status.`,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	loops, err := loop.ListAll()
	if err != nil {
		return fmt.Errorf("failed to list loops: %w", err)
	}

	if len(loops) == 0 {
		fmt.Println("No loops registered.")
		return nil
	}

	for _, l := range loops {
		status := loop.GetStatus(l)
		icon := "⚫"
		if status == "running" {
			icon = "🟢"
		}
		fmt.Printf("%s %s\n", icon, l.Name)
	}

	return nil
}
