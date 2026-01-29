package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system dependencies",
	Long:  `Verify that all required tools are installed and configured correctly.`,
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("\033[1m\033[36mChecking dependencies...\033[0m")
	fmt.Println()

	allGood := true

	// Check git
	if _, err := exec.LookPath("git"); err != nil {
		printError("git: not found")
		fmt.Println("  Install: https://git-scm.com/downloads")
		allGood = false
	} else {
		out, _ := exec.Command("git", "--version").Output()
		printSuccess(fmt.Sprintf("git: %s", string(out[:len(out)-1])))
	}

	// Check Claude CLI
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		printError("claude: not found")
		fmt.Println("  Install: npm install -g @anthropic-ai/claude-code")
		allGood = false
	} else {
		out, _ := exec.Command(claudePath, "--version").Output()
		if len(out) > 0 {
			printSuccess(fmt.Sprintf("claude: %s", string(out[:len(out)-1])))
		} else {
			printSuccess(fmt.Sprintf("claude: found at %s", claudePath))
		}
	}

	// Check gh CLI (for PR creation)
	if _, err := exec.LookPath("gh"); err != nil {
		printWarn("gh: not found (optional, needed for auto PR creation)")
		fmt.Println("  Install: https://cli.github.com")
	} else {
		out, _ := exec.Command("gh", "--version").Output()
		lines := string(out)
		if idx := len(lines); idx > 0 {
			// Get first line only
			if newline := strings.Index(lines, "\n"); newline > 0 {
				lines = lines[:newline]
			}
			printSuccess(fmt.Sprintf("gh: %s", lines))
		}
	}

	fmt.Println()

	if allGood {
		printSuccess("All required dependencies installed!")
		return nil
	}

	return fmt.Errorf("some dependencies are missing")
}
