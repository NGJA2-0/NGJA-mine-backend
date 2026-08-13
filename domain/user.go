package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents the user entity
type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	NIC      string             `json:"nic" bson:"nic"`
	Password string             `json:"-" bson:"password"` // never expose password in JSON
}

// UserSignupRequest represents the payload for signup
type UserSignupRequest struct {
	Name     string `json:"name"`
	NIC      string `json:"nic"`
	Password string `json:"password"`
}

// UserLoginRequest represents the payload for login
type UserLoginRequest struct {
	NIC      string `json:"nic"`
	Password string `json:"password"`
}

// UserRepository defines the database operations for User
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByNIC(ctx context.Context, nic string) (*User, error)
}

// UserUsecase defines the business logic operations for User
type UserUsecase interface {
	Signup(ctx context.Context, req *UserSignupRequest) (*User, string, error) // returns user, token, error
	Login(ctx context.Context, req *UserLoginRequest) (*User, string, error)   // returns user, token, error
}
