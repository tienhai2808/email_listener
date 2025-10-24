package main

import (
	"context"

	"golang.org/x/oauth2"
)

type service interface {
	login(ctx context.Context, code string) (*user, error)

	listenForNotifications(ctx context.Context)
}

type repository interface {
	findOrCreate(ctx context.Context, email, name, picture string) (*user, error)

	saveToken(ctx context.Context, email string, token *oauth2.Token) error

	getToken(ctx context.Context, email string) (*oauth2.Token, error)

	saveHistoryID(ctx context.Context, email string, historyID uint64) error

	getHistoryID(ctx context.Context, email string) (uint64, error)
}

type apiResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type loginRequest struct {
	Code string `json:"code" binding:"required"`
}

type user struct {
	ID        string        `json:"id"`
	Email     string        `json:"email"`
	Name      string        `json:"name"`
	Picture   string        `json:"picture"`
	Token     *oauth2.Token `json:"-"`
	HistoryID uint64        `json:"-"`
}
