package rest

import (
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"go-api-practice/internal/services"
)

type LoanHandler struct {
	logger      *logrus.Logger
	loanService services.LoanService
}

func NewLoanHandler(logger *logrus.Logger, loanService services.LoanService) *LoanHandler {
	return &LoanHandler{
		logger:      logger,
		loanService: loanService,
	}
}

type ApplyLoanRequest struct {
	FullName      string  `json:"fullName"`
	MonthlyIncome float64 `json:"monthlyIncome"`
	LoanAmount    float64 `json:"loanAmount"`
	LoanPurpose   string  `json:"loanPurpose"`
	Age           int     `json:"age"`
	PhoneNumber   string  `json:"phoneNumber"`
	Email         string  `json:"email"`
}

var (
	phonePattern        = regexp.MustCompile(`^[0-9]{10}$`)
	allowedLoanPurposes = map[string]struct{}{
		"education": {},
		"home":      {},
		"car":       {},
		"business":  {},
		"personal":  {},
	}
)

func (h *LoanHandler) Apply(c echo.Context) error {
	var req ApplyLoanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, invalidBodyMissingFields())
	}

	if errResp := validateApplyLoanRequest(req); errResp != nil {
		return c.JSON(http.StatusBadRequest, errResp)
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Message: "Invalid request body",
			Reason:  "email must be a valid email",
		})
	}

	result, err := h.loanService.Apply(c.Request().Context(), services.ApplyLoanInput{
		FullName:      req.FullName,
		MonthlyIncome: req.MonthlyIncome,
		LoanAmount:    req.LoanAmount,
		LoanPurpose:   req.LoanPurpose,
		Age:           req.Age,
		PhoneNumber:   req.PhoneNumber,
		Email:         req.Email,
	})
	if err != nil {
		var ineligibleErr *services.IneligibleError
		if errors.As(err, &ineligibleErr) && result != nil {
			resp := ApplyLoanResponse{
				ApplicationID: result.ApplicationID,
				Eligible:      result.Eligible,
				Reason:        ineligibleErr.Error(),
				Timestamp:     result.Timestamp,
			}
			return c.JSON(http.StatusOK, resp)
		}
		h.logger.Errorf("apply loan failed: %v", err)
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Message: "Unable to process request",
			Reason:  "server error",
		})
	}

	resp := ApplyLoanResponse{
		ApplicationID: result.ApplicationID,
		Eligible:      result.Eligible,
		Reason:        result.Reason,
		Timestamp:     result.Timestamp,
	}

	h.logger.Infof("loan application submitted for %s", req.FullName)
	return c.JSON(http.StatusOK, resp)
}

func validateApplyLoanRequest(req ApplyLoanRequest) *errorResponse {
	fullName := strings.TrimSpace(req.FullName)
	loanPurpose := strings.TrimSpace(req.LoanPurpose)
	email := strings.TrimSpace(req.Email)

	if fullName == "" ||
		req.MonthlyIncome == 0 ||
		req.LoanAmount == 0 ||
		loanPurpose == "" ||
		req.Age == 0 ||
		req.PhoneNumber == "" ||
		email == "" {
		return invalidBodyMissingFields()
	}

	if len(fullName) < 2 || len(fullName) > 255 {
		return &errorResponse{
			Message: "Invalid request body",
			Reason:  "fullName must be between 2 and 255 characters",
		}
	}

	if req.MonthlyIncome < 5000 || req.MonthlyIncome > 5000000 {
		return &errorResponse{
			Message: "Invalid request body",
			Reason:  "monthlyIncome must be between 5,000 and 5,000,000",
		}
	}

	if req.LoanAmount < 1000 || req.LoanAmount > 5000000 {
		return &errorResponse{
			Message: "Invalid request body",
			Reason:  "loanAmount must be between 1,000 and 5,000,000",
		}
	}

	if _, ok := allowedLoanPurposes[strings.ToLower(loanPurpose)]; !ok {
		return &errorResponse{
			Message: "Invalid request body",
			Reason:  "loanPurpose must be one of: education, home, car, business, personal",
		}
	}

	if req.Age <= 0 {
		return &errorResponse{
			Message: "Invalid request body",
			Reason:  "age must be greater than 0",
		}
	}

	if !phonePattern.MatchString(req.PhoneNumber) {
		return &errorResponse{
			Message: "Invalid request body",
			Reason:  "phoneNumber must be 10 digits",
		}
	}

	return nil
}

func invalidBodyMissingFields() *errorResponse {
	return &errorResponse{
		Message: "Invalid request body",
		Reason:  "missing required fields: fullName, monthlyIncome, loanAmount, loanPurpose, age, phoneNumber, email",
	}
}
