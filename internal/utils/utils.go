package utils

import (
	"regexp"
)

const urlPattern = "^/update/([^/]+)/([^/]+)/([^/]+$)"

var matcher = regexp.MustCompile(urlPattern)

func ParseURL(url string) []string {
	return matcher.FindStringSubmatch(url)
}
