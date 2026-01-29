package cmd

import (
	"fmt"
	"os/exec"

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

	// Check mysql (optional)
	if _, err := exec.LookPath("mysql"); err != nil {
		printWarn("mysql: not found (optional, needed for database hooks)")
	} else {
		out, _ := exec.Command("mysql", "--version").Output()
		printSuccess(fmt.Sprintf("mysql: %s", string(out[:len(out)-1])))
	}

	fmt.Println()

	if allGood {
		printSuccess("All required dependencies installed!")
		return nil
	}

	return fmt.Errorf("some dependencies are missing")
}
