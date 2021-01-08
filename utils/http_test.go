package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDownload(t *testing.T) {

	url := "https://raw.githubusercontent.com/reliablyhq/cli/main/README.md"

	//content := []byte("temporary file's content")
	tmpfile, err := ioutil.TempFile("", "readme")
	if err != nil {
		t.Error("Couldn't create temporary file")
	}

	defer os.RemoveAll(tmpfile.Name()) // clean up once we exit the function

	err = DownloadFile(tmpfile.Name(), url)
	if err != nil {
		t.Error("Couldn't download Pod policy")
	}

	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Error("Couldn't read from downloaded file")
	}

	t.Logf("File content:\n%s", content)

}
