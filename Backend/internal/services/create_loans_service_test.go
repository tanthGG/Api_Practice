package services

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"go-api-practice/internal/models"
)

type mockLoanRepository struct {
	saveCalled  bool
	lastSaved   *models.LoanApplication
	err         error
	getResult   *models.LoanApplication
	getErr      error
	listCalled  bool
	listPage    int
	listLimit   int
	listPurpose *string
	listResult  []models.LoanApplication
	listTotal   int
	listErr     error
}

func (m *mockLoanRepository) Save(ctx context.Context, application *models.LoanApplication) error {
	m.saveCalled = true
	m.lastSaved = application
	return m.err
}

func (m *mockLoanRepository) GetByApplicationID(ctx context.Context, id string) (*models.LoanApplication, error) {
	return m.getResult, m.getErr
}

func (m *mockLoanRepository) List(ctx context.Context, page, limit int, purpose *string) ([]models.LoanApplication, int, error) {
	m.listCalled = true
	m.listPage = page
	m.listLimit = limit
	m.listPurpose = purpose
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listResult, m.listTotal, nil
}

func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}

func TestLoanServiceApply(t *testing.T) {
	baseInput := ApplyLoanInput{
		FullName:      "Somchai Toree",
		MonthlyIncome: 20000,
		LoanAmount:    50000,
		LoanPurpose:   "home",
		Age:           30,
		PhoneNumber:   "0851234554",
		Email:         "demo@example.com",
	}

	tests := []struct {
		name         string
		modifyInput  func(inp ApplyLoanInput) ApplyLoanInput
		wantReason   string
		wantEligible bool
		expectSave   bool
		expectErr    bool
		errCheck     func(error) bool
	}{
		{
			name:         "eligible application saved",
			wantReason:   "Eligible under base rules",
			wantEligible: true,
			expectSave:   true,
			expectErr:    false,
		},
		{
			name: "income below threshold",
			modifyInput: func(inp ApplyLoanInput) ApplyLoanInput {
				inp.MonthlyIncome = 9000
				return inp
			},
			wantReason:   "Monthly income is insufficient",
			wantEligible: false,
			expectSave:   false,
			expectErr:    true,
			errCheck: func(err error) bool {
				var ineligible *IneligibleError
				return errors.As(err, &ineligible) && ineligible.Reason == "Monthly income is insufficient"
			},
		},
		{
			name: "age out of range",
			modifyInput: func(inp ApplyLoanInput) ApplyLoanInput {
				inp.Age = 65
				return inp
			},
			wantReason:   "Age not in range (must be between 20-60)",
			wantEligible: false,
			expectSave:   false,
			expectErr:    true,
			errCheck: func(err error) bool {
				var ineligible *IneligibleError
				return errors.As(err, &ineligible) && ineligible.Reason == "Age not in range (must be between 20-60)"
			},
		},
		{
			name: "business purpose rejected",
			modifyInput: func(inp ApplyLoanInput) ApplyLoanInput {
				inp.LoanPurpose = "business"
				return inp
			},
			wantReason:   "Business loans not supported",
			wantEligible: false,
			expectSave:   false,
			expectErr:    true,
			errCheck: func(err error) bool {
				var ineligible *IneligibleError
				return errors.As(err, &ineligible) && ineligible.Reason == "Business loans not supported"
			},
		},
		{
			name: "loan amount exceeds cap",
			modifyInput: func(inp ApplyLoanInput) ApplyLoanInput {
				inp.LoanAmount = inp.MonthlyIncome*12 + 1
				return inp
			},
			wantReason:   "Loan amount cannot exceed 12 months of income",
			wantEligible: false,
			expectSave:   false,
			expectErr:    true,
			errCheck: func(err error) bool {
				var ineligible *IneligibleError
				return errors.As(err, &ineligible) && ineligible.Reason == "Loan amount cannot exceed 12 months of income"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockLoanRepository{}
			service := NewLoanService(repo, newTestLogger())

			input := baseInput
			if tt.modifyInput != nil {
				input = tt.modifyInput(baseInput)
			}

			result, err := service.Apply(context.Background(), input)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errCheck != nil && !tt.errCheck(err) {
					t.Fatalf("error did not match expectation: %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatalf("expected result, got nil")
			}

			if result.Eligible != tt.wantEligible {
				t.Errorf("Eligible = %v, want %v", result.Eligible, tt.wantEligible)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}

			if tt.expectSave && !repo.saveCalled {
				t.Fatalf("expected repository Save to be called")
			}
			if !tt.expectSave && repo.saveCalled {
				t.Fatalf("expected repository Save NOT to be called")
			}

			if tt.expectSave {
				if repo.lastSaved == nil {
					t.Fatalf("expected saved application")
				}
				if repo.lastSaved.FullName != input.FullName {
					t.Errorf("saved FullName = %s, want %s", repo.lastSaved.FullName, input.FullName)
				}
			}
		})
	}

	t.Run("repository error", func(t *testing.T) {
		repo := &mockLoanRepository{err: errors.New("write failed")}
		service := NewLoanService(repo, newTestLogger())

		_, err := service.Apply(context.Background(), baseInput)
		if err == nil || !errors.Is(err, repo.err) {
			t.Fatalf("expected repo error, got %v", err)
		}
	})
}

func TestLoanServiceGetStatus(t *testing.T) {
	now := time.Now()
	application := &models.LoanApplication{
		ID:            "loan-123",
		FullName:      "Somchai Toree",
		MonthlyIncome: 20000,
		LoanAmount:    50000,
		LoanPurpose:   "home",
		Age:           30,
		PhoneNumber:   "0851234554",
		Email:         "demo@example.com",
		CreatedAt:     now,
	}

	t.Run("success", func(t *testing.T) {
		repo := &mockLoanRepository{getResult: application}
		service := NewLoanService(repo, newTestLogger())

		result, err := service.GetStatus(context.Background(), "loan-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ApplicationID != "loan-123" || !result.Eligible {
			t.Fatalf("unexpected result: %+v", result)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockLoanRepository{getErr: sql.ErrNoRows}
		service := NewLoanService(repo, newTestLogger())

		_, err := service.GetStatus(context.Background(), "missing")
		if !errors.Is(err, ErrLoanNotFound) {
			t.Fatalf("expected ErrLoanNotFound, got %v", err)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repoErr := errors.New("db down")
		repo := &mockLoanRepository{getErr: repoErr}
		service := NewLoanService(repo, newTestLogger())

		_, err := service.GetStatus(context.Background(), "loan-123")
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected repo error, got %v", err)
		}
	})
}

func TestLoanServiceListLoans(t *testing.T) {
	app := models.LoanApplication{
		ID:            "loan-123",
		FullName:      "Somchai",
		MonthlyIncome: 20000,
		LoanAmount:    50000,
		LoanPurpose:   "home",
		Age:           30,
		PhoneNumber:   "0851234554",
		Email:         "demo@example.com",
		CreatedAt:     time.Now(),
	}

	t.Run("eligible list returns data", func(t *testing.T) {
		repo := &mockLoanRepository{
			listResult: []models.LoanApplication{app},
			listTotal:  1,
		}
		service := NewLoanService(repo, newTestLogger())

		result, err := service.ListLoans(context.Background(), 1, 10, nil, nil)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(result.Applications) != 1 || result.TotalPages != 1 {
			t.Fatalf("unexpected result: %+v", result)
		}
		if !repo.listCalled || repo.listPage != 1 || repo.listLimit != 10 {
			t.Fatalf("repository list not called with expected params")
		}
	})

	t.Run("eligible=false returns empty without repo call", func(t *testing.T) {
		repo := &mockLoanRepository{}
		service := NewLoanService(repo, newTestLogger())
		flag := false

		result, err := service.ListLoans(context.Background(), 1, 10, &flag, nil)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(result.Applications) != 0 || result.TotalPages != 0 {
			t.Fatalf("expected empty result, got %+v", result)
		}
		if repo.listCalled {
			t.Fatalf("repo should not be called when eligible=false")
		}
	})

	t.Run("repo error bubbles up", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockLoanRepository{listErr: repoErr}
		service := NewLoanService(repo, newTestLogger())

		_, err := service.ListLoans(context.Background(), 1, 10, nil, nil)
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected repo error, got %v", err)
		}
	})
}
