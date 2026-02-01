package models

import "time"

type LoanApplication struct {
	ID string `gorm:"type:char(36);primaryKey"` //UUID

	FullName      string  `gorm:"column:full_name;type:varchar(255);not null"`
	MonthlyIncome float64 `gorm:"column:monthly_income;type:decimal(12,2);not null"`
	LoanAmount    float64 `gorm:"column:loan_amount;type:decimal(12,2);not null"`
	LoanPurpose   string  `gorm:"column:loan_purpose;type:varchar(50);not null"`
	Age           uint    `gorm:"column:age;not null"`
	PhoneNumber   string  `gorm:"column:phone_number;type:varchar(20);not null"`
	Email         string  `gorm:"column:email;type:varchar(255);not null"`

	// GORM time.Time (DATETIME(3))
	CreatedAt time.Time `gorm:"column:created_at;type:datetime(3);autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime(3);autoUpdateTime"`

	// Unix timestamp (milliseconds)
	UnixTS int64 `gorm:"column:unix_ts;not null"`
}

// TableName explicitly sets table name
func (LoanApplication) TableName() string {
	return "loan_applications"
}
