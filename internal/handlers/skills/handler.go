package skills

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"go-api-practice/internal/models"
	"go-api-practice/internal/repositories"
	"go-api-practice/internal/services"
)

// SkillHandler exposes HTTP handlers for skill resources.
type SkillHandler struct {
	service *services.SkillService
	logger  *logrus.Logger
}

func NewSkillHandler(service *services.SkillService, logger *logrus.Logger) *SkillHandler {
	return &SkillHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSkill handles POST /skills requests.
func (h *SkillHandler) CreateSkill(c echo.Context) error {
	var input models.SkillInput
	if err := c.Bind(&input); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid JSON payload")
	}

	ctx := c.Request().Context()
	created, err := h.service.CreateSkill(ctx, input)
	if err != nil {
		return respondError(c, statusFromError(err), err.Error())
	}

	return respondSuccess(c, http.StatusCreated, created)
}

func statusFromError(err error) int {
	var validationErr services.ValidationError
	if errors.As(err, &validationErr) {
		return http.StatusBadRequest
	}
	switch {
	case errors.Is(err, repositories.ErrDuplicateKey):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func respondSuccess(c echo.Context, status int, data interface{}) error {
	payload := map[string]interface{}{
		"status": "success",
	}
	if data != nil {
		payload["data"] = data
	}
	return c.JSON(status, payload)
}

func respondError(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]interface{}{
		"status":  "error",
		"message": message,
	})
}
