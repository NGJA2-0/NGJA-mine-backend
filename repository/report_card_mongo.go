package repository

import (
	"context"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type reportCardMongoRepo struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewReportCardRepository creates a new report card repository.
// It connects to the existing "Minisahana" database and uses the "report card data" collection.
func NewReportCardRepository(db *mongo.Database) domain.ReportCardRepository {
	minisahanaDB := db.Client().Database("Minisahana")
	return &reportCardMongoRepo{
		db:         minisahanaDB,
		collection: minisahanaDB.Collection("report card data"),
	}
}

func (r *reportCardMongoRepo) Create(ctx context.Context, card *domain.ReportCard) error {
	card.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, card)
	return err
}

func (r *reportCardMongoRepo) GetLatestRefNumber(ctx context.Context) (string, error) {
	opts := options.FindOne().SetSort(bson.M{"createdAt": -1})
	var result domain.ReportCard
	err := r.collection.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil // No records yet
		}
		return "", err
	}
	return result.RefNumber, nil
}
