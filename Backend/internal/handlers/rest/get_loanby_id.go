package rest

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"go-api-practice/internal/services"
)

type LoanStatusResponse struct {
	ApplicationID string  `json:"applicationId"`
	FullName      string  `json:"fullName"`
	MonthlyIncome float64 `json:"monthlyIncome"`
	LoanAmount    float64 `json:"loanAmount"`
	LoanPurpose   string  `json:"loanPurpose"`
	Age           int     `json:"age"`
	PhoneNumber   string  `json:"phoneNumber"`
	Email         string  `json:"email"`
	Eligible      bool    `json:"eligible"`
	Reason        string  `json:"reason"`
	Timestamp     string  `json:"timestamp"`
}

func (h *LoanHandler) GetStatus(c echo.Context) error {
	applicationID := strings.TrimSpace(c.Param("applicationId"))
	if applicationID == "" {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Message: "Invalid request",
			Reason:  "applicationId is required",
		})
	}

	result, err := h.loanService.GetStatus(c.Request().Context(), applicationID)
	if err != nil {
		if errors.Is(err, services.ErrLoanNotFound) {
			return c.JSON(http.StatusNotFound, errorResponse{
				Message: "Loan application not found",
				Reason:  "applicationId not found: " + applicationID,
			})
		}
		h.logger.Errorf("get loan status failed: %v", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Message: "Unable to process request",
			Reason:  "server error",
		})
	}

	resp := LoanStatusResponse{
		ApplicationID: result.ApplicationID,
		FullName:      result.FullName,
		MonthlyIncome: result.MonthlyIncome,
		LoanAmount:    result.LoanAmount,
		LoanPurpose:   result.LoanPurpose,
		Age:           result.Age,
		PhoneNumber:   result.PhoneNumber,
		Email:         result.Email,
		Eligible:      result.Eligible,
		Reason:        result.Reason,
		Timestamp:     result.Timestamp,
	}
	return c.JSON(http.StatusOK, resp)
}
