package detect

import (
	"regexp"
	"strings"
)

var (
	// semantic version
	// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	reSemVer = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

	// calendar version
	reCalVer = regexp.MustCompile(`^\d{4,8}\.\d+(-[0-9a-zA-Z-]+)?$`)

	// sequential version
	reSeqVer = regexp.MustCompile(`^\d+(\.\d+)?(\.\d+)?(\.\d+)?(-[0-9a-zA-Z-]+)?$`)
)

func IsTag(tag string) bool {
	tag = strings.TrimPrefix(tag, "v")

	if tag == "" {
		return false
	}

	if reSemVer.MatchString(tag) {
		return true
	}

	if reCalVer.MatchString(tag) {
		return true
	}

	if reSeqVer.MatchString(tag) {
		return true
	}

	return false
}
