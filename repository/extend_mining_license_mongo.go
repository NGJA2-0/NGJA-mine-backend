package repository

import (
	"context"
	"errors"

	"math"
	"strconv"
	"strings"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type extendMiningLicenseMongoRepo struct {
	// Intentionally shares the same underlying collection as mining licenses
	// so that all versions (original, edited, extended) live in one place.
	collection         *mongo.Collection
	countersCollection *mongo.Collection
}

// NewExtendMiningLicenseRepository creates a new repository that saves extend
// records into the SAME "mechanized_gem_mining_licenses" collection so that
// original records and all their versioned edits/extensions are co-located.
func NewExtendMiningLicenseRepository(db *mongo.Database) domain.ExtendMiningLicenseRepository {
	return &extendMiningLicenseMongoRepo{
		collection:         db.Collection("mechanized_gem_mining_licenses"),
		countersCollection: db.Collection("counters"),
	}
}

// Create inserts a new license extend application document into MongoDB.
func (r *extendMiningLicenseMongoRepo) Create(ctx context.Context, license *domain.ExtendMiningLicense) error {
	license.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, license)
	return err
}

// GetByID fetches a single license by its MongoDB ObjectID string.
func (r *extendMiningLicenseMongoRepo) GetByID(ctx context.Context, id string) (*domain.ExtendMiningLicense, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid license ID format")
	}

	var license domain.ExtendMiningLicense
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
func (r *extendMiningLicenseMongoRepo) GetAll(ctx context.Context) ([]domain.ExtendMiningLicense, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var licenses []domain.ExtendMiningLicense
	if err = cursor.All(ctx, &licenses); err != nil {
		return nil, err
	}
	if licenses == nil {
		licenses = []domain.ExtendMiningLicense{}
	}
	return licenses, nil
}

// GetByTIN returns all license applications for a specific TIN number.
func (r *extendMiningLicenseMongoRepo) GetByTIN(ctx context.Context, tin string, page int, limit int) (*domain.PaginatedExtendMiningLicenses, error) {
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

	var licenses []domain.ExtendMiningLicense
	if err = cursor.All(ctx, &licenses); err != nil {
		return nil, err
	}
	if licenses == nil {
		licenses = []domain.ExtendMiningLicense{}
	}

	return &domain.PaginatedExtendMiningLicenses{
		Data:       licenses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetForMap returns license applications matching the given filters,
// restricted to records that have at least one GPS point.
func (r *extendMiningLicenseMongoRepo) GetForMap(ctx context.Context, filters domain.MapFilters) ([]domain.ExtendMiningLicense, error) {
	filter := bson.M{
		"gpsPoints": bson.M{"$exists": true, "$ne": bson.A{}},
	}

	if filters.District != "" {
		filter["district"] = filters.District
	}
	if filters.Village != "" {
		filter["village"] = filters.Village
	}
	if filters.TIN != "" {
		filter["tin"] = filters.TIN
	}
	if filters.NIC != "" {
		filter["nic"] = filters.NIC
	}
	if filters.GMLNumber != "" {
		filter["gmlNumber"] = filters.GMLNumber
	}
	if filters.LandName != "" {
		filter["landName"] = bson.M{"$regex": filters.LandName, "$options": "i"}
	}
	if filters.RegionalOffice != "" {
		filter["regionalOffice"] = filters.RegionalOffice
	}

	if filters.Search != "" {
		searchRegex := bson.M{"$regex": filters.Search, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"tin": searchRegex},
			bson.M{"nic": searchRegex},
			bson.M{"gmlNumber": searchRegex},
			bson.M{"landName": searchRegex},
		}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var licenses []domain.ExtendMiningLicense
	if err = cursor.All(ctx, &licenses); err != nil {
		return nil, err
	}
	if licenses == nil {
		licenses = []domain.ExtendMiningLicense{}
	}
	return licenses, nil
}

// UpdateStatus changes only the status field of a license document.
func (r *extendMiningLicenseMongoRepo) UpdateStatus(ctx context.Context, id string, status string) error {
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

// GetNextBaseReferenceNumber gets the next auto-incrementing integer for brand-new records.
// Shared counter with mining licenses so all records have globally unique base numbers.
func (r *extendMiningLicenseMongoRepo) GetNextBaseReferenceNumber(ctx context.Context) (int64, error) {
	filter := bson.M{"_id": "mining_license_ref"}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int64 `bson:"seq"`
	}

	err := r.countersCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}

// GetMaxVersionByBaseRef finds the highest version suffix currently stored for
// a given base reference number (e.g. "REF_2").
// It queries all documents whose referenceNumber starts with "<baseRef>."
// and returns the maximum integer suffix found.
// Returns 0 if no versioned documents exist yet.
func (r *extendMiningLicenseMongoRepo) GetMaxVersionByBaseRef(ctx context.Context, baseRef string) (int, error) {
	// Match all docs whose referenceNumber begins with "baseRef."
	// e.g. baseRef = "REF_2" matches "REF_2.1", "REF_2.2", but NOT "REF_20"
	prefix := baseRef + "."
	filter := bson.M{
		"referenceNumber": bson.M{
			"$regex": "^" + prefix,
		},
	}

	// We only need the referenceNumber field
	findOptions := options.Find().SetProjection(bson.M{"referenceNumber": 1})
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	maxVersion := 0
	for cursor.Next(ctx) {
		var doc struct {
			ReferenceNumber string `bson:"referenceNumber"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		// Extract the suffix after the dot
		after, found := strings.CutPrefix(doc.ReferenceNumber, prefix)
		if !found {
			continue
		}
		// Only consider single-level versions (e.g. "1" in "REF_2.1")
		if strings.Contains(after, ".") {
			continue
		}
		if v, err := strconv.Atoi(after); err == nil && v > maxVersion {
			maxVersion = v
		}
	}

	return maxVersion, cursor.Err()
}
