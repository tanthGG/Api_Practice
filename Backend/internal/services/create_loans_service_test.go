package services

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/sirupsen/logrus"

	"go-api-practice/internal/models"
)

type mockLoanRepository struct {
	saveCalled bool
	lastSaved  *models.LoanApplication
	err        error
}

func (m *mockLoanRepository) Save(ctx context.Context, application *models.LoanApplication) error {
	m.saveCalled = true
	m.lastSaved = application
	return m.err
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
