package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type server struct {
	http *http.Server
}

func newServer(cfg *config) (*server, error) {
	ctx := context.Background()
	credsJSON := []byte(cfg.Google.CredentialsJSON)

	psClient, err := pubsub.NewClient(ctx, cfg.Google.ProjectID, option.WithCredentialsJSON(credsJSON))
	if err != nil {
		return nil, fmt.Errorf("khởi tạo Google Pubsub thất bại: %w", err)
	}

	repo := newRepository()
	svc := newService(repo, cfg, psClient)
	hdl := newHandler(svc)
	r := gin.Default()

	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		return nil, fmt.Errorf("thiết lập Proxy thất bại: %w", err)
	}

	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	r.Use(cors.New(corsConfig))

	router(r, hdl)

	addr := fmt.Sprintf(":%d", cfg.App.Port)

	go svc.listenForNotifications(ctx)

	http := &http.Server{
		Addr:           addr,
		Handler:        r,
		MaxHeaderBytes: 5 * 1024 * 1024,
	}

	return &server{
		http,
	}, nil
}

func (s *server) start() error {
	return s.http.ListenAndServe()
}

func (s *server) shutdown(ctx context.Context) {
	if s.http != nil {
		if err := s.http.Shutdown(ctx); err != nil {
			log.Printf("Dừng http server thất bại: %v", err)
			return
		}
	}

	log.Println("Dừng server thành công")
}
