package repository

import (
	"context"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type paymentOverviewMongoRepo struct {
	collection *mongo.Collection
}

func NewPaymentOverviewRepository(db *mongo.Database) domain.PaymentOverviewRepository {
	return &paymentOverviewMongoRepo{
		collection: db.Client().Database("Minisahana").Collection("report card data"),
	}
}

// GetOverview sums paid / unpaid money per (year, month) over ALL report cards.
// Every month entry counts as the card's monthly `amount`.
func (r *paymentOverviewMongoRepo) GetOverview(ctx context.Context, fromYear, toYear int) ([]domain.PaymentOverviewRow, error) {
	yearRange := bson.M{"$gte": fromYear, "$lte": toYear}
	isPaid := bson.M{"$eq": bson.A{"$months.paid", true}}

	pipeline := mongo.Pipeline{
		// cheap pre-filter (can use an index on months.year)
		bson.D{{Key: "$match", Value: bson.M{"months.year": yearRange}}},
		bson.D{{Key: "$unwind", Value: "$months"}},
		// the pre-filter matched whole cards, so filter the unwound months again
		bson.D{{Key: "$match", Value: bson.M{"months.year": yearRange}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{"year": "$months.year", "month": "$months.month"},
			"paid": bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, "$amount", 0}}},
			"unpaid": bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 0, "$amount"}}},
		}}},
		bson.D{{Key: "$project", Value: bson.M{
			"_id":    0,
			"year":   "$_id.year",
			"month":  "$_id.month",
			"paid":   1,
			"unpaid": 1,
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	rows := []domain.PaymentOverviewRow{}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}