package utils

import (
	"crypto/hmac"
	"crypto/sha256"
)

func Hash(key string, src []byte) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(src)
	return h.Sum(nil)
}
