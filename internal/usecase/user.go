package usecase

import (
	"context"
	"errors"
	"log/slog"
	"mz-monitoring/internal/domain"
	"mz-monitoring/pkg/jwt"
)

type UserUsecase struct {
	userRepo  domain.UserRepository
	secretKey []byte
}

func NewUserUsecase(r domain.UserRepository, secret []byte) *UserUsecase {
	return &UserUsecase{r, secret}
}

func (u *UserUsecase) Create(ctx context.Context, name, password, email string) (*domain.User, error) {

	user := &domain.User{Name: name, Password: password, Email: email}

	err := u.userRepo.Create(ctx, user)
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
		return "", errors.New("not found")
	}

	if us.Password != password {
		return "", errors.New("wrong password")
	}

	token, err := jwt.GenerateToken(us.ID, u.secretKey)
	if err != nil {
		slog.Error("an error occurred", "err", err)
		return "", err
	}
	return token, nil

}
