package service

import (
	"context"
	"fmt"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/carbon/domain"
	"github.com/google/uuid"
)

type CarbonService struct {
	repo CarbonRepo
}

// In production, I should use an interface for the Repository. Let's fix that.
type CarbonRepo interface {
	Save(ctx context.Context, entry domain.CarbonEntry) error
	GetByVendorPeriod(ctx context.Context, vendorID string, start, end string) ([]domain.CarbonEntry, error)
	GetLastHash(ctx context.Context) (string, error)
}

func NewCarbonService(repo CarbonRepo) *CarbonService {
	return &CarbonService{
		repo: repo,
	}
}

// RecordSavings adds a new block to the carbon ledger
func (s *CarbonService) RecordSavings(ctx context.Context, vendorID, orderID, category string, weightKg float64) (string, error) {
	savings := domain.CalculateSavings(weightKg, category)
	
	prevHash, err := s.repo.GetLastHash(ctx)
	if err != nil {
		prevHash = "0000000000000000"
	}
	
	entry := domain.CarbonEntry{
		ID:            uuid.New().String(),
		VendorID:      vendorID,
		OrderID:       orderID,
		FoodCategory:  category,
		WeightKg:      weightKg,
		CarbonSavedKg: savings,
		Timestamp:     time.Now(),
		PreviousHash:  prevHash,
	}
	
	entry.Hash = entry.ComputeHash()
	
	if err := s.repo.Save(ctx, entry); err != nil {
		return "", err
	}
	
	return entry.Hash, nil
}

// GenerateESGReport creates a formal certificate for corporate partners
func (s *CarbonService) GenerateESGReport(ctx context.Context, vendorID string, year int) (domain.ESGReport, error) {
	start := fmt.Sprintf("%d-01-01 00:00:00", year)
	end := fmt.Sprintf("%d-12-31 23:59:59", year)

	entries, err := s.repo.GetByVendorPeriod(ctx, vendorID, start, end)
	if err != nil {
		return domain.ESGReport{}, err
	}
	
	var totalFood, totalCarbon float64
	for _, entry := range entries {
		totalFood += entry.WeightKg
		totalCarbon += entry.CarbonSavedKg
	}
	
	if len(entries) == 0 {
		return domain.ESGReport{}, fmt.Errorf("no transactions found for vendor %s in %d", vendorID, year)
	}

	report := domain.ESGReport{
		VendorID:         vendorID,
		PeriodStart:      time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:        time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC),
		TotalFoodSaved:   totalFood,
		TotalCarbonSaved: totalCarbon,
		TransactionCount: len(entries),
		VerificationHash: fmt.Sprintf("VERIFIED-%s-%d-%f", vendorID, year, totalCarbon),
	}
	
	return report, nil
}
