package context

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// URL from CI env var

// https://gitlab.com/reliably/reliably-discovery-demo
// https://gitlab.com/reliably/reliably-discovery-demo.git
// git@gitlab.com:reliably/reliably-discovery-demo.git

//git urls

/*

GitRemote {

	server
	path
}

name:  "gitlab https url",
url:   "https://gitlab.com/reliably/reliably-discovery-demo.git",
wants: []string{"reliably", "reliably-discovery-demo"},
},
{
name:  "gitlab ssh url",
url:   "git@gitlab.com:reliably/reliably-discovery-demo.git",
wants: []string{"reliably", "reliably-discovery-demo"},

// for the 3 ways of using a Gitlab Repo, i want it to be known as the same source
//

*/

// Currently not testing runtime or source
func TestNewContext(t *testing.T) {

	envs := map[string]string{
		"GitHub":   "../../tests/github.env",
		"CircleCI": "../../tests/circleci.env",
	}

	for k, v := range envs {
		localEnv = readEnvFile(v)
		ctx := *NewContext()
		assert.True(t, ctx.IsCI, "NewContext CI undetected: "+k)
		assert.Equal(t, StringMap(localEnv), ctx.Environ, "Context.Environ does not equal local environment: "+k)
	}
}

func readEnvFile(filePath string) (envData map[string]string) {
	envFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println("File reading error", err)
	}

	envData = make(map[string]string)
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()
		if equalIndex := strings.Index(line, "="); equalIndex >= 0 {
			val := line[equalIndex+1:]
			if keyIndex := strings.Index(line, "export "); keyIndex >= 0 {
				envData[line[7:equalIndex]] = val
			} else {
				envData[line[:equalIndex]] = val
			}

		}
	}

	return
}
