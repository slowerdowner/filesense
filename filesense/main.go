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
	"archive/zip"
	"encoding/xml"
	"github.com/cozy/goexif2/tiff"
	"github.com/denormal/go-gitignore"
	pdf "github.com/unidoc/unipdf/v3/model"
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
	Pdf     map[string]string `json:"pdf,omitempty"`
	Ooxml   map[string]string `json:"ooxml,omitempty"`
}

func main() {
	rootDir := flag.String("dir", ".", "The directory to scan.")
	outputFile := flag.String("out", "filesense_output.json", "The name of the output file.")
	ignoreFile := flag.String("ignore-file", ".filesenseignore", "The path to the ignore file.")
	flag.Parse()

	var ignorer gitignore.GitIgnore
	// Only try to load the ignore file if it exists
	if _, err := os.Stat(*ignoreFile); err == nil {
		var parseErr error
		ignorer, parseErr = gitignore.NewFromFile(*ignoreFile)
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

		// Extract metadata from PDF files
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".pdf") {
			file, err := os.Open(path)
			if err != nil {
				log.Printf("failed to open file %s: %v", path, err)
			} else {
				defer file.Close()
				pdfReader, err := pdf.NewPdfReader(file)
				if err != nil {
					log.Printf("failed to create pdf reader for %s: %v", path, err)
				} else {
					info, err := pdfReader.GetPdfInfo()
					if err != nil {
						log.Printf("failed to get pdf info from %s: %v", path, err)
					}
					if info != nil {
						fileInfo.Pdf = make(map[string]string)
						if info.Author != nil {
								fileInfo.Pdf["Author"] = DecodePdfText(info.Author.String())
						}
						if info.Title != nil {
								fileInfo.Pdf["Title"] = DecodePdfText(info.Title.String())
						}
						if info.Subject != nil {
								fileInfo.Pdf["Subject"] = DecodePdfText(info.Subject.String())
						}
						if info.Creator != nil {
								fileInfo.Pdf["Creator"] = DecodePdfText(info.Creator.String())
						}
						if info.Producer != nil {
								fileInfo.Pdf["Producer"] = DecodePdfText(info.Producer.String())
						}
					}
				}
			}
		}

		files = append(files, fileInfo)

		// Extract metadata from OOXML files
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".docx") || strings.HasSuffix(strings.ToLower(path), ".xlsx") || strings.HasSuffix(strings.ToLower(path), ".pptx")) {
			r, err := zip.OpenReader(path)
			if err != nil {
				log.Printf("failed to open ooxml file %s: %v", path, err)
			} else {
				defer r.Close()
				for _, f := range r.File {
					if f.Name == "docProps/core.xml" {
						rc, err := f.Open()
						if err != nil {
							log.Printf("failed to open core.xml for %s: %v", path, err)
							break
						}
						defer rc.Close()

						decoder := xml.NewDecoder(rc)
						fileInfo.Ooxml = make(map[string]string)
						for {
							token, err := decoder.Token()
							if err != nil {
								break
							}
							switch se := token.(type) {
							case xml.StartElement:
								var s string
								if err := decoder.DecodeElement(&s, &se); err == nil {
									switch se.Name.Local {
									case "title":
										fileInfo.Ooxml["Title"] = s
									case "creator":
										fileInfo.Ooxml["Creator"] = s
									case "subject":
										fileInfo.Ooxml["Subject"] = s
									case "modified":
										fileInfo.Ooxml["Modified"] = s
									}
								}
							}
						}
						break
					}
				}
			}
		}

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

func DecodePdfText(s string) string {
	if len(s) > 2 && s[0] == 254 && s[1] == 255 {
		s = s[2:]
		out := make([]rune, len(s)/2)
		for i := 0; i < len(s); i += 2 {
			out[i/2] = rune(s[i])<<8 | rune(s[i+1])
		}
		return string(out)
	}
	return s
}
