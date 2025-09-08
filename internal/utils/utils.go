package utils

import (
	"fmt"
	"os"
	"regexp"
)

const urlPattern = "^/update/([^/]+)/([^/]+)/([^/]+$)"

var matcher = regexp.MustCompile(urlPattern)

func ParseURL(url string) []string {
	return matcher.FindStringSubmatch(url)
}

func LoadEnvVar[T any](key string, flagValue T, setter func(string) (T, error)) T {
	if val, ok := os.LookupEnv(key); ok {
		parsed, err := setter(val)
		if err != nil {
			panic(fmt.Sprintf("не удалось распарсить %s=%q: %v", key, val, err))
		}
		return parsed
	}
	return flagValue
}

func ToString[T any](ptr *T) string {
	if ptr == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *ptr)
}
