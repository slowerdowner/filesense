package main

import (
	"context"
	"archive/zip"
	"context"
	"fmt"
	"io"
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

// SelectZipFile opens a save file dialog and returns the selected path.
func (a *App) SelectZipFile() (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Zip File",
		DefaultFilename: "archive.zip",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Zip Files (*.zip)",
				Pattern:     "*.zip",
			},
		},
	})
}

// ZipFiles creates a zip archive from a list of file paths.
func (a *App) ZipFiles(filePaths []string, dest string) error {
	zipFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, filePath := range filePaths {
		fileToZip, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		info, err := fileToZip.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filePath // Use the full path for the name in the archive.
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			return err
		}
	}
	return nil
}
