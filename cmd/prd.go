package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hyperlab-be/ralph/internal/config"
	"github.com/hyperlab-be/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var prdCmd = &cobra.Command{
	Use:     "prd [story title]",
	Aliases: []string{"p"},
	Short:   "View PRD or add a story",
	Long: `View the PRD status or add a new story.

Without arguments: shows PRD status
With a story title: adds a new story

Examples:
  ralph prd                           # Show PRD status
  ralph prd "Add user authentication" # Add a story
  ralph prd --new                     # Create new PRD interactively
  ralph prd --edit                    # Edit PRD in $EDITOR`,
	RunE: runPrd,
}

var (
	prdNew      bool
	prdEdit     bool
	storyCriteria []string
)

func init() {
	prdCmd.Flags().BoolVarP(&prdNew, "new", "n", false, "Create a new PRD")
	prdCmd.Flags().BoolVarP(&prdEdit, "edit", "e", false, "Edit PRD in $EDITOR")
	prdCmd.Flags().StringArrayVarP(&storyCriteria, "criteria", "c", nil, "Acceptance criteria (can be repeated)")
	rootCmd.AddCommand(prdCmd)
}

func runPrd(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a ralph project. Run 'ralph init' first")
	}

	// --new flag: create new PRD
	if prdNew {
		return createPRD(projectRoot)
	}

	// --edit flag: open in editor
	if prdEdit {
		return editPRD(projectRoot)
	}

	// With args: add a story
	if len(args) > 0 {
		return addStory(projectRoot, args[0])
	}

	// No args: show status
	return showPRD(projectRoot)
}

func showPRD(projectRoot string) error {
	p, err := prd.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	if p == nil {
		printWarn("No PRD found. Create one with 'ralph prd --new'")
		return nil
	}

	// Print PRD
	fmt.Printf("\033[1m\033[36mPRD: %s\033[0m\n", p.Name)
	if p.Description != "" {
		fmt.Printf("\033[2m%s\033[0m\n", p.Description)
	}
	fmt.Println()

	for _, story := range p.UserStories {
		status := " "
		if story.Passes {
			status = "✓"
		}
		fmt.Printf("[%s] %s. %s\n", status, story.ID, story.Title)
	}

	fmt.Println()
	fmt.Printf("Progress: %s (%d%%)\n", p.Progress(), p.ProgressPercent())

	return nil
}

func createPRD(projectRoot string) error {
	// Check if PRD exists
	existing, _ := prd.Load(projectRoot)
	if existing != nil {
		printWarn("PRD already exists")
		fmt.Print("Overwrite? (y/N) ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(response)) != "y" {
			return nil
		}
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\033[36mCreating new PRD...\033[0m")
	fmt.Println()

	fmt.Print("Project name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	p := &prd.PRD{
		Name:        name,
		Description: description,
		UserStories: []prd.Story{},
	}

	if err := prd.Save(projectRoot, p); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	printSuccess(fmt.Sprintf("PRD created at %s", prd.PRDPath(projectRoot)))
	printInfo("Add stories with 'ralph prd \"Story title\"'")

	return nil
}

func addStory(projectRoot string, title string) error {
	p, err := prd.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	if p == nil {
		return fmt.Errorf("no PRD found. Create one with 'ralph prd --new'")
	}

	story := prd.Story{
		Title:              title,
		AcceptanceCriteria: storyCriteria,
		Passes:             false,
	}

	p.AddStory(story)

	if err := prd.Save(projectRoot, p); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	printSuccess(fmt.Sprintf("Added story %s: %s", p.UserStories[len(p.UserStories)-1].ID, title))

	return nil
}

func editPRD(projectRoot string) error {
	prdPath := prd.PRDPath(projectRoot)

	// Check if file exists
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return fmt.Errorf("no PRD found. Create one with 'ralph prd --new'")
	}

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Open editor
	editorCmd := exec.Command(editor, prdPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}
