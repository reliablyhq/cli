package utils

import (
	"github.com/pkg/browser"
)

// OpenInBrowser opens the url in a web browser based on OS
func OpenInBrowser(url string) error {
	return browser.OpenURL(url)
}
