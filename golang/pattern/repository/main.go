package main

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type Order struct {
	ID        string    `json:"id" :"id"`
	UserID    string    `json:"user_id" :"user_id"`
	Amount    float64   `json:"amount" :"amount"`
	Status    string    `json:"status" :"status"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderRepository interface {
	Create(Order) (string, error)
	Get(id string) (Order, error)
	Update(Order) (string, error)
}

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (s *OrderService) PlaceOrder(ctx context.Context, userID string, amount float64) (*Order, error) {
	order := &Order{
		ID:        uuid.NewString(),
		UserID:    userID,
		Amount:    amount,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	_, err := s.repo.Create(*order)
	if err != nil {
		return nil, err
	}
	return order, nil
}
