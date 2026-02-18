package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"go-api-practice/internal/models"
	"go-api-practice/internal/repositories"
)

type LoanService interface {
	Apply(ctx context.Context, input ApplyLoanInput) (*ApplyLoanResult, error)
	GetStatus(ctx context.Context, applicationID string) (*LoanStatusResult, error)
	ListLoans(ctx context.Context, page, limit int, eligible *bool, purpose *string) (*LoanListResult, error)
}

type ApplyLoanInput struct {
	FullName      string
	MonthlyIncome float64
	LoanAmount    float64
	LoanPurpose   string
	Age           int
	PhoneNumber   string
	Email         string
}

type ApplyLoanResult struct {
	ApplicationID string
	Eligible      bool
	Reason        string
	Timestamp     string
}

type LoanListResult struct {
	Applications []LoanStatusResult
	Page         int
	TotalPages   int
}

var ErrIneligible = errors.New("loan application not eligible")

type loanService struct {
	repo   repositories.LoanRepository
	logger *logrus.Logger
}

func NewLoanService(repo repositories.LoanRepository, logger *logrus.Logger) LoanService {
	return &loanService{
		repo:   repo,
		logger: logger,
	}
}

func (s *loanService) Apply(ctx context.Context, input ApplyLoanInput) (*ApplyLoanResult, error) {
	now := time.Now().UTC()

	eligible, reason := evaluateEligibility(input)
	result := &ApplyLoanResult{
		Eligible:  eligible,
		Reason:    reason,
		Timestamp: now.Format(time.RFC3339),
	}
	if !eligible {
		return result, &IneligibleError{Reason: reason}
	}

	application := &models.LoanApplication{
		ID:            uuid.NewString(),
		FullName:      input.FullName,
		MonthlyIncome: input.MonthlyIncome,
		LoanAmount:    input.LoanAmount,
		LoanPurpose:   input.LoanPurpose,
		Age:           uint(input.Age),
		PhoneNumber:   input.PhoneNumber,
		Email:         input.Email,
		CreatedAt:     now,
		UpdatedAt:     now,
		UnixTS:        now.UnixMilli(),
	}

	if err := s.repo.Save(ctx, application); err != nil {
		return nil, err
	}

	s.logger.WithField("application_id", application.ID).Info("loan application saved")

	result.ApplicationID = application.ID
	return result, nil
}

func evaluateEligibility(input ApplyLoanInput) (bool, string) {
	if input.MonthlyIncome < 10000 {
		return false, "Monthly income is insufficient"
	}

	if input.Age < 20 || input.Age > 60 {
		return false, "Age not in range (must be between 20-60)"
	}

	if strings.EqualFold(input.LoanPurpose, "business") {
		return false, "Business loans not supported"
	}

	maxLoan := input.MonthlyIncome * 12
	if input.LoanAmount > maxLoan {
		return false, "Loan amount cannot exceed 12 months of income"
	}

	return true, "Eligible under base rules"
}

type IneligibleError struct {
	Reason string
}

func (e *IneligibleError) Error() string {
	if e.Reason == "" {
		return ErrIneligible.Error()
	}
	return e.Reason
}
