package service

import (
	"encoding/json"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/logger"
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

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	logger.Log.Info(fmt.Sprintf(string(data)))
	if len(data) == 0 {
		return nil
	}

	var metrics []models.Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}
	for i := range metrics {
		err := st.UpdateMetric(metrics[i])
		if err != nil {
			return err
		}
	}
	return nil
}
