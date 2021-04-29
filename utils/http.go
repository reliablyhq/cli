package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	client := &http.Client{
		Timeout: time.Second * 2,
	}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("No file found at URL: %v", url)
	}

	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// DownloadURL downloads from a given URL and returns the response body
func DownloadURL(url string) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Get(url)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("No file found at URL: %v", url)
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, errors.New("Unable to read response body")
	}

	return bs, nil
}
