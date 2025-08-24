
package services

import (
	"context"
	"fmt"
	"github.com/sumanthd032/go-shorty/internal/repositories/db"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	queries *db.Queries
}

func NewUserService(queries *db.Queries) *UserService {
	return &UserService{queries: queries}
}

func (s *UserService) Register(ctx context.Context, email, password string) (db.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, fmt.Errorf("could not hash password: %w", err)
	}

	params := db.CreateUserParams{
		Email:        email,
		PasswordHash: hashedPassword,
	}

	user, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		// You'd check for specific DB errors here, e.g., duplicate email
		return db.User{}, fmt.Errorf("could not create user: %w", err)
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (db.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return db.User{}, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		return db.User{}, fmt.Errorf("invalid credentials")
	}
	return user, nil
}