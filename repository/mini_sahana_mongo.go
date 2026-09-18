package repository

import (
	"context"
	"errors"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *miniSahanaMongoRepo) GetLatestRefNumber(ctx context.Context) (string, error) {
	opts := options.FindOne().SetSort(bson.M{"createdAt": -1})
	var result domain.MiniSahanaForm
	err := r.collection.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil // No records yet
		}
		return "", err
	}
	return result.RefNumber, nil
}

func (r *miniSahanaMongoRepo) GetByID(ctx context.Context, id string) (*domain.MiniSahanaForm, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid id format")
	}

	var form domain.MiniSahanaForm
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&form)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("record not found")
		}
		return nil, err
	}
	return &form, nil
}

func (r *miniSahanaMongoRepo) Search(ctx context.Context, query string) ([]*domain.MiniSahanaSearchSuggestion, error) {
	var suggestions []*domain.MiniSahanaSearchSuggestion

	// Regex for partial matching, case-insensitive
	filter := bson.M{
		"$or": []bson.M{
			{"applicantFullNameSinhala": bson.M{"$regex": query, "$options": "i"}},
			{"bankAccountNumber": bson.M{"$regex": query, "$options": "i"}},
			{"nic": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().
		SetLimit(10).
		SetProjection(bson.M{
			"applicantFullNameSinhala": 1,
			"bankAccountNumber":        1,
			"nic":                      1,
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
		suggestions = []*domain.MiniSahanaSearchSuggestion{}
	}
	return suggestions, nil
}

func (r *miniSahanaMongoRepo) Update(ctx context.Context, id string, form *domain.MiniSahanaForm) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	form.ID = oid
	_, err = r.collection.ReplaceOne(ctx, bson.M{"_id": oid}, form)
	return err
}

func (r *miniSahanaMongoRepo) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid id format")
	}

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("record not found")
	}
	return nil
}
