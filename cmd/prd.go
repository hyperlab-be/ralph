package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hyperlab/ralph/internal/config"
	"github.com/hyperlab/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var prdCmd = &cobra.Command{
	Use:     "prd [command]",
	Aliases: []string{"p"},
	Short:   "Manage PRD (Product Requirement Document)",
	Long:    `View, create, or edit the PRD for the current project.`,
	RunE:    runPrdShow,
}

var prdCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new PRD",
	RunE:  runPrdCreate,
}

var prdAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a story to the PRD",
	RunE:  runPrdAdd,
}

var prdEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit PRD in your editor",
	RunE:  runPrdEdit,
}

func init() {
	prdCmd.AddCommand(prdCreateCmd)
	prdCmd.AddCommand(prdAddCmd)
	prdCmd.AddCommand(prdEditCmd)
	rootCmd.AddCommand(prdCmd)
}

func runPrdShow(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a rl project")
	}

	p, err := prd.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	if p == nil {
		printWarn("No PRD found. Create one with 'rl prd create'")
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

func runPrdCreate(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a rl project. Run 'rl init' first")
	}

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
	printInfo("Add stories with 'rl prd add'")

	return nil
}

func runPrdAdd(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a rl project")
	}

	p, err := prd.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	if p == nil {
		return fmt.Errorf("no PRD found. Create one with 'rl prd create'")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\033[36mAdding new story...\033[0m")
	fmt.Println()

	fmt.Print("Title: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)

	fmt.Print("Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Println("Acceptance criteria (one per line, empty line to finish):")
	var criteria []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		criteria = append(criteria, line)
	}

	story := prd.Story{
		Title:              title,
		Description:        description,
		AcceptanceCriteria: criteria,
		Passes:             false,
	}

	p.AddStory(story)

	if err := prd.Save(projectRoot, p); err != nil {
		return fmt.Errorf("failed to save PRD: %w", err)
	}

	printSuccess(fmt.Sprintf("Story %d added", len(p.UserStories)))

	return nil
}

func runPrdEdit(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a rl project")
	}

	prdPath := prd.PRDPath(projectRoot)

	// Check if file exists
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return fmt.Errorf("no PRD found. Create one with 'rl prd create'")
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
