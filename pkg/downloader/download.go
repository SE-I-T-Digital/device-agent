package downloader

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

// ExtractFilename extracts the filename from a URL
func ExtractFilename(inputurl string) string {
	u, err := url.Parse(inputurl)
	if err != nil {
		return ""
	}
	return path.Base(u.Path)
}

// Download downloads a file from a URL and saves it to a local file
func Download(url string, filepath string) error {
	// Create the file
	filename := ExtractFilename(url)
	out, err := os.Create(filepath + "/" + filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
