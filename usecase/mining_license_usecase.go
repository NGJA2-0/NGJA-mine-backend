package usecase

import (
	"context"
	"errors"
	"time"

	"my-fiber-app/domain"
)

type miningLicenseUsecase struct {
	repo domain.MiningLicenseRepository
}

// NewMiningLicenseUsecase creates a new usecase for mining license applications.
func NewMiningLicenseUsecase(repo domain.MiningLicenseRepository) domain.MiningLicenseUsecase {
	return &miningLicenseUsecase{repo: repo}
}

// Submit validates and persists a new license application.
func (u *miningLicenseUsecase) Submit(ctx context.Context, license *domain.MechanizedGemMiningLicense) (*domain.MechanizedGemMiningLicense, error) {
	// ── Required-field validation ─────────────────────────────────────────────
	if license.ApplicantName == "" {
		return nil, errors.New("applicantName is required")
	}
	if license.ApplicantAddress == "" {
		return nil, errors.New("applicantAddress is required")
	}
	if license.ApplicantPhone == "" {
		return nil, errors.New("applicantPhone is required")
	}
	if license.NIC == "" {
		return nil, errors.New("nic is required")
	}
	if license.GMLNumber == "" {
		return nil, errors.New("gmlNumber is required")
	}
	if len(license.GPSPoints) == 0 {
		return nil, errors.New("at least one gpsPoint is required")
	}
	if license.LandName == "" {
		return nil, errors.New("landName is required")
	}
	if license.LandNature == "" {
		return nil, errors.New("landNature is required")
	}
	if license.IsRatnapuraLand == "" {
		return nil, errors.New("isRatnapuraLand is required")
	}
	if license.District == "" {
		return nil, errors.New("district is required")
	}
	if license.RegionalOffice == "" {
		return nil, errors.New("regionalOffice is required")
	}
	if license.LicenseeType == "" {
		return nil, errors.New("licenseeType is required")
	}
	if license.ExistingPits == "" {
		return nil, errors.New("existingPits is required")
	}

	// ── Conditional: expense party ───────────────────────────────────────────
	if license.HasExpenseParty {
		if license.ExpenseName == "" {
			return nil, errors.New("expenseName is required when hasExpenseParty is true")
		}
		if license.ExpenseAddress == "" {
			return nil, errors.New("expenseAddress is required when hasExpenseParty is true")
		}
		if license.ExpensePhone == "" {
			return nil, errors.New("expensePhone is required when hasExpenseParty is true")
		}
	}

	// ── Conditional: Ratnapura land attachments ───────────────────────────────
	if license.IsRatnapuraLand == "yes" {
		if license.WrittenEvidenceAttachmentUrl == "" {
			return nil, errors.New("writtenEvidenceAttachmentUrl is required for Ratnapura land")
		}
		if license.AffidavitAttachmentUrl == "" {
			return nil, errors.New("affidavitAttachmentUrl is required for Ratnapura land")
		}
	}

	// ── Conditional: existing pits history ───────────────────────────────────
	if license.ExistingPits == "yes" {
		if license.PrevLicenseFirstDate == "" {
			return nil, errors.New("prevLicenseFirstDate is required when existingPits is yes")
		}
		if license.ExtensionCount == nil {
			return nil, errors.New("extensionCount is required when existingPits is yes")
		}
		if license.MinedGemValue == "" {
			return nil, errors.New("minedGemValue is required when existingPits is yes")
		}
		if license.ConditionBreach == "" {
			return nil, errors.New("conditionBreach is required when existingPits is yes")
		}
		if license.OwnershipComplaint == "" {
			return nil, errors.New("ownershipComplaint is required when existingPits is yes")
		}
	}

	// ── Conditional: breach / complaint details ───────────────────────────────
	if license.ConditionBreach == "yes" && license.ConditionBreachDetails == "" {
		return nil, errors.New("conditionBreachDetails is required when conditionBreach is yes")
	}
	if license.OwnershipComplaint == "yes" && license.ComplaintDetails == "" {
		return nil, errors.New("complaintDetails is required when ownershipComplaint is yes")
	}

	// ── Defaults ──────────────────────────────────────────────────────────────
	now := time.Now().UTC()
	license.CreatedAt = now
	license.UpdatedAt = now

	if license.Status == "" {
		license.Status = "draft"
	}

	// ── Persist ───────────────────────────────────────────────────────────────
	if err := u.repo.Create(ctx, license); err != nil {
		return nil, err
	}
	return license, nil
}

// GetByID retrieves a single license by its ID.
func (u *miningLicenseUsecase) GetByID(ctx context.Context, id string) (*domain.MechanizedGemMiningLicense, error) {
	return u.repo.GetByID(ctx, id)
}

// GetAll retrieves all license applications.
func (u *miningLicenseUsecase) GetAll(ctx context.Context) ([]domain.MechanizedGemMiningLicense, error) {
	return u.repo.GetAll(ctx)
}

// UpdateStatus updates only the status of a license application.
func (u *miningLicenseUsecase) UpdateStatus(ctx context.Context, id string, status string) error {
	allowed := map[string]bool{"draft": true, "submitted": true, "approved": true, "rejected": true}
	if !allowed[status] {
		return errors.New("invalid status value; must be one of: draft, submitted, approved, rejected")
	}
	return u.repo.UpdateStatus(ctx, id, status)
}
