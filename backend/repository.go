package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"golang.org/x/oauth2"
)

type repositoryImpl struct {
	users map[string]*user
	mu    sync.RWMutex
}

func newRepository() repository {
	return &repositoryImpl{
		users: make(map[string]*user),
	}
}

func (r *repositoryImpl) findOrCreate(ctx context.Context, email, name, picture string) (*user, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, exists := r.users[email]; exists {
		user.Name = name
		user.Picture = picture
		return user, nil
	}

	newUser := &user{
		ID:      fmt.Sprintf("user_%d", len(r.users)+1),
		Email:   email,
		Name:    name,
		Picture: picture,
	}
	r.users[email] = newUser

	log.Printf("Đã tạo người dùng mới: %s", email)
	return newUser, nil
}

func (r *repositoryImpl) saveToken(ctx context.Context, email string, token *oauth2.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, exists := r.users[email]; exists {
		user.Token = token
		log.Printf("Đã cập nhật token cho: %s", email)
		return nil
	}
	return fmt.Errorf("không tìm thấy user với email: %s", email)
}

func (r *repositoryImpl) getToken(ctx context.Context, email string) (*oauth2.Token, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if user, exists := r.users[email]; exists && user.Token != nil {
		return user.Token, nil
	}
	return nil, fmt.Errorf("không tìm thấy token cho email: %s", email)
}

func (r *repositoryImpl) saveHistoryID(ctx context.Context, email string, historyID uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, exists := r.users[email]; exists {
		user.HistoryID = historyID
		log.Printf("Đã cập nhật historyID cho: %s", email)
		return nil
	}
	return fmt.Errorf("không tìm thấy user với email: %s", email)
}

func (r *repositoryImpl) getHistoryID(ctx context.Context, email string) (uint64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if user, exists := r.users[email]; exists && user.HistoryID != 0 {
		return user.HistoryID, nil
	}
	return 0, fmt.Errorf("không tìm thấy historyID cho email: %s", email)
}
