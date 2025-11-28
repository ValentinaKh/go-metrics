package service

import (
	"context"
	"encoding/json"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"io"
	"os"

	models "github.com/ValentinaKh/go-metrics/internal/model"
)

// LoadMetrics загружает метрики из файла
func LoadMetrics(fileName string, st Storage) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Log.Error(err.Error())
		}
	}(file)

	decoder := json.NewDecoder(file)

	var lastResult []models.Metrics
	for {
		if err := decoder.Decode(&lastResult); err != nil {
			if err == io.EOF { // Достигнут конец файла
				break
			}
			return err
		}
	}
	for i := range lastResult {
		err := st.UpdateMetric(context.TODO(), lastResult[i])
		if err != nil {
			return err
		}
	}
	return nil
}
