package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"my-fiber-app/domain"
)

type extendMiningLicenseUsecase struct {
	repo domain.ExtendMiningLicenseRepository
}

// NewExtendMiningLicenseUsecase creates a new usecase for extend mining license applications.
func NewExtendMiningLicenseUsecase(repo domain.ExtendMiningLicenseRepository) domain.ExtendMiningLicenseUsecase {
	return &extendMiningLicenseUsecase{repo: repo}
}

// nextVersionedRef derives the base ref from an existing referenceNumber (stripping any suffix),
// queries the DB for the highest version already stored, and returns the next version string.
// e.g. "REF_2"   -> finds max suffix -> returns "REF_2.1"
//
//	"REF_2.1" -> finds max suffix -> returns "REF_2.2"
func (u *extendMiningLicenseUsecase) nextVersionedRef(ctx context.Context, existingRef string) (string, error) {
	baseRef := existingRef
	if parts := strings.Split(existingRef, "."); len(parts) >= 2 {
		baseRef = parts[0]
	}

	maxVersion, err := u.repo.GetMaxVersionByBaseRef(ctx, baseRef)
	if err != nil {
		return "", fmt.Errorf("failed to determine next version: %w", err)
	}
	return fmt.Sprintf("%s.%d", baseRef, maxVersion+1), nil
}

// Submit validates and persists a new license extend application.
// If the incoming payload already has a referenceNumber (i.e., the form was pre-filled
// from an existing record), the new document is saved to the SAME collection with an
// incremented version suffix (REF_2 → REF_2.1, REF_2.1 → REF_2.2, etc.).
// If there is no referenceNumber, a brand-new base reference is generated.
func (u *extendMiningLicenseUsecase) Submit(ctx context.Context, license *domain.ExtendMiningLicense) (*domain.ExtendMiningLicense, error) {
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

	// ── Defaults ──────────────────────────────────────────────────────────────
	now := time.Now().UTC()
	license.CreatedAt = now
	license.UpdatedAt = now

	if license.Status == "" {
		license.Status = "draft"
	}

	// ── Reference Number ──────────────────────────────────────────────────────
	// If the payload already carries a referenceNumber it means this is an
	// extension of an existing record — compute the next version suffix.
	// Otherwise, generate a fresh base reference number.
	if license.ReferenceNumber != "" {
		nextRef, err := u.nextVersionedRef(ctx, license.ReferenceNumber)
		if err != nil {
			return nil, err
		}
		license.ReferenceNumber = nextRef
	} else {
		seq, err := u.repo.GetNextBaseReferenceNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate reference number: %w", err)
		}
		license.ReferenceNumber = fmt.Sprintf("REF_%d", seq)
	}

	// ── Clear any ID the frontend may have echoed back so a fresh one is assigned ──
	license.ID = [12]byte{}

	// ── Persist ───────────────────────────────────────────────────────────────
	if err := u.repo.Create(ctx, license); err != nil {
		return nil, err
	}
	return license, nil
}

// GetByID retrieves a single license by its ID.
func (u *extendMiningLicenseUsecase) GetByID(ctx context.Context, id string) (*domain.ExtendMiningLicense, error) {
	return u.repo.GetByID(ctx, id)
}

// GetAll retrieves all license applications.
func (u *extendMiningLicenseUsecase) GetAll(ctx context.Context) ([]domain.ExtendMiningLicense, error) {
	return u.repo.GetAll(ctx)
}

// GetByTIN retrieves all license applications for a specific TIN number.
func (u *extendMiningLicenseUsecase) GetByTIN(ctx context.Context, tin string, page int, limit int) (*domain.PaginatedExtendMiningLicenses, error) {
	if tin == "" {
		return nil, errors.New("TIN number is required")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return u.repo.GetByTIN(ctx, tin, page, limit)
}

// GetForMap retrieves license applications for the map view, filtered by
// any combination of district, village, TIN, NIC, GML number, or land name.
func (u *extendMiningLicenseUsecase) GetForMap(ctx context.Context, filters domain.MapFilters) ([]domain.ExtendMiningLicense, error) {
	return u.repo.GetForMap(ctx, filters)
}

// UpdateStatus updates only the status of a license application.
func (u *extendMiningLicenseUsecase) UpdateStatus(ctx context.Context, id string, status string) error {
	allowed := map[string]bool{"draft": true, "submitted": true, "approved": true, "rejected": true}
	if !allowed[status] {
		return errors.New("invalid status value; must be one of: draft, submitted, approved, rejected")
	}
	return u.repo.UpdateStatus(ctx, id, status)
}

// Edit creates a new versioned copy of an existing extend license application.
// The suffix is always based on the highest version already present in the DB,
// so interleaved edits from both the edit and extend endpoints are handled safely.
func (u *extendMiningLicenseUsecase) Edit(ctx context.Context, id string, updatedLicense *domain.ExtendMiningLicense) (*domain.ExtendMiningLicense, error) {
	// 1. Fetch the existing license
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing license: %w", err)
	}

	if existing.ReferenceNumber == "" {
		return nil, errors.New("existing license does not have a reference number")
	}

	// 2. Compute next version
	nextRef, err := u.nextVersionedRef(ctx, existing.ReferenceNumber)
	if err != nil {
		return nil, err
	}
	updatedLicense.ReferenceNumber = nextRef

	// 3. Keep the original status
	updatedLicense.Status = existing.Status

	// 4. Set timestamps
	now := time.Now().UTC()
	updatedLicense.CreatedAt = now
	updatedLicense.UpdatedAt = now

	// 5. Persist as a brand new document (no ID → repository assigns one)
	updatedLicense.ID = [12]byte{}

	if err := u.repo.Create(ctx, updatedLicense); err != nil {
		return nil, err
	}

	return updatedLicense, nil
}
