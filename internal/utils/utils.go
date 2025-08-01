package utils

import (
	"fmt"
	"regexp"
)

const urlPattern = "^/update/([^/]+)/([^/]+)/([^/]+$)"

var matcher = regexp.MustCompile(urlPattern)

func ParseUrl(url string) []string {
	return matcher.FindStringSubmatch(url)
}

func ValueOrNil[T any](ptr *T, nilValue string) string {
	if ptr == nil {
		return nilValue
	}
	return fmt.Sprintf("%v", *ptr)
}
