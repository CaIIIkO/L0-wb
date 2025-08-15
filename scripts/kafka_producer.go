package main

import (
	"L0-wb/internal/domain"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	mrand "math/rand"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()
	broker := []string{os.Getenv("KAFKA_BROKER")}
	topic := os.Getenv("KAFKA_TOPIC")

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: broker,
		Topic:   topic,
	})
	defer writer.Close()

	for i := 1; i <= 5; i++ {
		order := generateOrder()

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order: %v", err)
			continue
		}

		err = writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(order.OrderUID),
				Value: data,
			},
		)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
			continue
		}

		log.Printf("Sent order %s", order.OrderUID)
		time.Sleep(1 * time.Second)
	}
}

func generateOrderUID() string {
	b := make([]byte, 8)
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}
	hexPart := fmt.Sprintf("%x", b)
	return hexPart + "test"
}

func generateOrder() domain.Order {
	now := time.Now()
	orderUID := generateOrderUID()
	return domain.Order{
		OrderUID:    orderUID,
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: domain.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: domain.Payment{
			Transaction:  orderUID,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtID:      mrand.Int() % 1000000,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       now,
		OofShard:          "1",
	}
}
