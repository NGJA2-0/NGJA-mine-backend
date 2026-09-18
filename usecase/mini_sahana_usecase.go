package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"my-fiber-app/domain"
)

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

func (u *miniSahanaUsecase) Create(ctx context.Context, form *domain.MiniSahanaForm, userID string) error {
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

	form.CreatedAt = time.Now()

	return u.repo.Create(ctx, form)
}

func (u *miniSahanaUsecase) GetByID(ctx context.Context, id string) (*domain.MiniSahanaForm, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *miniSahanaUsecase) Search(ctx context.Context, query string) ([]*domain.MiniSahanaForm, error) {
	if strings.TrimSpace(query) == "" {
		return []*domain.MiniSahanaForm{}, nil
	}
	return u.repo.Search(ctx, query)
}

func (u *miniSahanaUsecase) Update(ctx context.Context, id string, form *domain.MiniSahanaForm, userID string) error {
	// First fetch existing to get RefNumber and validate it exists
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate same fields as Create
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

	// Update RefNumber versioning
	// If existing is "A1", new should be "A1.1". If existing is "A1.1", new should be "A1.2"
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

	// Keep CreatedBy and CreatedAt from existing
	form.CreatedBy = existing.CreatedBy
	form.CreatedAt = existing.CreatedAt

	return u.repo.Update(ctx, id, form)
}

func (u *miniSahanaUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}
