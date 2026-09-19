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

func (r *reportCardMongoRepo) Search(ctx context.Context, query string) ([]*domain.ReportCardSearchSuggestion, error) {
	var suggestions []*domain.ReportCardSearchSuggestion

	// Regex for partial matching, case-insensitive
	filter := bson.M{
		"$or": []bson.M{
			{"fullName":  bson.M{"$regex": query, "$options": "i"}},
			{"accNumber": bson.M{"$regex": query, "$options": "i"}},
			{"nic":       bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().
		SetLimit(10).
		SetProjection(bson.M{
			"fullName":      1,
			"accNumber":     1,
			"nic":           1,
			"grade":         1,
			"startDate":     1,
			"endDate":       1,
			"amount":        1,
			"totalDuration": 1,
			"totalAmount":   1,
		})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &suggestions); err != nil {
		return nil, err
	}

	if suggestions == nil {
		suggestions = []*domain.ReportCardSearchSuggestion{}
	}
	return suggestions, nil
}
