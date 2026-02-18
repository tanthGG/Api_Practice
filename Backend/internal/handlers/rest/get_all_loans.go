package rest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type ListLoansResponse struct {
	Applications []LoanStatusResponse `json:"applications"`
	Page         int                  `json:"page"`
	TotalPages   int                  `json:"totalPages"`
}

func (h *LoanHandler) ListLoans(c echo.Context) error {
	page := parsePositiveInt(c.QueryParam("page"), 1)
	limit := parsePositiveInt(c.QueryParam("limit"), 10)

	var eligiblePtr *bool
	if val := strings.TrimSpace(c.QueryParam("eligible")); val != "" {
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return c.JSON(http.StatusBadRequest, errorResponse{
				Message: "Invalid request",
				Reason:  "eligible must be true or false",
			})
		}
		eligiblePtr = &parsed
	}

	var purposePtr *string
	if val := strings.TrimSpace(c.QueryParam("purpose")); val != "" {
		lower := strings.ToLower(val)
		if _, ok := allowedLoanPurposes[lower]; !ok {
			return c.JSON(http.StatusBadRequest, errorResponse{
				Message: "Invalid request",
				Reason:  "loanPurpose must be one of: education, home, car, business, personal",
			})
		}
		purposePtr = &lower
	}

	result, err := h.loanService.ListLoans(c.Request().Context(), page, limit, eligiblePtr, purposePtr)
	if err != nil {
		h.logger.Errorf("list loans failed: %v", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Message: "Unable to process request",
			Reason:  "server error",
		})
	}

	resp := ListLoansResponse{
		Page:       page,
		TotalPages: result.TotalPages,
	}

	for _, app := range result.Applications {
		resp.Applications = append(resp.Applications, LoanStatusResponse{
			ApplicationID: app.ApplicationID,
			FullName:      app.FullName,
			MonthlyIncome: app.MonthlyIncome,
			LoanAmount:    app.LoanAmount,
			LoanPurpose:   app.LoanPurpose,
			Age:           app.Age,
			PhoneNumber:   app.PhoneNumber,
			Email:         app.Email,
			Eligible:      app.Eligible,
			Reason:        app.Reason,
			Timestamp:     app.Timestamp,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
