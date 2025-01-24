package otelfox

import "fmt"

const version = "v0.20.0"

var semver = fmt.Sprintf("semver:%s", version)

func Version() string {
	return version
}

func SemVersion() string {
	return semver
}
