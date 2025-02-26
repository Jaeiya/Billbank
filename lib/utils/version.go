package utils

import "fmt"

var (
	appVersion string
	codeName   string
	goVersion  string
	commitSha  string
)

/*
GetVersion returns the application version set by the build process.
If the binary is a dev build, then it returns a truncated version
of the latest commit hash.
*/
func GetVersion() string {
	if appVersion == "" {
		if commitSha == "" {
			return "invalid version"
		}
		return fmt.Sprintf("%s-%s", codeName, commitSha[:8])
	}
	return appVersion
}

func GetGoVersion() string {
	return goVersion
}
