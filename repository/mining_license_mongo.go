package repository

import (
	"context"
	"errors"
	"regexp"

	"math"
	"sort"
	"strconv"
	"strings"

	"my-fiber-app/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type miningLicenseMongoRepo struct {
	collection         *mongo.Collection
	countersCollection *mongo.Collection
}

// NewMiningLicenseRepository creates a new repository backed by the
// "mechanized_gem_mining_licenses" collection.
func NewMiningLicenseRepository(db *mongo.Database) domain.MiningLicenseRepository {
	return &miningLicenseMongoRepo{
		collection:         db.Collection("mechanized_gem_mining_licenses"),
		countersCollection: db.Collection("counters"),
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

// GetByTIN returns license applications for a specific TIN number, deduplicated
// so that only the latest edition per base reference number is returned
// (e.g. REF_2.1, REF_2.2 -> only REF_2.2 is returned; REF_5 with no suffix
// counts as version 0 and is returned as-is if it's the only edition).
// Results are sorted by base reference number descending (newest ref first),
// and pagination is applied to the deduplicated set.
func (r *miningLicenseMongoRepo) GetByTIN(ctx context.Context, tin string, page int, limit int) (*domain.PaginatedMiningLicenses, error) {
	filter := bson.M{"tin": tin}

	// Fetch everything for this TIN first — we can't paginate at the DB level
	// because dedup has to happen across the full result set.
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var all []domain.MechanizedGemMiningLicense
	if err = cursor.All(ctx, &all); err != nil {
		return nil, err
	}

	// Keep only the highest version per base reference number.
	type latestEntry struct {
		license domain.MechanizedGemMiningLicense
		version int
	}
	latestByBaseRef := make(map[string]latestEntry)

	for _, lic := range all {
		baseRef := lic.ReferenceNumber
		version := 0
		if idx := strings.Index(lic.ReferenceNumber, "."); idx != -1 {
			baseRef = lic.ReferenceNumber[:idx]
			if v, err := strconv.Atoi(lic.ReferenceNumber[idx+1:]); err == nil {
				version = v
			}
		}

		if existing, ok := latestByBaseRef[baseRef]; !ok || version > existing.version {
			latestByBaseRef[baseRef] = latestEntry{license: lic, version: version}
		}
	}

	deduped := make([]domain.MechanizedGemMiningLicense, 0, len(latestByBaseRef))
	for _, entry := range latestByBaseRef {
		deduped = append(deduped, entry.license)
	}

	// Sort by base reference number, newest (highest numeric value) first.
	sort.Slice(deduped, func(i, j int) bool {
		return baseRefNumber(deduped[i].ReferenceNumber) > baseRefNumber(deduped[j].ReferenceNumber)
	})

	// Paginate the deduplicated, sorted slice in-memory.
	total := int64(len(deduped))
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	skip := (page - 1) * limit
	if skip > len(deduped) {
		skip = len(deduped)
	}
	end := skip + limit
	if end > len(deduped) {
		end = len(deduped)
	}
	pageItems := deduped[skip:end]
	if pageItems == nil {
		pageItems = []domain.MechanizedGemMiningLicense{}
	}

	return &domain.PaginatedMiningLicenses{
		Data:       pageItems,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// baseRefNumber extracts the numeric part of a base reference number,
// e.g. "REF_9" -> 9, "REF_2.1" -> 2. Returns -1 if it can't be parsed,
// so unparsable refs sort last.
func baseRefNumber(referenceNumber string) int {
	baseRef := referenceNumber
	if idx := strings.Index(referenceNumber, "."); idx != -1 {
		baseRef = referenceNumber[:idx]
	}
	parts := strings.Split(baseRef, "_")
	if len(parts) < 2 {
		return -1
	}
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return -1
	}
	return n
}

// GetForMap returns license applications matching the given filters,
// restricted to records that have at least one GPS point.
func (r *miningLicenseMongoRepo) GetForMap(ctx context.Context, filters domain.MapFilters) ([]domain.MechanizedGemMiningLicense, error) {
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

	var licenses []domain.MechanizedGemMiningLicense
	if err = cursor.All(ctx, &licenses); err != nil {
		return nil, err
	}
	if licenses == nil {
		licenses = []domain.MechanizedGemMiningLicense{}
	}
	return licenses, nil
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

// GetNextBaseReferenceNumber gets the next auto-incrementing integer for reference numbers.
func (r *miningLicenseMongoRepo) GetNextBaseReferenceNumber(ctx context.Context) (int64, error) {
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
func (r *miningLicenseMongoRepo) GetMaxVersionByBaseRef(ctx context.Context, baseRef string) (int, error) {
	prefix := baseRef + "."
	filter := bson.M{
		"referenceNumber": bson.M{
			"$regex": "^" + prefix,
		},
	}

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
		after, found := strings.CutPrefix(doc.ReferenceNumber, prefix)
		if !found {
			continue
		}
		// Only count single-level suffixes (e.g. "1" in "REF_2.1")
		if strings.Contains(after, ".") {
			continue
		}
				if v, err := strconv.Atoi(after); err == nil && v > maxVersion {
			maxVersion = v
		}
	}

	return maxVersion, cursor.Err()
}

// GetByBaseReferenceNumber returns a slim, paginated view of every edition
// (the base ref itself plus any "<baseRef>.N" versions) for a given base
// reference number, sorted by version ascending (base ref first, then .1, .2...).
// Only the fields required by MiningLicenseSummary are fetched from Mongo.
func (r *miningLicenseMongoRepo) GetByBaseReferenceNumber(ctx context.Context, baseRef string, page int, limit int) (*domain.PaginatedMiningLicenseSummaries, error) {
	// Matches exactly "REF_4" or "REF_4.<anything>", but not "REF_40".
	pattern := "^" + regexp.QuoteMeta(baseRef) + "($|\\.)"
	filter := bson.M{
		"referenceNumber": bson.M{"$regex": pattern},
	}

	projection := bson.M{
		"_id":              1,
		"referenceNumber":  1,
		"applicantName":    1,
		"privateSaleValue": 1,
		"createdBy":        1,
		"createdAt":        1,
		"updatedAt":        1,
		"status":           1,
	}
	findOptions := options.Find().SetProjection(projection)

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var all []domain.MiningLicenseSummary
	if err = cursor.All(ctx, &all); err != nil {
		return nil, err
	}

	// Sort by version ascending: base ref (version 0) first, then .1, .2, ...
	sort.Slice(all, func(i, j int) bool {
		return summaryVersion(all[i].ReferenceNumber) < summaryVersion(all[j].ReferenceNumber)
	})

	total := int64(len(all))
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	skip := (page - 1) * limit
	if skip > len(all) {
		skip = len(all)
	}
	end := skip + limit
	if end > len(all) {
		end = len(all)
	}
	pageItems := all[skip:end]
	if pageItems == nil {
		pageItems = []domain.MiningLicenseSummary{}
	}

	return &domain.PaginatedMiningLicenseSummaries{
		Data:       pageItems,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// summaryVersion extracts the numeric version suffix from a reference number,
// e.g. "REF_4" -> 0, "REF_4.1" -> 1, "REF_4.12" -> 12.
func summaryVersion(referenceNumber string) int {
	idx := strings.Index(referenceNumber, ".")
	if idx == -1 {
		return 0
	}
	v, err := strconv.Atoi(referenceNumber[idx+1:])
	if err != nil {
		return 0
	}
	return v
}

