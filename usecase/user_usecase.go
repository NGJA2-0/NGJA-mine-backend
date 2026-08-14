package usecase

import (
	"context"
	"errors"
	"time"

	"my-fiber-app/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo  domain.UserRepository
	jwtSecret string
}

// NewUserUsecase creates a new user usecase
func NewUserUsecase(ur domain.UserRepository, secret string) domain.UserUsecase {
	return &userUsecase{
		userRepo:  ur,
		jwtSecret: secret,
	}
}

func (u *userUsecase) Signup(ctx context.Context, req *domain.UserSignupRequest) (*domain.User, string, error) {
	// Check if user already exists
	existingUser, _ := u.userRepo.GetUserByNIC(ctx, req.NIC)
	if existingUser != nil {
		return nil, "", errors.New("user with this NIC already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &domain.User{
		Name:     req.Name,
		NIC:      req.NIC,
		Password: string(hashedPassword),
	}

	// Create user in DB
	err = u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := u.generateToken(user.ID.Hex())
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (u *userUsecase) Login(ctx context.Context, req *domain.UserLoginRequest) (*domain.User, string, error) {
	// Fetch user by NIC
	user, err := u.userRepo.GetUserByNIC(ctx, req.NIC)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := u.generateToken(user.ID.Hex())
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (u *userUsecase) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.jwtSecret))
}

// GetUserByID retrieves a user by their MongoDB ObjectID string
func (u *userUsecase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return u.userRepo.GetUserByID(ctx, id)
}
