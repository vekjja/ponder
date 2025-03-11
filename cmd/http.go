package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Create HTTP Client
var httpClient = &http.Client{
	Timeout: time.Second * 60,
}

// download file from url and save to local directory
func httpDownloadFile(url string, filePath string) string {
	// Replace spaces with underscores
	filePath = strings.ReplaceAll(filePath, " ", "_")
	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File already exists, so rename the new file
		dir := filepath.Dir(filePath)
		ext := filepath.Ext(filePath)
		name := filepath.Base(filePath[:len(filePath)-len(ext)])
		i := 1
		for {
			newName := fmt.Sprintf("%s_%d%s", name, i, ext)
			newFilepath := filepath.Join(dir, newName)
			_, err := os.Stat(newFilepath)
			if os.IsNotExist(err) {
				// New filename is available, use it
				filePath = newFilepath
				break
			}
			i++
		}
	}

	// Create the file
	out, err := os.Create(filePath)
	catchErr(err)
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	catchErr(err)
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	catchErr(err)
	return filePath
}
