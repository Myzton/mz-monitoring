package usecase

import (
	"context"
	"errors"
	"mz-monitoring/internal/domain"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(r domain.UserRepository) *userUsecase {
	return &userUsecase{r}
}

func (u *userUsecase) Create(ctx context.Context, name, password, email string) (*domain.User, error) {
	existingUser, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}
	user := &domain.User{Name: name, Password: password, Email: email}

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
