package init

import "os"

func getDefaultAppName() string {
	return "my-app"
}

func getDefaultAppOwner() string {
	return "me"
}

func getDefaultRepository() string {
	cwd, _ := os.Getwd()
	return cwd
}
