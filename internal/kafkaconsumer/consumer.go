package kafkaconsumer

import (
	"L0-wb/internal/domain"
	"L0-wb/internal/repository/postgres"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	repo   *postgres.Repository
}

func NewConsumer(brokers []string, topic, groupID string, repo *postgres.Repository) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	return &Consumer{
		reader: reader,
		repo:   repo,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.reader.Close()

	backoff := time.Second
	const maxBackoff = 10 * time.Second

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			// временные сетевые/координаторные проблемы — ретрай
			log.Printf("kafka fetch error: %v; retrying in %s...", err, backoff)
			select {
			case <-time.After(backoff):
				if backoff < maxBackoff {
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
				}
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// сбрасываем backoff на успехе
		backoff = time.Second

		var order domain.Order
		if err := json.Unmarshal(m.Value, &order); err != nil || order.OrderUID == "" {
			log.Printf("skip invalid message: %v uid=%q", err, order.OrderUID)
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if err := c.repo.InsertOrder(ctx, &order); err != nil {
			//не делаем коммит чтобы перечитать позже
			log.Printf("db error: %v; will retry later (not committing offset)", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("kafka commit error: %v; ignoring", err)
		}
		log.Printf("sucsess insert order :%s", order.OrderUID)
	}
}
