package service

import (
	"encoding/json"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"io"
	"os"
)

func LoadMetrics(fileName string, st Storage) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

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
		err := st.UpdateMetric(lastResult[i])
		if err != nil {
			return err
		}
	}
	return nil
}
