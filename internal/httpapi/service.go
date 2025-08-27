package httpapi

import (
	"L0-wb/internal/domain"
	"L0-wb/internal/repository/cache"
	"L0-wb/internal/repository/postgres"
	"context"
)

type Service struct {
	repo  *postgres.Repository
	cache *cache.Cache
}

func NewService(repo *postgres.Repository, cache *cache.Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) GetOrder(ctx context.Context, orderUid string) (*domain.Order, error) {

	order, ok := s.cache.Get(orderUid)
	if ok {
		return &order, nil
	}

	order, err := s.repo.GetOrder(ctx, orderUid)
	if err != nil {
		return nil, err
	}
	s.cache.Set(order)

	return &order, nil
}
