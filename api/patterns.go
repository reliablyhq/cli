package api

import (
	//"errors"
	"fmt"
	//"net/http"
	"os"
)

// DownloadPatternsBundle downloads the patterns as a bundle from API
// and writes it to a local fila at given path
func DownloadPatternsBundle(client *Client, hostname string, filepath string) error {

	newFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("Can't create file at path '%s'", filepath)
	}
	defer newFile.Close()

	err = client.DownloadFile(hostname, "GET", "patterns", nil, newFile)
	if err != nil {
		return err
	}

	/*
		detectedFileType := http.DetectContentType(content)
		switch detectedFileType {
			case "application/gzip":
					break
			default:
				return fmt.Errorf("Invalid file type received '%s'", detectedFileType)
		}
	*/

	/*
		if _, err := newFile.Write(content); err != nil {
			return fmt.Errorf("Can't write to file at path '%s'", filepath)
		}
	*/

	return nil
}
