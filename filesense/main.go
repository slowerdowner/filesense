package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cozy/goexif2/exif"
	"github.com/cozy/goexif2/tiff"
	"github.com/denormal/go-gitignore"
)

// exifWalker is a helper struct to walk the EXIF data.
type exifWalker struct {
	exifInfo map[string]string
}

// Walk implements the exif.Walker interface.
func (w exifWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	w.exifInfo[string(name)] = tag.String()
	return nil
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

func main() {
	rootDir := flag.String("dir", ".", "The directory to scan.")
	outputFile := flag.String("out", "filesense_output.json", "The name of the output file.")
	flag.Parse()

	ignoreFile := filepath.Join(*rootDir, ".filesenseignore")
	var ignorer gitignore.GitIgnore
	// Only try to load the ignore file if it exists
	if _, err := os.Stat(ignoreFile); err == nil {
		var parseErr error
		ignorer, parseErr = gitignore.NewFromFile(ignoreFile)
		if parseErr != nil {
			fmt.Printf("error parsing ignore file: %v\n", parseErr)
			os.Exit(1)
		}
	}

	var files []FileInfo

	err := filepath.Walk(*rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Don't process the root directory itself, and check this *before* the ignorer.
		if path == *rootDir {
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

		// Extract EXIF data for JPEG files
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

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", rootDir, err)
	}

	// Write the JSON output
	jsonData, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		fmt.Printf("error marshalling json: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(*outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("error writing json file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully wrote metadata to %s\n", *outputFile)
}
