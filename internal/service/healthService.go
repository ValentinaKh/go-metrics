package service

import "context"

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	rep HealthRepository
}

func NewHealthService(rep HealthRepository) *HealthService {
	return &HealthService{rep: rep}
}

func (s *HealthService) CheckDB(ctx context.Context) error {
	return s.rep.Ping(ctx)
}
