package repository

import (
	"context"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type miniSahanaMongoRepo struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewMiniSahanaRepository creates a new mini sahana form repository
// It connects to the "Minisahana" database and "applications" collection.
func NewMiniSahanaRepository(db *mongo.Database) domain.MiniSahanaRepository {
	// The requirement: "new database called, Minisahana and under that there needs to be collection as applications"
	// We get a new database instance using the existing client
	minisahanaDB := db.Client().Database("Minisahana")
	return &miniSahanaMongoRepo{
		db:         minisahanaDB,
		collection: minisahanaDB.Collection("applications"),
	}
}

func (r *miniSahanaMongoRepo) Create(ctx context.Context, form *domain.MiniSahanaForm) error {
	form.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, form)
	return err
}
