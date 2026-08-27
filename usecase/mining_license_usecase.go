package usecase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

	// ── Generate Reference Number ─────────────────────────────────────────────
	if license.ReferenceNumber == "" {
		seq, err := u.repo.GetNextBaseReferenceNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate reference number: %w", err)
		}
		license.ReferenceNumber = fmt.Sprintf("REF_%d", seq)
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

// GetByTIN retrieves all license applications for a specific TIN number.
func (u *miningLicenseUsecase) GetByTIN(ctx context.Context, tin string, page int, limit int) (*domain.PaginatedMiningLicenses, error) {
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
func (u *miningLicenseUsecase) GetForMap(ctx context.Context, filters domain.MapFilters) ([]domain.MechanizedGemMiningLicense, error) {
	return u.repo.GetForMap(ctx, filters)
}

// UpdateStatus updates only the status of a license application.
func (u *miningLicenseUsecase) UpdateStatus(ctx context.Context, id string, status string) error {
	allowed := map[string]bool{"draft": true, "submitted": true, "approved": true, "rejected": true}
	if !allowed[status] {
		return errors.New("invalid status value; must be one of: draft, submitted, approved, rejected")
	}
	return u.repo.UpdateStatus(ctx, id, status)
}

// Edit creates a new versioned copy of an existing mining license application.
// The new document is stored with an incremented version suffix on the reference
// number, e.g. REF_2 → REF_2.1, REF_2.1 → REF_2.2.
// The suffix is always based on the highest version already present in the DB,
// so interleaved edits from both the edit and extend endpoints are handled safely.
func (u *miningLicenseUsecase) Edit(ctx context.Context, id string, updatedLicense *domain.MechanizedGemMiningLicense) (*domain.MechanizedGemMiningLicense, error) {
	// 1. Fetch the existing license
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing license: %w", err)
	}

	// 2. Extract the base reference number (strip any existing suffix)
	// e.g. "REF_2.1" -> "REF_2",  "REF_2" -> "REF_2"
	baseRef := existing.ReferenceNumber
	if baseRef == "" {
		return nil, errors.New("existing license does not have a reference number")
	}
	if parts := strings.Split(baseRef, "."); len(parts) >= 2 {
		baseRef = parts[0]
	}

	// 3. Find the highest version suffix currently in the collection for this base ref
	maxVersion, err := u.repo.GetMaxVersionByBaseRef(ctx, baseRef)
	if err != nil {
		return nil, fmt.Errorf("failed to determine next version: %w", err)
	}
	updatedLicense.ReferenceNumber = fmt.Sprintf("%s.%d", baseRef, maxVersion+1)

	// 4. Keep the original status
	updatedLicense.Status = existing.Status

	// 5. Set timestamps
	now := time.Now().UTC()
	updatedLicense.CreatedAt = now
	updatedLicense.UpdatedAt = now

	// 6. Persist as a brand new document (no ID → repository assigns one)

		if err := u.repo.Create(ctx, updatedLicense); err != nil {
		return nil, err
	}

	return updatedLicense, nil
}

// GetByReferenceNumber returns every edition (version) stored for a given
// reference number. The user may pass either a base ref ("REF_4") or a
// versioned one ("REF_4.2") — either way it's normalized down to the base
// ref before querying, so all related editions are returned.
func (u *miningLicenseUsecase) GetByReferenceNumber(ctx context.Context, refNumber string, page int, limit int) (*domain.PaginatedMiningLicenseSummaries, error) {
	if refNumber == "" {
		return nil, errors.New("reference number is required")
	}

	baseRef := refNumber
	if parts := strings.Split(baseRef, "."); len(parts) >= 2 {
		baseRef = parts[0]
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	return u.repo.GetByBaseReferenceNumber(ctx, baseRef, page, limit)
}

// CompareVersions compares a specific license version against its immediate predecessor.
func (u *miningLicenseUsecase) CompareVersions(ctx context.Context, id string) (*domain.CompareResult, error) {
	// 1. Fetch current raw doc
	currDoc, err := u.repo.GetRawByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch license: %w", err)
	}

	// 2. Determine referenceNumber
	refVal, ok := currDoc["referenceNumber"].(string)
	if !ok || refVal == "" {
		return nil, errors.New("record does not have a valid reference number")
	}

	// 3. Determine previous referenceNumber
	parts := strings.Split(refVal, ".")
	if len(parts) == 1 {
		return nil, errors.New("No previous version exists to compare")
	}

	baseRef := parts[0]
	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.New("invalid version format in reference number")
	}

	var prevRef string
	if version == 1 {
		prevRef = baseRef
	} else {
		prevRef = fmt.Sprintf("%s.%d", baseRef, version-1)
	}

	// 4. Fetch previous raw doc
	prevDoc, err := u.repo.GetRawByReferenceNumber(ctx, prevRef)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch previous version (%s): %w", prevRef, err)
	}

	// 5. Compare fields
	changes := make(map[string]domain.FieldChange)
	isExtended := false

	// Compare curr against prev
	for k, vCurr := range currDoc {
		if k == "_id" || k == "createdAt" || k == "updatedAt" {
			continue
		}
		vPrev, exists := prevDoc[k]
		if !exists || !reflect.DeepEqual(vCurr, vPrev) {
			changes[k] = domain.FieldChange{Old: vPrev, New: vCurr}
			if k == "privateSaleValue" || k == "auctionSaleValue" {
				isExtended = true
			}
		}
	}

	// Check fields removed in curr
	for k, vPrev := range prevDoc {
		if k == "_id" || k == "createdAt" || k == "updatedAt" {
			continue
		}
		if _, exists := currDoc[k]; !exists {
			changes[k] = domain.FieldChange{Old: vPrev, New: nil}
			if k == "privateSaleValue" || k == "auctionSaleValue" {
				isExtended = true
			}
		}
	}

	result := &domain.CompareResult{
		Changes: changes,
	}
	if isExtended {
		result.Message = "The record was extended"
	} else {
		result.Message = "Comparison successful"
	}

	return result, nil
}
func (u *miningLicenseUsecase) GetLatestByReferenceNumber(ctx context.Context, baseRef string) (*domain.LatestMiningLicenseInfo, error) {
	if baseRef == "" {
		return nil, errors.New("reference number is required")
	}
	
	if parts := strings.Split(baseRef, "."); len(parts) >= 2 {
		baseRef = parts[0]
	}

	return u.repo.GetLatestByReferenceNumber(ctx, baseRef)
}
func (u *miningLicenseUsecase) GetAllLatest(ctx context.Context) ([]domain.LatestMiningLicenseInfo, error) {
	return u.repo.GetAllLatest(ctx)
}

// dsDivisionsByDistrict is the hardcoded canonical list of DS divisions
// per district. It's used only to sanity-check which regionalOffice values
// are legitimate for a given district — it is NEVER returned directly as
// dropdown data. Dropdown options always come from what actually exists
// in the DB (see GetLatestRegionalOffices below).
var dsDivisionsByDistrict = map[string][]string{
	"කොළඹ":         {"කොළඹ", "තිඹිරිගස්යාය", "කොලොන්නාව", "කඩුවෙල", "කැස්බෑව", "මහරගම", "හෝමාගම", "ශ්‍රී ජයවර්ධනපුර කෝට්ටේ", "දෙහිවල-ගල්කිස්ස", "මොරටුව", "පාදුක්ක", "සීතාවක", "රත්මලාන"},
	"ගම්පහ":        {"ගම්පහ", "මීගමුව", "ජා-ඇල", "වත්තල", "කටාන", "දිවුලපිටිය", "මිරිගම", "මිනුවන්ගොඩ", "අත්තනගල්ල", "දොම්පේ", "බියගම", "කැලණිය", "මහර"},
	"කළුතර":        {"කළුතර", "බේරුවල", "පානදුර", "බණ්ඩාරගම", "හෝරණ", "බුලත්සිංහල", "මිල්ලනිය", "පැලින්දනුවර", "අගලවත්ත", "මතුගම", "වලල්ලාවිට", "ඉංගිරිය", "දොඩංගොඩ", "මදුරාවෙල"},
	"මහනුවර":       {"මහනුවර හතරගම් කෝරළය", "ගඟවට කෝරළය", "පාතදුම්බර", "උඩුනුවර", "යටිනුවර", "හාරිස්පත්තුව", "උඩපළාත", "අකුරණ", "දොළුව", "පන්විල", "මිනිපේ", "මැදදුම්බර", "හතරලියද්ද", "දෙල්තොට", "කුණ්ඩසාලේ", "තුම්පනේ", "පතහේවහෙට", "ගඟ ඉහළ කෝරළය", "පූජාපිටිය", "වත්තේගම"},
	"මාතලේ":        {"මාතලේ", "රත්තොට", "උකුවෙල", "පල්ලේපොල", "යටවත්ත", "නාඑල", "ගලේවෙල", "දඹුල්ල", "විල්ගමුව", "අඹන්ගඟ කෝරළය", "ලග්ගල-පල්ලේගම"},
	"නුවරඑළිය":     {"නුවරඑළිය", "හඟුරන්කෙත", "වලපනේ", "කොත්මලේ", "අඹගමුව"},
	"ගාල්ල":        {"ගාල්ල හතරගම් කෝරළය", "බෝප-පොද්දල", "අක්මීමන", "යක්කලමුල්ල", "බද්දේගම", "එල්පිටිය", "නාගොඩ", "නෙළුව", "තවලම", "බෙන්තොට", "බලපිටිය", "අම්බලන්ගොඩ", "කරන්දෙණිය", "හබරාදුව", "ඉමදුව", "වැලිවිටිය-දිවිතුර", "හික්කඩුව", "ගොනාපිනුවල", "නියාගම"},
	"මාතර":         {"මාතර හතරගම් කෝරළය", "දෙවිනුවර", "වැලිගම", "අකුරැස්ස", "කඹුරුපිටිය", "කොටපොල", "පස්ගොඩ", "පිටබැද්දර", "මාලිම්බඩ", "මුලටියන", "අතුරලිය", "හක්මන", "තිහගොඩ", "දික්වැල්ල", "කිරින්ද පුහුල්වැල්ල", "වැලිපිටිය"},
	"හම්බන්තොට":    {"හම්බන්තොට", "තංගල්ල", "තිස්සමහාරාමය", "අම්බලන්තොට", "බෙලිඅත්ත", "වලස්මුල්ල", "වීරකැටිය", "අඟුණකොලපැලැස්ස", "සූරියවැව", "ලුණුගම්වෙහෙර", "කටුවන", "තණමල්විල"},
	"යාපනය":        {"යාපනය", "නල්ලූර්", "කොපායි", "උඩුවිල්", "තෙල්ලිප්පලෙයි", "සන්දිලිප්පායි", "චාවකච්චේරිය", "පොයින්ට් පේද්‍රෝ", "කාරයිනගර්", "කයිට්ස්", "වේලනෛ", "ඩෙල්ෆ්ට්", "උතුරු වඩමාරච්චිය", "නැගෙනහිර වඩමාරච්චිය", "තෙන්මරච්චිය"},
	"කිලිනොච්චිය":  {"කිලිනොච්චිය", "කරච්චි", "පච්චිලෛපල්ලි", "පූනාකරි"},
	"මන්නාරම":      {"මන්නාරම නගරය", "බටහිර මන්තෙයි", "නානට්ටාන්", "මඩු", "මුසලි"},
	"වවුනියාව":     {"වවුනියාව", "උතුරු වවුනියාව", "දකුණු වවුනියාව", "වෙන්ගලච්චෙට්ටිකුලම්"},
	"මුලතිව්":      {"නැගෙනහිර මන්තෙයි", "මරිතයිම්පට්ටු", "ඔද්දුසුඩාන්", "පුතුක්කුඩියිරුප්පු", "තුනුක්කායි"},
	"මඩකලපුව":      {"උතුරු මන්මුනෛ", "මන්මුනෛ පත්තු", "දකුණු මන්මුනෛ සහ එරුවිල් පත්තු", "බටහිර මන්මුනෛ", "නිරිතදිග මන්මුනෛ", "කෝරළෛ පත්තු", "උතුරු කෝරළෛ පත්තු", "දකුණු කෝරළෛ පත්තු", "බටහිර කෝරළෛ පත්තු", "මධ්‍යම කෝරළෛ පත්තු", "එරාවුර් පත්තු", "එරාවුර් නගරය", "පොරතිවු පත්තු", "කත්තන්කුඩි"},
	"අම්පාර":       {"අම්පාර", "උහන", "දමන", "මහඔය", "පඩියතලාව", "දෙහිආටකන්දිය", "අද්දාලච්චේන", "අක්කරෙයිපත්තුව", "සයින්දමරුදු", "නින්තාවූර්", "කල්මුණේ", "කරෛතිවු", "සම්මන්තුරේ", "අලෙයාඩිවෙම්බු", "තිරුක්කෝවිල්", "පොත්තුවිල්", "ලාහුගල", "නාවිතන්වෙලි"},
	"ත්‍රිකුණාමලය": {"ත්‍රිකුණාමලය නගරය සහ ග්‍රාවට්ස්", "කින්නියා", "මුතූර්", "කුච්චවේලි", "ගෝමරන්කඩවල", "මොරවැව", "තඹලගමුව", "කන්තලේ", "සේරුවිල", "පදවි ශ්‍රී පුර", "වේරුගල්"},
	"කුරුණෑගල":     {"කුරුණෑගල", "මල්ලවපිටිය", "මාවතගම", "පොල්ගහවෙල", "වාරියපොල", "පන්නල", "රිදීගම", "ඉබ්බාගමුව", "අලව්ව", "බිංගිරිය", "නැගෙනහිර කුලියාපිටිය", "බටහිර කුලියාපිටිය", "බමුණකොටුව", "කටුපොත", "නැගෙනහිර පඬුවස්නුවර", "බටහිර පඬුවස්නුවර", "නිකවැරටිය", "මහව", "ගල්ගමුව", "ගනේවත්ත", "ගිරිබාව", "එහෙටුවෙව", "පොල්පිතිගම", "රස්නායකපුර", "වේරඹුගෙදර", "කොටවෙහෙර", "මාස්පොත", "නාරම්මල", "කොබෙයිගනේ", "අඹන්පොල"},
	"පුත්තලම":      {"පුත්තලම", "කල්පිටිය", "වනාතවිල්ලුව", "කරුවලගස්වැව", "අනමඩුව", "නාත්තණ්ඩිය", "මුන්දලම", "චිලාව", "අරච්චිකට්ටුව", "මාදම්පේ", "වෙන්නප්පුව", "මහව්ව"},
	"අනුරාධපුරය":   {"නුවරගම් පළාත නැගෙනහිර", "නුවරගම් පළාත මධ්‍යම", "කැකිරාව", "පලාගල", "තලාව", "මිහින්තලේ", "රඹෑව", "ගලෙන්බිඳුනුවැව", "කහටගස්දිගිලිය", "හොරොව්පොතාන", "මහාවිලච්චිය", "මැදවච්චිය", "පදවිය", "ගල්නෑව", "ඉපලෝගම", "නාච්චාදුව", "තිරප්පනේ", "කැබිතිගොල්ලෑව", "රාජාංගණය", "එප්පාවල", "නොච්චියාගම", "පලුගස්වැව"},
	"පොළොන්නරුව":   {"තාමන්කඩුව", "හිඟුරක්ගොඩ", "මැදිරිගිරිය", "ලංකාපුර", "ඇලහැර", "දිඹුලාගල", "වැලිකන්ද"},
	"බදුල්ල":       {"බදුල්ල", "බණ්ඩාරවෙල", "හපුතලේ", "ඇල්ල", "ලුණුගල", "මහියංගනය", "මීගහකිවුල", "පස්සර", "රිදීමාලියද්ද", "සොරණාතොට", "උව-පරණගම", "වැලිමඩ", "කන්දකැටිය", "හල්දුම්මුල්ල", "හාලි-ඇල"},
	"මොණරාගල":      {"මොණරාගල", "වැල්ලවාය", "බුත්තල", "කතරගම", "බිබිල", "මැදගම", "මඩුල්ල", "සියඹලාණ්ඩුව", "බදල්කුඹුර", "සෙවනගල", "තණමල්විල"},
	"රත්නපුර":      {"රත්නපුර", "බලංගොඩ", "එහෙළියගොඩ", "කලවාන", "කුරුවිට", "නිවිතිගල", "පැල්මඩුල්ල", "කොලොන්න", "කහවත්ත", "එලපාත", "අයගම", "ගොඩකවෙල", "ඉඹුල්පේ", "ඕපනායක", "වැලිගෙපොල", "කිරිඇල්ල"},
	"කෑගල්ල":       {"කෑගල්ල", "මාවනැල්ල", "රඹුක්කන", "වරකාපොල", "රුවන්වැල්ල", "යටියන්තොට", "දෙරණියගල", "ගාලිගමුව", "බුලත්කොහුපිටිය", "දෙහිඕවිට", "අරණායක"},
}

// GetLatestDistricts returns the distinct district values found among the
// latest edition of every application.
func (u *miningLicenseUsecase) GetLatestDistricts(ctx context.Context) ([]string, error) {
	docs, err := u.repo.GetAllLatestFull(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	districts := make([]string, 0)
	for _, doc := range docs {
		if doc.District == "" || seen[doc.District] {
			continue
		}
		seen[doc.District] = true
		districts = append(districts, doc.District)
	}
	sort.Strings(districts)
	return districts, nil
}

// GetLatestRegionalOffices returns the distinct regionalOffice values found
// among the latest edition of applications belonging to the given district.
// Values are additionally cross-checked against dsDivisionsByDistrict so a
// stray/typo'd regionalOffice on a record can never leak into the dropdown.
func (u *miningLicenseUsecase) GetLatestRegionalOffices(ctx context.Context, district string) ([]string, error) {
	if district == "" {
		return nil, errors.New("district is required")
	}

	validOffices := make(map[string]bool)
	for _, office := range dsDivisionsByDistrict[district] {
		validOffices[office] = true
	}

	docs, err := u.repo.GetAllLatestFull(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	offices := make([]string, 0)
	for _, doc := range docs {
		if doc.District != district || doc.RegionalOffice == "" {
			continue
		}
		if !validOffices[doc.RegionalOffice] || seen[doc.RegionalOffice] {
			continue
		}
		seen[doc.RegionalOffice] = true
		offices = append(offices, doc.RegionalOffice)
	}
	sort.Strings(offices)
	return offices, nil
}

// GetLatestFiltered returns the latest edition of every application matching
// both district and regionalOffice, projected to the slim result set.
func (u *miningLicenseUsecase) GetLatestFiltered(ctx context.Context, district string, regionalOffice string) ([]domain.FilteredLicenseSummary, error) {
	if district == "" || regionalOffice == "" {
		return nil, errors.New("district and regionalOffice are required")
	}

	docs, err := u.repo.GetAllLatestFull(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]domain.FilteredLicenseSummary, 0)
	for _, doc := range docs {
		if doc.District != district || doc.RegionalOffice != regionalOffice {
			continue
		}
		results = append(results, domain.FilteredLicenseSummary{
			ID:             doc.ID,
			ApplicantName:  doc.ApplicantName,
			ApplicantPhone: doc.ApplicantPhone,
			TIN:            doc.TIN,
			GMLNumber:      doc.GMLNumber,
			GPSPoints:      doc.GPSPoints,
			CreatedBy:      doc.CreatedBy,
			CreatedAt:      doc.CreatedAt,
			UpdatedAt:      doc.UpdatedAt,
			Status:         doc.Status,
		})
	}
	return results, nil
}
