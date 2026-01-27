package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go-api-practice/internal/models"
)

var (
	// ErrDuplicateKey indicates the key already exists.
	ErrDuplicateKey = errors.New("skill key already exists")
)

// SkillRepository defines persistence operations needed for create.
type SkillRepository interface {
	Create(ctx context.Context, skill *models.Skill) error
}

// MySQLSkillRepository is a MySQL-backed repository implementation.
type MySQLSkillRepository struct {
	db *sql.DB
}

func NewMySQLSkillRepository(db *sql.DB) *MySQLSkillRepository {
	return &MySQLSkillRepository{db: db}
}

func (r *MySQLSkillRepository) Create(ctx context.Context, skill *models.Skill) error {
	now := time.Now().UTC()
	skill.CreatedAt = now
	skill.UpdatedAt = now

	tagsJSON, err := json.Marshal(skill.Tags)
	if err != nil {
		return err
	}

	const query = "INSERT INTO skills (`key`, name, description, logo, tags, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err = r.db.ExecContext(ctx, query, skill.Key, skill.Name, skill.Description, skill.Logo, string(tagsJSON), skill.CreatedAt, skill.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicateKey
		}
		return err
	}
	return nil
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "1062")
}
