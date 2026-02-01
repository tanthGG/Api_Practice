package repositories

import (
	"context"
	"database/sql"

	"go-api-practice/internal/models"
)

// LoanRepository defines persistence contract for loan applications.
type LoanRepository interface {
	Save(ctx context.Context, application *models.LoanApplication) error
}

type loanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) Save(ctx context.Context, application *models.LoanApplication) error {
	query := `
		INSERT INTO loan_applications (
			id, full_name, monthly_income, loan_amount, loan_purpose,
			age, phone_number, email, created_at, updated_at, unix_ts
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		application.ID,
		application.FullName,
		application.MonthlyIncome,
		application.LoanAmount,
		application.LoanPurpose,
		application.Age,
		application.PhoneNumber,
		application.Email,
		application.CreatedAt,
		application.UpdatedAt,
		application.UnixTS,
	)
	return err
}
