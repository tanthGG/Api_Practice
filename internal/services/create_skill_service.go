package services

import (
	"context"
	"fmt"
	"strings"

	"go-api-practice/internal/models"
	"go-api-practice/internal/repositories"
)

// ValidationError represents user input errors.
type ValidationError struct {
	Field string
	Msg   string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Msg
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// SkillService contains the create skill business logic.
type SkillService struct {
	repo repositories.SkillRepository
}

func NewSkillService(repo repositories.SkillRepository) *SkillService {
	return &SkillService{repo: repo}
}

func (s *SkillService) CreateSkill(ctx context.Context, input models.SkillInput) (*models.Skill, error) {
	if err := validateSkillInput(input, true); err != nil {
		return nil, err
	}
	skill := &models.Skill{
		Key:         strings.TrimSpace(input.Key),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Logo:        strings.TrimSpace(input.Logo),
		Tags:        cleanTags(input.Tags),
	}
	if err := s.repo.Create(ctx, skill); err != nil {
		return nil, err
	}
	return skill, nil
}

func validateSkillInput(input models.SkillInput, requireKey bool) error {
	if requireKey && strings.TrimSpace(input.Key) == "" {
		return ValidationError{Field: "key", Msg: "is required"}
	}
	if strings.TrimSpace(input.Name) == "" {
		return ValidationError{Field: "name", Msg: "is required"}
	}
	if strings.TrimSpace(input.Description) == "" {
		return ValidationError{Field: "description", Msg: "is required"}
	}
	if strings.TrimSpace(input.Logo) == "" {
		return ValidationError{Field: "logo", Msg: "is required"}
	}
	return nil
}

func cleanTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
