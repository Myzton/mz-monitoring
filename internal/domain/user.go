package domain

import (
	"context"
)

type User struct {
	ID       int
	Name     string
	Password string
	Email    string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}
