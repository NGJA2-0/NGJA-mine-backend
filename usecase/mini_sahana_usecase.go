package usecase

import (
	"context"
	"os"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"   
	"regexp"
	"strconv"
	"strings"
	"time"

	"my-fiber-app/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// fields that are metadata, never part of a diff
var diffExcludedFields = map[string]bool{
	"id": true, "refNumber": true, "createdAt": true,
	"createdBy": true, "updatedBy": true, "changes": true,"documents": true,
}

func toDiffMap(form *domain.MiniSahanaForm) (map[string]interface{}, error) {
	data, err := json.Marshal(form)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// diffFields compares old vs new maps, skipping metadata fields, and returns
// only the keys that actually changed.
func diffFields(oldMap, newMap map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	oldVals := map[string]interface{}{}
	newVals := map[string]interface{}{}
	for key, newVal := range newMap {
		if diffExcludedFields[key] {
			continue
		}
		oldVal := oldMap[key]
		if !reflect.DeepEqual(oldVal, newVal) {
			oldVals[key] = oldVal
			newVals[key] = newVal
		}
	}
	return oldVals, newVals
}

type miniSahanaUsecase struct {
	repo     domain.MiniSahanaRepository
	userRepo domain.UserRepository
}

func NewMiniSahanaUsecase(repo domain.MiniSahanaRepository, userRepo domain.UserRepository) domain.MiniSahanaUsecase {
	return &miniSahanaUsecase{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (u *miniSahanaUsecase) Create(ctx context.Context, form *domain.MiniSahanaForm, userID string, files map[string]*domain.DocUpload) error {
	// Validate NIC: old format (9 digits + V/X) or new format (12 digits)
	form.NIC = strings.ToUpper(strings.TrimSpace(form.NIC))
	oldNicRegex := regexp.MustCompile(`^[0-9]{9}[VX]$`)
	newNicRegex := regexp.MustCompile(`^[0-9]{12}$`)

	if !oldNicRegex.MatchString(form.NIC) && !newNicRegex.MatchString(form.NIC) {
		return errors.New("nic format is invalid")
	}

	// Validate dateOfBirth: valid date, not in the future
	dob, err := time.Parse("2006-01-02", form.DateOfBirth)
	if err != nil {
		return errors.New("dateOfBirth must be a valid ISO date (YYYY-MM-DD)")
	}
	if dob.After(time.Now()) {
		return errors.New("dateOfBirth cannot be in the future")
	}

	// Validate eligibilityCategory: at least one value from ["a", "b", "c"]
	validCategories := map[string]bool{"a": true, "b": true, "c": true}
	hasValidCategory := false
	for _, cat := range form.EligibilityCategory {
		if validCategories[cat] {
			hasValidCategory = true
			break
		}
	}
	if !hasValidCategory {
		return errors.New("eligibilityCategory must contain at least one of 'a', 'b', or 'c'")
	}

	// Fetch User to get name
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("failed to fetch user details")
	}
	form.CreatedBy = user.Name

	// Generate RefNumber
	latestRef, _ := u.repo.GetLatestRefNumber(ctx)
	nextNum := 1
	if latestRef != "" {
		// Example: "A1" or "A1.1" -> parse the integer after 'A' and before '.'
		refStr := strings.TrimPrefix(latestRef, "A")
		if idx := strings.Index(refStr, "."); idx != -1 {
			refStr = refStr[:idx]
		}
		if num, err := strconv.Atoi(refStr); err == nil {
			nextNum = num + 1
		}
	}
	form.RefNumber = fmt.Sprintf("A%d", nextNum)

		for _, slot := range domain.MiniSahanaDocumentSlots {
		if slot.Required && files[slot.Key] == nil {
			return fmt.Errorf("%s is required", slot.Label)
		}
	}

	form.ID = primitive.NewObjectID()
	appID := form.ID.Hex()
	form.CreatedAt = time.Now()
	form.Changes = []domain.MiniSahanaChangeEntry{}
	form.Documents = []domain.MiniSahanaDocument{}

	for _, slot := range domain.MiniSahanaDocumentSlots {
		f := files[slot.Key]
		if f == nil {
			continue
		}
		rel := domain.DocumentPath(appID, slot.Key, 1)
		if err := f.Save(rel); err != nil {
			_ = os.RemoveAll(domain.DocumentDir(appID))
			return fmt.Errorf("failed to save %s", slot.Label)
		}
		form.Documents = append(form.Documents, domain.MiniSahanaDocument{
			Key: slot.Key, Label: slot.Label, CurrentVersion: 1,
			Versions: []domain.MiniSahanaDocVersion{{
				Version: 1, FileName: f.FileName, StoredPath: rel, Size: f.Size,
				UploadedBy: user.Name, UploadedAt: time.Now(),
			}},
		})
	}

	if err := u.repo.Create(ctx, form); err != nil {
		_ = os.RemoveAll(domain.DocumentDir(appID))
		return err
	}
	return nil
}

func (u *miniSahanaUsecase) GetByID(ctx context.Context, id string) (*domain.MiniSahanaForm, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *miniSahanaUsecase) Search(ctx context.Context, query string) ([]*domain.MiniSahanaSearchSuggestion, error) {
	if strings.TrimSpace(query) == "" {
		return []*domain.MiniSahanaSearchSuggestion{}, nil
	}
	return u.repo.Search(ctx, query)
}

func (u *miniSahanaUsecase) Update(ctx context.Context, id string, form *domain.MiniSahanaForm, userID string) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// ── existing validation block stays exactly as-is ──
	form.NIC = strings.ToUpper(strings.TrimSpace(form.NIC))
	oldNicRegex := regexp.MustCompile(`^[0-9]{9}[VX]$`)
	newNicRegex := regexp.MustCompile(`^[0-9]{12}$`)
	if !oldNicRegex.MatchString(form.NIC) && !newNicRegex.MatchString(form.NIC) {
		return errors.New("nic format is invalid")
	}

	dob, err := time.Parse("2006-01-02", form.DateOfBirth)
	if err != nil {
		return errors.New("dateOfBirth must be a valid ISO date (YYYY-MM-DD)")
	}
	if dob.After(time.Now()) {
		return errors.New("dateOfBirth cannot be in the future")
	}

	validCategories := map[string]bool{"a": true, "b": true, "c": true}
	hasValidCategory := false
	for _, cat := range form.EligibilityCategory {
		if validCategories[cat] {
			hasValidCategory = true
			break
		}
	}
	if !hasValidCategory {
		return errors.New("eligibilityCategory must contain at least one of 'a', 'b', or 'c'")
	}
	// ── end existing validation block ──

	// ── ADD: Normal Edit can never change bankAccountNumber ──
	form.BankAccountNumber = existing.BankAccountNumber
	form.Documents = existing.Documents

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("failed to fetch user details")
	}

	// ── existing RefNumber versioning block stays exactly as-is ──
	baseRef := existing.RefNumber
	version := 0
	if idx := strings.Index(baseRef, "."); idx != -1 {
		verStr := baseRef[idx+1:]
		if v, err := strconv.Atoi(verStr); err == nil {
			version = v
		}
		baseRef = baseRef[:idx]
	}
	form.RefNumber = fmt.Sprintf("%s.%d", baseRef, version+1)

	form.CreatedBy = existing.CreatedBy
	form.CreatedAt = existing.CreatedAt
	// ── end existing block ──

	// ── ADD: diff and append to Changes ──
	oldMap, _ := toDiffMap(existing)
	newMap, _ := toDiffMap(form)
	oldVals, newVals := diffFields(oldMap, newMap)

	form.Changes = existing.Changes
	if len(newVals) > 0 {
		form.Changes = append(form.Changes, domain.MiniSahanaChangeEntry{
			EditType:  "normal",
			ChangedBy: user.Name,
			ChangedAt: time.Now(),
			OldValues: oldVals,
			NewValues: newVals,
		})
	}
	form.UpdatedBy = user.Name
	// ── end ADD ──

	return u.repo.Update(ctx, id, form)
}

func (u *miniSahanaUsecase) UpdateAccountNumber(ctx context.Context, id string, newAccountNumber string, userID string) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	newAccountNumber = strings.TrimSpace(newAccountNumber)
	if newAccountNumber == "" {
		return errors.New("bankAccountNumber is required")
	}

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("failed to fetch user details")
	}

	form := *existing // shallow copy of every existing field
	form.BankAccountNumber = newAccountNumber

	baseRef := existing.RefNumber
	version := 0
	if idx := strings.Index(baseRef, "."); idx != -1 {
		verStr := baseRef[idx+1:]
		if v, err := strconv.Atoi(verStr); err == nil {
			version = v
		}
		baseRef = baseRef[:idx]
	}
	form.RefNumber = fmt.Sprintf("%s.%d", baseRef, version+1)

	form.Changes = existing.Changes
	if existing.BankAccountNumber != newAccountNumber {
		form.Changes = append(form.Changes, domain.MiniSahanaChangeEntry{
			EditType:  "acc_number",
			ChangedBy: user.Name,
			ChangedAt: time.Now(),
			OldValues: map[string]interface{}{"bankAccountNumber": existing.BankAccountNumber},
			NewValues: map[string]interface{}{"bankAccountNumber": newAccountNumber},
		})
	}
	form.UpdatedBy = user.Name
	form.CreatedBy = existing.CreatedBy
	form.CreatedAt = existing.CreatedAt

	return u.repo.Update(ctx, id, &form)
}

func (u *miniSahanaUsecase) OpenDocument(ctx context.Context, id, docKey string, version int) (*domain.MiniSahanaDocVersion, error) {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, d := range existing.Documents {
		if d.Key != docKey {
			continue
		}
		for _, v := range d.Versions {
			if v.Version == version {
				vv := v
				return &vv, nil
			}
		}
	}
	return nil, errors.New("document not found")
}

func (u *miniSahanaUsecase) UpdateDocument(ctx context.Context, id, docKey string, file *domain.DocUpload, userID string) (*domain.MiniSahanaForm, error) {
	var slot *domain.MiniSahanaDocumentSlot
	for i := range domain.MiniSahanaDocumentSlots {
		if domain.MiniSahanaDocumentSlots[i].Key == docKey {
			slot = &domain.MiniSahanaDocumentSlots[i]
		}
	}
	if slot == nil {
		return nil, errors.New("unknown document type")
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch user details")
	}

	idx := -1
	for i, d := range existing.Documents {
		if d.Key == docKey {
			idx = i
		}
	}
	if idx == -1 { // first upload for an older record like A7
		existing.Documents = append(existing.Documents, domain.MiniSahanaDocument{Key: slot.Key, Label: slot.Label})
		idx = len(existing.Documents) - 1
	}
	doc := &existing.Documents[idx]

	oldDesc := ""
	if n := len(doc.Versions); n > 0 {
		last := doc.Versions[n-1]
		oldDesc = fmt.Sprintf("v%d · %s", last.Version, last.FileName)
	}

	newVer := len(doc.Versions) + 1
	rel := domain.DocumentPath(existing.ID.Hex(), docKey, newVer)
	if err := file.Save(rel); err != nil {
		return nil, errors.New("failed to save PDF file")
	}
	doc.Versions = append(doc.Versions, domain.MiniSahanaDocVersion{
		Version: newVer, FileName: file.FileName, StoredPath: rel, Size: file.Size,
		UploadedBy: user.Name, UploadedAt: time.Now(),
	})
	doc.CurrentVersion = newVer

	existing.Changes = append(existing.Changes, domain.MiniSahanaChangeEntry{
		EditType:  "document",
		ChangedBy: user.Name,
		ChangedAt: time.Now(),
		OldValues: map[string]interface{}{"documents." + docKey: oldDesc},
		NewValues: map[string]interface{}{"documents." + docKey: fmt.Sprintf("v%d · %s", newVer, file.FileName)},
	})
	existing.UpdatedBy = user.Name
	existing.RefNumber = nextRef(existing.RefNumber)

	if err := u.repo.Update(ctx, id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (u *miniSahanaUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func nextRef(ref string) string {
	base, version := ref, 0
	if idx := strings.Index(ref, "."); idx != -1 {
		if v, err := strconv.Atoi(ref[idx+1:]); err == nil {
			version = v
		}
		base = ref[:idx]
	}
	return fmt.Sprintf("%s.%d", base, version+1)
}