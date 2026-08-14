package repository

import (
	"context"
	"errors"

	"math"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type miningLicenseMongoRepo struct {
	collection *mongo.Collection
}

// NewMiningLicenseRepository creates a new repository backed by the
// "mechanized_gem_mining_licenses" collection.
func NewMiningLicenseRepository(db *mongo.Database) domain.MiningLicenseRepository {
	return &miningLicenseMongoRepo{
		collection: db.Collection("mechanized_gem_mining_licenses"),
	}
}

// Create inserts a new license application document into MongoDB.
func (r *miningLicenseMongoRepo) Create(ctx context.Context, license *domain.MechanizedGemMiningLicense) error {
	license.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, license)
	return err
}

// GetByID fetches a single license by its MongoDB ObjectID string.
func (r *miningLicenseMongoRepo) GetByID(ctx context.Context, id string) (*domain.MechanizedGemMiningLicense, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid license ID format")
	}

	var license domain.MechanizedGemMiningLicense
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&license)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("license not found")
		}
		return nil, err
	}
	return &license, nil
}

// GetAll returns every license application document in the collection.
func (r *miningLicenseMongoRepo) GetAll(ctx context.Context) ([]domain.MechanizedGemMiningLicense, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var licenses []domain.MechanizedGemMiningLicense
	if err = cursor.All(ctx, &licenses); err != nil {
		return nil, err
	}
	if licenses == nil {
		licenses = []domain.MechanizedGemMiningLicense{}
	}
	return licenses, nil
}

// GetByTIN returns all license applications for a specific TIN number.
func (r *miningLicenseMongoRepo) GetByTIN(ctx context.Context, tin string, page int, limit int) (*domain.PaginatedMiningLicenses, error) {
	filter := bson.M{"tin": tin}
	
	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Calculate skip and total pages
	skip := (page - 1) * limit
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Setup options
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var licenses []domain.MechanizedGemMiningLicense
	if err = cursor.All(ctx, &licenses); err != nil {
		return nil, err
	}
	if licenses == nil {
		licenses = []domain.MechanizedGemMiningLicense{}
	}
	
	return &domain.PaginatedMiningLicenses{
		Data:       licenses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateStatus changes only the status field of a license document.
func (r *miningLicenseMongoRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid license ID format")
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("license not found")
	}
	return nil
}
