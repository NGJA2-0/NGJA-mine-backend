package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"my-fiber-app/domain"
)

type miniSahanaUsecase struct {
	repo domain.MiniSahanaRepository
}

func NewMiniSahanaUsecase(repo domain.MiniSahanaRepository) domain.MiniSahanaUsecase {
	return &miniSahanaUsecase{
		repo: repo,
	}
}

func (u *miniSahanaUsecase) Create(ctx context.Context, form *domain.MiniSahanaForm) error {
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

	form.CreatedAt = time.Now()

	return u.repo.Create(ctx, form)
}
