package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cozy/goexif2/exif"
	"github.com/cozy/goexif2/tiff"
	"github.com/denormal/go-gitignore"
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

// FileInfo holds metadata for a single file.
type FileInfo struct {
	Name    string            `json:"name"`
	Path    string            `json:"path"`
	Size    int64             `json:"size"`
	ModTime time.Time         `json:"mod_time"`
	IsDir   bool              `json:"is_dir"`
	Exif    map[string]string `json:"exif,omitempty"`
}

// exifWalker is a helper struct to walk the EXIF data.
type exifWalker struct {
	exifInfo map[string]string
}

// Walk implements the exif.Walker interface.
func (w exifWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	w.exifInfo[string(name)] = tag.String()
	return nil
}

// ScanDirectory opens a directory dialog and scans the selected directory.
func (a *App) ScanDirectory() ([]FileInfo, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory to Scan",
	})
	if err != nil {
		return nil, err
	}
	if selection == "" {
		return nil, fmt.Errorf("no directory selected")
	}

	return a.scan(selection)
}

func (a *App) scan(rootDir string) ([]FileInfo, error) {
	ignoreFile := filepath.Join(rootDir, ".filesenseignore")
	var ignorer gitignore.GitIgnore
	if _, err := os.Stat(ignoreFile); err == nil {
		var parseErr error
		ignorer, parseErr = gitignore.NewFromFile(ignoreFile)
		if parseErr != nil {
			log.Printf("error parsing ignore file: %v\n", parseErr)
		}
	}

	var files []FileInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == rootDir {
			return nil
		}

		if ignorer != nil {
			match := ignorer.Match(path)
			if match != nil && match.Ignore() {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		fileInfo := FileInfo{
			Name:    info.Name(),
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}

		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".jpg") || strings.HasSuffix(strings.ToLower(path), ".jpeg")) {
			file, err := os.Open(path)
			if err != nil {
				log.Printf("failed to open file %s: %v", path, err)
			} else {
				defer file.Close()
				x, err := exif.Decode(file)
				if err != nil {
					log.Printf("failed to decode exif from %s: %v", path, err)
				} else {
					fileInfo.Exif = make(map[string]string)
					walker := exifWalker{exifInfo: fileInfo.Exif}
					err = x.Walk(walker)
					if err != nil {
						log.Printf("error walking exif data for %s: %v", path, err)
					}
				}
			}
		}

		files = append(files, fileInfo)
		return nil
	})

	return files, err
}

// Change represents a single file modification.
type Change struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
	NewName string    `json:"new_name"`
	NewPath string    `json:"new_path"`
}

// ApplyChanges applies the given list of changes.
func (a *App) ApplyChanges(changesJSON string) error {
	var changes []Change
	if err := json.Unmarshal([]byte(changesJSON), &changes); err != nil {
		return fmt.Errorf("error unmarshalling changes: %v", err)
	}

	// Sort changes by path length descending to handle nested renames correctly
	// (children before parents)
	sort.Slice(changes, func(i, j int) bool {
		return len(changes[i].Path) > len(changes[j].Path)
	})

	for _, change := range changes {
        // Basic conflict check
		if _, err := os.Stat(change.NewPath); !os.IsNotExist(err) {
            // For now, fail on conflict.
            // TODO: Better conflict resolution
			return fmt.Errorf("destination already exists: %s", change.NewPath)
		}

		if err := os.Rename(change.Path, change.NewPath); err != nil {
			return fmt.Errorf("failed to rename %s to %s: %v", change.Path, change.NewPath, err)
		}
	}

	return nil
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
