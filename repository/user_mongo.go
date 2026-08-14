package repository

import (
	"context"
	"errors"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userMongoRepo struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *mongo.Database) domain.UserRepository {
	return &userMongoRepo{
		db:         db,
		collection: db.Collection("users"),
	}
}

// CreateUser inserts a new user into MongoDB
func (r *userMongoRepo) CreateUser(ctx context.Context, user *domain.User) error {
	user.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// GetUserByNIC fetches a user by their NIC
func (r *userMongoRepo) GetUserByNIC(ctx context.Context, nic string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"nic": nic}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID fetches a user by their MongoDB ObjectID string
func (r *userMongoRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
