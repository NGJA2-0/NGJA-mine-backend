package repository

import (
	"context"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type applicationPaymentMongoRepo struct {
	collection *mongo.Collection
}

func NewApplicationPaymentRepository(db *mongo.Database) domain.ApplicationPaymentRepository {
	return &applicationPaymentMongoRepo{
		collection: db.Client().Database("Minisahana").Collection("report card data"),
	}
}

// GetByApplication unwinds the months of every report card of the application into one flat list,
// returns one page of it, and the totals over ALL months (paid / unpaid count and amount).
// Each month is multiplied by the amount of the card it belongs to.
func (r *applicationPaymentMongoRepo) GetByApplication(ctx context.Context, applicationID string, skip int64, limit int64) ([]*domain.ApplicationPaymentMonth, *domain.ApplicationPaymentSummary, error) {
	isPaid := bson.M{"$eq": bson.A{"$months.paid", true}}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"applicationId": applicationID}}},
		bson.D{{Key: "$unwind", Value: "$months"}},
		bson.D{{Key: "$facet", Value: bson.M{
			// totals over ALL months, never affected by pagination
			"summary": mongo.Pipeline{
				bson.D{{Key: "$group", Value: bson.M{
					"_id":          nil,
					"totalMonths":  bson.M{"$sum": 1},
					"totalAmount":  bson.M{"$sum": "$amount"},
					"paidMonths":   bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 1, 0}}},
					"paidAmount":   bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, "$amount", 0}}},
					"unpaidMonths": bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 0, 1}}},
					"unpaidAmount": bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 0, "$amount"}}},
				}}},
			},
			// one page of the flat month list: newest year first, January -> December inside a year
			"rows": mongo.Pipeline{
				bson.D{{Key: "$sort", Value: bson.D{
					{Key: "months.year", Value: -1},
					{Key: "months.month", Value: 1},
					{Key: "_id", Value: 1},
				}}},
				bson.D{{Key: "$skip", Value: skip}},
				bson.D{{Key: "$limit", Value: limit}},
				bson.D{{Key: "$project", Value: bson.M{
					"_id":          0,
					"cardId":       "$_id",
					"refNumber":    1,
					"currentGrade": "$current_grade",
					"amount":       1,
					"year":         "$months.year",
					"month":        "$months.month",
					"label":        "$months.label",
					"paid":         "$months.paid",
					"paidAt":       "$months.paidAt",
					"paidBy":       "$months.paidBy",
					"paidById":     "$months.paidById",
				}}},
			},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var out []struct {
		Summary []domain.ApplicationPaymentSummary `bson:"summary"`
		Rows    []*domain.ApplicationPaymentMonth  `bson:"rows"`
	}
	if err = cursor.All(ctx, &out); err != nil {
		return nil, nil, err
	}

	summary := &domain.ApplicationPaymentSummary{} // stays all zeros when nothing matches
	rows := []*domain.ApplicationPaymentMonth{}
	if len(out) > 0 {
		if len(out[0].Summary) > 0 {
			summary = &out[0].Summary[0]
		}
		if out[0].Rows != nil {
			rows = out[0].Rows
		}
	}
	return rows, summary, nil
}