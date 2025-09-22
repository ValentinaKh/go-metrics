package provider

import (
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemProvider struct {
}

func NewSystemProvider() *SystemProvider {
	return &SystemProvider{}
}

func (p *SystemProvider) Collect() ([]models.Metrics, error) {
	var result []models.Metrics

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	value := float64(vmStat.Total)
	result = append(result, models.Metrics{
		ID:    models.TotalMemory,
		MType: models.Gauge,
		Value: &value,
	})

	value = float64(vmStat.Free)
	result = append(result, models.Metrics{
		ID:    models.FreeMemory,
		MType: models.Gauge,
		Value: &value,
	})

	percents, err := cpu.Percent(0, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU stats: %w", err)
	}

	for i, pct := range percents {
		v := pct
		result = append(result, models.Metrics{
			ID:    fmt.Sprintf("CPUutilization%d", i+1),
			MType: models.Gauge,
			Value: &v,
		})
	}

	return result, nil
}
