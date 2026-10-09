package repository

import (
	"context"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type annualReportMongoRepo struct {
	collection *mongo.Collection
}

// NewAnnualReportRepository uses the same "Minisahana" database and "report card data" collection
// as the report card repository.
func NewAnnualReportRepository(db *mongo.Database) domain.AnnualReportRepository {
	return &annualReportMongoRepo{
		collection: db.Client().Database("Minisahana").Collection("report card data"),
	}
}

// perCardSums are the $group accumulators over the per-card fields "n" (months in the year)
// and "paidN" (paid months in the year). cardsField is the name of the card counter.
func perCardSums(cardsField string) bson.M {
	unpaidN := bson.M{"$subtract": bson.A{"$n", "$paidN"}}
	return bson.M{
		cardsField:     bson.M{"$sum": 1},
		"count":        bson.M{"$sum": "$n"},
		"paidCount":    bson.M{"$sum": "$paidN"},
		"unpaidCount":  bson.M{"$sum": unpaidN},
		"totalAmount":  bson.M{"$sum": bson.M{"$multiply": bson.A{"$amount", "$n"}}},
		"paidAmount":   bson.M{"$sum": bson.M{"$multiply": bson.A{"$amount", "$paidN"}}},
		"unpaidAmount": bson.M{"$sum": bson.M{"$multiply": bson.A{"$amount", unpaidN}}},
	}
}

func groupStage(id interface{}, sums bson.M) bson.D {
	doc := bson.M{"_id": id}
	for k, v := range sums {
		doc[k] = v
	}
	return bson.D{{Key: "$group", Value: doc}}
}

// GetAnnualReport returns the report cards that have at least one month in the given year
// (optionally only one current_grade), each with ONLY its months of that year.
// limit = 0 means "no pagination". The aggregates are always calculated over ALL matching cards.
func (r *annualReportMongoRepo) GetAnnualReport(ctx context.Context, year int, grade string, skip int64, limit int64) ([]*domain.AnnualReportRow, *domain.AnnualAggregates, error) {
	filter := bson.M{"months": bson.M{"$elemMatch": bson.M{"year": year}}}
	if grade != "" {
		filter["current_grade"] = grade
	}
	match := bson.D{{Key: "$match", Value: filter}}

	// The months of the selected year only
	monthsOfYear := bson.M{"$filter": bson.M{
		"input": "$months",
		"as":    "m",
		"cond":  bson.M{"$eq": bson.A{"$$m.year", year}},
	}}
	aggOpts := options.Aggregate().SetAllowDiskUse(true)

	// 1) Aggregates over ALL matching cards: grand total, month-wise and grade-wise
	isPaid := bson.M{"$eq": bson.A{"$ms.paid", true}}
	aggPipeline := mongo.Pipeline{
		match,
		bson.D{{Key: "$project", Value: bson.M{
			"current_grade": 1,
			"amount":        1,
			"ms":            monthsOfYear,
		}}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"n": bson.M{"$size": "$ms"},
			"paidN": bson.M{"$size": bson.M{"$filter": bson.M{
				"input": "$ms",
				"as":    "p",
				"cond":  bson.M{"$eq": bson.A{"$$p.paid", true}},
			}}},
		}}},
		bson.D{{Key: "$facet", Value: bson.M{
			"totals": mongo.Pipeline{
				groupStage(nil, perCardSums("totalCards")),
			},
			"byGrade": mongo.Pipeline{
				groupStage("$current_grade", perCardSums("cards")),
			},
			"byMonth": mongo.Pipeline{
				bson.D{{Key: "$unwind", Value: "$ms"}},
				groupStage("$ms.month", bson.M{
					"count":        bson.M{"$sum": 1},
					"paidCount":    bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 1, 0}}},
					"unpaidCount":  bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 0, 1}}},
					"totalAmount":  bson.M{"$sum": "$amount"},
					"paidAmount":   bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, "$amount", 0}}},
					"unpaidAmount": bson.M{"$sum": bson.M{"$cond": bson.A{isPaid, 0, "$amount"}}},
				}),
				bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
			},
		}}},
	}

	aggCursor, err := r.collection.Aggregate(ctx, aggPipeline, aggOpts)
	if err != nil {
		return nil, nil, err
	}
	var facets []struct {
		Totals  []domain.AnnualSummary    `bson:"totals"`
		ByMonth []domain.AnnualMonthTotal `bson:"byMonth"`
		ByGrade []domain.AnnualGradeTotal `bson:"byGrade"`
	}
	if err = aggCursor.All(ctx, &facets); err != nil {
		return nil, nil, err
	}

	agg := &domain.AnnualAggregates{}
	if len(facets) > 0 {
		if len(facets[0].Totals) > 0 {
			agg.Summary = facets[0].Totals[0]
		}
		agg.ByMonth = facets[0].ByMonth
		agg.ByGrade = facets[0].ByGrade
	}

	// 2) The rows (one page, or everything when limit = 0)
	rowsPipeline := mongo.Pipeline{
		match,
		bson.D{{Key: "$sort", Value: bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}}},
	}
	if skip > 0 {
		rowsPipeline = append(rowsPipeline, bson.D{{Key: "$skip", Value: skip}})
	}
	if limit > 0 {
		rowsPipeline = append(rowsPipeline, bson.D{{Key: "$limit", Value: limit}})
	}
	// Replace the months array with only the months of the selected year
	rowsPipeline = append(rowsPipeline, bson.D{{Key: "$addFields", Value: bson.M{"months": monthsOfYear}}})

	rowsCursor, err := r.collection.Aggregate(ctx, rowsPipeline, aggOpts)
	if err != nil {
		return nil, nil, err
	}
	var rows []*domain.AnnualReportRow
	if err = rowsCursor.All(ctx, &rows); err != nil {
		return nil, nil, err
	}
	if rows == nil {
		rows = []*domain.AnnualReportRow{}
	}

	return rows, agg, nil
}