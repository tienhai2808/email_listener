package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	s, err := newServer(cfg)
	if err != nil {
		log.Fatalf("chạy server thất bại: %v", err)
	}

	ch := make(chan error, 1)
	go func() {
		if err := s.start(); err != nil {
			ch <- err
		}
	}()

	log.Println("Chạy server thành công")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-ch:
		log.Printf("Chạy server thất bại: %v", err)
	case <-ctx.Done():
		log.Println("Có tín hiệu dừng server")
	}

	sCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.shutdown(sCtx)
}
