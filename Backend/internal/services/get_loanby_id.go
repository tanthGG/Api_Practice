package services

import (
	"context"
	"database/sql"
	"errors"
)

type LoanStatusResult struct {
	ApplicationID string
	FullName      string
	MonthlyIncome float64
	LoanAmount    float64
	LoanPurpose   string
	Age           int
	PhoneNumber   string
	Email         string
	Eligible      bool
	Reason        string
	Timestamp     string
}

var ErrLoanNotFound = errors.New("loan application not found")

func (s *loanService) GetStatus(ctx context.Context, applicationID string) (*LoanStatusResult, error) {
	if applicationID == "" {
		return nil, errors.New("applicationID is required")
	}

	application, err := s.repo.GetByApplicationID(ctx, applicationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLoanNotFound
		}
		return nil, err
	}

	result := newLoanStatusResult(application)
	return &result, nil
}
