package otelfox

import "fmt"

const version = "v0.22.1"

var semver = fmt.Sprintf("semver:%s", version)

func Version() string {
	return version
}

func SemVersion() string {
	return semver
}
