package main

import (
	"L0-wb/internal/httpapi"
	"L0-wb/internal/kafkaconsumer"
	"L0-wb/internal/repository/postgres"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	fmt.Println("start")

	// Подключение к базе данных
	dsn := os.Getenv("DATABASE_DSN")
	pool := postgres.Connect(dsn)
	err := pool.Ping(context.Background())
	if err != nil {
		fmt.Printf("err conection db: %s", err)
		return
	} else {
		fmt.Println("Sucses connetion db")
	}

	// Создание репозитория
	repo := postgres.NewRepository(pool)

	// Создание консьюмера
	consumer := kafkaconsumer.NewConsumer(
		[]string{os.Getenv("KAFKA_BROKER")},
		os.Getenv("KAFKA_TOPIC"),
		os.Getenv("KAFKA_GROUP_ID"),
		repo)

	// Запуск консьюмера
	go func() {
		log.Println("Kafka consumer started")
		if err := consumer.Run(ctx); err != nil {
			log.Fatalf("Kafka consumer stopped: %v", err)
			cancel()
		}
	}()

	service := httpapi.NewService(repo)
	handler := httpapi.NewHandler(service)

	port := os.Getenv("PORT")
	// Запуск http сервера
	go func() {
		log.Printf("HTTP server started http://localhost:%s", port)

		mux := http.NewServeMux()
		mux.HandleFunc("/order/", handler.GetOrder) //GET

		if err := http.ListenAndServe(":"+port, mux); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server stopped: %v", err)
			cancel()
		}
	}()

	// Ожидание сигнала завершения
	<-ctx.Done()
	log.Println("Shutting down...")

	time.Sleep(time.Second)
}
