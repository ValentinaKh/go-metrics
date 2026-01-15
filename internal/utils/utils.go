package utils

import (
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"go.uber.org/zap"
	"net"
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

func GetLocalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Log.Error("Error closing connection", zap.Error(err))
		}
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
