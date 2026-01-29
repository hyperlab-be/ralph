package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperlab/rl/internal/config"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:     "logs [name]",
	Aliases: []string{"log", "l"},
	Short:   "View logs of a loop",
	Long:    `View the session logs of a loop. Use -f to follow (tail) the logs.`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runLogs,
}

var followLogs bool
var numLines int

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow logs in real-time")
	logsCmd.Flags().IntVarP(&numLines, "lines", "n", 50, "Number of lines to show")
	rootCmd.AddCommand(logsCmd)
}

func runLogs(cmd *cobra.Command, args []string) error {
	var projectRoot string

	if len(args) > 0 {
		loopName := args[0]
		loop, err := config.GetLoop(loopName)
		if err != nil {
			return fmt.Errorf("failed to get loop: %w", err)
		}
		if loop == nil {
			return fmt.Errorf("loop not found: %s", loopName)
		}
		projectRoot = loop.Path
	} else {
		cwd, _ := os.Getwd()
		var err error
		projectRoot, err = config.FindProjectRoot(cwd)
		if err != nil {
			return fmt.Errorf("not in a rl project and no loop name provided")
		}
	}

	logFile := filepath.Join(projectRoot, ".rl", "session.log")

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		printWarn("No logs found")
		return nil
	}

	if followLogs {
		return tailFollow(logFile)
	}

	return tailLast(logFile, numLines)
}

func tailLast(filename string, n int) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines (simple approach for now)
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Print last n lines
	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	for i := start; i < len(lines); i++ {
		fmt.Println(lines[i])
	}

	return nil
}

func tailFollow(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Seek to end
	file.Seek(0, io.SeekEnd)

	printInfo(fmt.Sprintf("Following %s (Ctrl+C to stop)", filename))
	fmt.Println()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return err
		}
		fmt.Print(line)
	}
}
