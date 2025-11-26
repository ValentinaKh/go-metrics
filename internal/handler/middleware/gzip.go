// Package middleware содержит middleware для rest вызовов
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

type compressWriter struct {
	w    http.ResponseWriter
	zw   *gzip.Writer
	init bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{w: w}
}

func (c *compressWriter) initWriter() {
	if c.init {
		return
	}
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(c.w)
	c.zw = zw
	c.init = true
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.init {
		c.initWriter()
	}
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusOK {
		c.initWriter()
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if !c.init {
		return nil
	}
	err := c.zw.Close()
	c.zw.Reset(io.Discard)
	gzipWriterPool.Put(c.zw)
	c.zw = nil
	c.init = false
	return err
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
