package services

import (
	"context"
	"time"

	"go-api-practice/internal/models"
)

func (s *loanService) ListLoans(ctx context.Context, page, limit int, eligible *bool, purpose *string) (*LoanListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	if eligible != nil && !*eligible {
		return &LoanListResult{
			Applications: []LoanStatusResult{},
			Page:         page,
			TotalPages:   0,
		}, nil
	}

	applications, total, err := s.repo.List(ctx, page, limit, purpose)
	if err != nil {
		return nil, err
	}

	result := &LoanListResult{
		Page:       page,
		TotalPages: 0,
	}
	if total > 0 {
		result.TotalPages = (total + limit - 1) / limit
	}

	for i := range applications {
		application := applications[i]
		status := newLoanStatusResult(&application)
		result.Applications = append(result.Applications, status)
	}

	return result, nil
}

func newLoanStatusResult(application *models.LoanApplication) LoanStatusResult {
	return LoanStatusResult{
		ApplicationID: application.ID,
		FullName:      application.FullName,
		MonthlyIncome: application.MonthlyIncome,
		LoanAmount:    application.LoanAmount,
		LoanPurpose:   application.LoanPurpose,
		Age:           int(application.Age),
		PhoneNumber:   application.PhoneNumber,
		Email:         application.Email,
		Eligible:      true,
		Reason:        "Eligible under base rules",
		Timestamp:     application.CreatedAt.Format(time.RFC3339),
	}
}
