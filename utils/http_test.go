package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadFile(t *testing.T) {

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

	t.Logf("File content:\n%s\n...", content[:50])

}

func TestDownloadURL(t *testing.T) {

	url := "https://raw.githubusercontent.com/reliablyhq/cli/main/README.md"

	body, err := DownloadURL(url)
	assert.NoError(t, err, "Could not download URL")
	assert.NotEqual(t, "", body, "")

	t.Logf("%s\n...", body[:50])

	// Invalid URL
	url = "https://raw.githubusercontent.com/reliablyhq/cli/main/README.doesnotexist"
	body, err = DownloadURL(url)
	assert.Error(t, err, "Expected error not fetched")
	assert.Equal(t, []byte{}, body)

}
