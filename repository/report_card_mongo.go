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

	pipeline := mongo.Pipeline{
		// Join each report card to its owning application, by converting
		// applicationId (string) to the applications collection's ObjectId.
		bson.D{{Key: "$lookup", Value: bson.M{
			"from": "applications",
			"let":  bson.M{"appId": "$applicationId"},
			"pipeline": mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.M{
					"$expr": bson.M{"$eq": bson.A{"$_id", bson.M{"$toObjectId": "$$appId"}}},
				}}},
			},
			"as": "application",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{
			"path":                       "$application",
			"preserveNullAndEmptyArrays": true,
		}}},
		bson.D{{Key: "$match", Value: bson.M{
			"$or": []bson.M{
				{"fullName": bson.M{"$regex": query, "$options": "i"}},
				{"accNumber": bson.M{"$regex": query, "$options": "i"}},
				{"nic": bson.M{"$regex": query, "$options": "i"}},
				{"applied_grade": bson.M{"$regex": query, "$options": "i"}},
				{"current_grade": bson.M{"$regex": query, "$options": "i"}},
				{"application.licenseRegionalOfficeAndZone": bson.M{"$regex": query, "$options": "i"}},
			},
		}}},
		bson.D{{Key: "$limit", Value: int64(10)}},
		bson.D{{Key: "$project", Value: bson.M{
			"fullName":       1,
			"accNumber":      1,
			"nic":            1,
			"applied_grade":  1,
			"current_grade":  1,
			"startDate":      1,
			"endDate":        1,
			"amount":         1,
			"totalDuration":  1,
			"totalAmount":    1,
			"regionalOffice": "$application.licenseRegionalOfficeAndZone",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
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

func (r *reportCardMongoRepo) GetByApplicationID(ctx context.Context, applicationID string, page int, limit int) (*domain.PaginatedReportCards, error) {
	filter := bson.M{"applicationId": applicationID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.M{"createdAt": -1}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []*domain.ReportCard
	if err = cursor.All(ctx, &cards); err != nil {
		return nil, err
	}

	if cards == nil {
		cards = []*domain.ReportCard{}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &domain.PaginatedReportCards{
		Data:       cards,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
