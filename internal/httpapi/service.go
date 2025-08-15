package httpapi

import (
	"L0-wb/internal/domain"
	"L0-wb/internal/repository/postgres"
	"context"
)

type Service struct {
	repo *postgres.Repository
}

func NewService(repo *postgres.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOrder(ctx context.Context, orderUid string) (*domain.Order, error) {
	order, err := s.repo.GetOrder(ctx, orderUid)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
