package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// LoadJSON opens a file dialog to select a JSON file and returns its content.
func (a *App) LoadJSON() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select JSON File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if selection == "" {
		return "", fmt.Errorf("no file selected")
	}

	data, err := os.ReadFile(selection)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// SaveChanges opens a save file dialog and writes the provided data to the selected file.
func (a *App) SaveChanges(jsonData string) error {
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Changes",
		DefaultFilename: "filesense_changes.json",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	if err != nil {
		return err
	}
	if savePath == "" {
		return fmt.Errorf("save cancelled")
	}

	return os.WriteFile(savePath, []byte(jsonData), 0644)
}
