package repositories

import (
	"context"
	"database/sql"

	"go-api-practice/internal/models"
)

// LoanRepository defines persistence contract for loan applications.
type LoanRepository interface {
	Save(ctx context.Context, application *models.LoanApplication) error
	GetByApplicationID(ctx context.Context, id string) (*models.LoanApplication, error)
	List(ctx context.Context, page, limit int, purpose *string) ([]models.LoanApplication, int, error)
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

func (r *loanRepository) GetByApplicationID(ctx context.Context, id string) (*models.LoanApplication, error) {
	query := `
		SELECT id, full_name, monthly_income, loan_amount, loan_purpose,
			age, phone_number, email, created_at, updated_at, unix_ts
		FROM loan_applications
		WHERE id = ?
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, id)
	var application models.LoanApplication
	if err := row.Scan(
		&application.ID,
		&application.FullName,
		&application.MonthlyIncome,
		&application.LoanAmount,
		&application.LoanPurpose,
		&application.Age,
		&application.PhoneNumber,
		&application.Email,
		&application.CreatedAt,
		&application.UpdatedAt,
		&application.UnixTS,
	); err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *loanRepository) List(ctx context.Context, page, limit int, purpose *string) ([]models.LoanApplication, int, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	whereClause := ""
	var args []any
	if purpose != nil && *purpose != "" {
		whereClause = "WHERE LOWER(loan_purpose) = ?"
		args = append(args, *purpose)
	}

	countQuery := "SELECT COUNT(*) FROM loan_applications " + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []models.LoanApplication{}, 0, nil
	}

	query := `
		SELECT id, full_name, monthly_income, loan_amount, loan_purpose,
			age, phone_number, email, created_at, updated_at, unix_ts
		FROM loan_applications
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var applications []models.LoanApplication
	for rows.Next() {
		var application models.LoanApplication
		if err := rows.Scan(
			&application.ID,
			&application.FullName,
			&application.MonthlyIncome,
			&application.LoanAmount,
			&application.LoanPurpose,
			&application.Age,
			&application.PhoneNumber,
			&application.Email,
			&application.CreatedAt,
			&application.UpdatedAt,
			&application.UnixTS,
		); err != nil {
			return nil, 0, err
		}
		applications = append(applications, application)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return applications, total, nil
}
