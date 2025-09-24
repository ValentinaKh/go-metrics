package middleware

import (
	"net/http"
)

type hashWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func NewHashWriter(writer http.ResponseWriter) *hashWriter {
	return &hashWriter{
		ResponseWriter: writer,
		statusCode:     200,
		body:           make([]byte, 0),
	}
}

func (hrw *hashWriter) Write(b []byte) (int, error) {
	hrw.body = append(hrw.body, b...)
	return len(b), nil
}

func (hrw *hashWriter) WriteHeader(statusCode int) {
	hrw.statusCode = statusCode
}
