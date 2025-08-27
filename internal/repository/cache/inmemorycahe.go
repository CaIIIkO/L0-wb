package cache

import (
	"L0-wb/internal/domain"
	"sync"
)

type Cache struct {
	mu sync.RWMutex
	m  map[string]domain.Order
}

func New() *Cache {
	return &Cache{
		m: make(map[string]domain.Order),
	}
}

func (c *Cache) Set(order domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[order.OrderUID] = order
}

func (c *Cache) Get(id string) (domain.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.m[id]
	return order, ok
}

func (c *Cache) Warm(orders []domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.m[order.OrderUID] = order
	}
}
