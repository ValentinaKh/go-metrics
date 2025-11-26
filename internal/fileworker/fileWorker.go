// Package fileworker предназначен для записи данных в файл в формате JSON
package fileworker

import (
	"encoding/json"
	"fmt"
	"os"
)

// Writer описывает интерфейс для записи данных
type Writer interface {
	Write(v any) error
	Close() error
}

// FileWriter  используется для записи данных в файл  в формате JSON
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

// Write записывает данные в файл в формате JSON
func (f *FileWriter) Write(v any) error {
	if err := f.encoder.Encode(v); err != nil {
		return err
	}
	return nil
}

// Close закрывает файл
func (f *FileWriter) Close() error {
	return f.file.Close()
}
