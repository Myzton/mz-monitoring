package usecase

import (
	"context"
	"errors"
	"log/slog"
	"mz-monitoring/internal/domain"
	"mz-monitoring/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo  domain.UserRepository
	secretKey []byte
}

func NewUserUsecase(r domain.UserRepository, secret []byte) *UserUsecase {
	return &UserUsecase{r, secret}
}

func (u *UserUsecase) Create(ctx context.Context, name, password, email string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "err", err)
		return nil, err
	}

	user := &domain.User{
		Name:     name,
		Password: string(hashedPassword),
		Email:    email,
	}

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserUsecase) Login(ctx context.Context, email string, password string) (string, error) {
	us, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if us == nil {
		return "", errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(us.Password), []byte(password))
	if err != nil {
		return "", errors.New("wrong password")
	}

	token, err := jwt.GenerateToken(us.ID, u.secretKey)
	if err != nil {
		slog.Error("an error occurred during token generation", "err", err)
		return "", err
	}
	return token, nil
}
