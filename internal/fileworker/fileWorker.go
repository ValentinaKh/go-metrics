package fileworker

import (
	"encoding/json"
	"fmt"
	"os"
)

type Writer interface {
	Write(v any) error
	Close() error
}

type FileWriter struct {
	encoder *json.Encoder
	file    *os.File
}

func NewFileWriter(fileName string) (*FileWriter, error) {
	if fileName == "" {
		return nil, fmt.Errorf("fileName is empty")
	}
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (f *FileWriter) Write(v any) error {
	if err := f.encoder.Encode(v); err != nil {
		return err
	}
	return nil
}

func (f *FileWriter) Close() error {
	return f.file.Close()
}
