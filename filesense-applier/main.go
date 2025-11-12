package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Change represents a single file modification from the JSON input.
type Change struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
	NewName string    `json:"new_name"`
	NewPath string    `json:"new_path"`
}

// Operation encapsulates a change and the logic to apply/revert it.
type Operation struct {
	Change
	moveCmd string
}

func (o *Operation) ForwardCommand() string {
	return fmt.Sprintf("%s \"%s\" \"%s\"", o.moveCmd, o.Path, o.NewPath)
}

func (o *Operation) ReverseCommand() string {
	return fmt.Sprintf("%s \"%s\" \"%s\"", o.moveCmd, o.NewPath, o.Path)
}

func (o *Operation) Execute() error {
	return os.Rename(o.Path, o.NewPath)
}

func main() {
	// Define a boolean flag for the dry-run mode.
	dryRun := flag.Bool("nerf", false, "Output the changelog but do not apply any commands.")

	// Parse the command-line flags.
	flag.Parse()

	// Check for the required JSON file path argument.
	if flag.NArg() < 1 {
		fmt.Println("Usage: filesense-applier [options] <path_to_changes.json>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	jsonPath := flag.Arg(0)

	// Read the JSON file.
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	// Unmarshal the JSON data.
	var changes []Change
	if err := json.Unmarshal(data, &changes); err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
		os.Exit(1)
	}

	// Determine the move command based on the OS.
	moveCmd := "mv"
	if runtime.GOOS == "windows" {
		moveCmd = "move"
	}

	// Create operations from changes.
	operations := make([]*Operation, len(changes))
	for i, change := range changes {
		operations[i] = &Operation{Change: change, moveCmd: moveCmd}
	}

	// Sort operations by path length, descending, to process children before parents.
	sort.Slice(operations, func(i, j int) bool {
		return len(operations[i].Path) > len(operations[j].Path)
	})

	// Print a summary of the planned operations.
	fmt.Printf("Successfully parsed %d operations from %s.\n", len(operations), jsonPath)
	if *dryRun {
		fmt.Println("Running in dry-run mode. No changes will be made.")
	}
	fmt.Println("Planned operations:")
	for _, op := range operations {
		fmt.Printf("  - Rename %s to %s\n", op.Path, op.NewPath)
	}

	// Generate the forward script before applying changes.
	if err := generateForwardScript(operations); err != nil {
		fmt.Printf("Error generating forward script: %v\n", err)
		os.Exit(1)
	}

	// Apply the changes if not in dry-run mode.
	var successfulOps []*Operation
	if !*dryRun {
		var err error
		successfulOps, err = applyChanges(operations)
		if err != nil {
			fmt.Printf("Error applying changes: %v\n", err)
			os.Exit(1)
		}
	} else {
		successfulOps = operations
	}

	// Generate the reverse script after applying changes.
	if err := generateReverseScript(successfulOps); err != nil {
		fmt.Printf("Error generating reverse script: %v\n", err)
		os.Exit(1)
	}
}

func applyChanges(operations []*Operation) ([]*Operation, error) {
	fmt.Println("\nApplying changes...")
	reader := bufio.NewReader(os.Stdin)
	var finalOps []*Operation

operationLoop:
	for _, op := range operations {
		fmt.Printf("Processing: %s -> %s\n", op.Path, op.NewPath)

		// Check for conflict
		if _, err := os.Stat(op.NewPath); !os.IsNotExist(err) {
			fmt.Printf("  - CONFLICT: Destination '%s' already exists.\n", op.NewPath)
			for {
				fmt.Print("  - Choose an action: [S]kip (default), [R]ename: ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(strings.ToUpper(input))

				if input == "" || input == "S" {
					fmt.Println("    -> Skipped.")
					continue operationLoop
				} else if input == "R" {
					fmt.Print("    -> Enter new name: ")
					newName, _ := reader.ReadString('\n')
					newName = strings.TrimSpace(newName)
					op.NewPath = filepath.Join(filepath.Dir(op.NewPath), newName)
					fmt.Printf("    -> New path: %s\n", op.NewPath)
					continue
				}
				fmt.Println("    -> Invalid option.")
			}
		}

		// Apply the rename
		for {
			err := op.Execute()
			if err == nil {
				break // Success
			}

			fmt.Printf("  - ERROR: Could not rename %s: %v\n", op.Path, err)
			fmt.Print("  - Choose an action: [R]etry, [S]kip: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToUpper(input))

			if input == "S" {
				fmt.Println("    -> Skipped.")
				continue operationLoop
			} else if input == "R" {
				fmt.Println("    -> Retrying...")
				continue
			}
			fmt.Println("    -> Invalid option.")
		}
		finalOps = append(finalOps, op)
	}

	fmt.Println("Changes applied successfully.")
	return finalOps, nil
}

func generateForwardScript(ops []*Operation) error {
	timestamp := time.Now().Format("20060102-150405")
	scriptName := fmt.Sprintf("%s_filesense-changelog.sh", timestamp)
	shebang := "#!/bin/bash"

	if runtime.GOOS == "windows" {
		scriptName = fmt.Sprintf("%s_filesense-changelog.cmd", timestamp)
		shebang = "@echo off"
	}

	var script strings.Builder
	script.WriteString(shebang + "\n")

	for _, op := range ops {
		script.WriteString(op.ForwardCommand() + "\n")
	}

	if err := os.WriteFile(scriptName, []byte(script.String()), 0755); err != nil {
		return fmt.Errorf("writing forward script: %w", err)
	}
	fmt.Printf("Generated changelog script: %s\n", scriptName)
	return nil
}

func generateReverseScript(ops []*Operation) error {
	timestamp := time.Now().Format("20060102-150405")
	scriptName := fmt.Sprintf("%s_revert.sh", timestamp)
	shebang := "#!/bin/bash"

	if runtime.GOOS == "windows" {
		scriptName = fmt.Sprintf("%s_revert.cmd", timestamp)
		shebang = "@echo off"
	}

	var script strings.Builder
	script.WriteString(shebang + "\n")

	// Iterate in reverse for the revert script
	for i := len(ops) - 1; i >= 0; i-- {
		script.WriteString(ops[i].ReverseCommand() + "\n")
	}

	if err := os.WriteFile(scriptName, []byte(script.String()), 0755); err != nil {
		return fmt.Errorf("writing reverse script: %w", err)
	}
	fmt.Printf("Generated revert script: %s\n", scriptName)
	return nil
}
