package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"go-api-practice/internal/services"
)

type loanServiceMock struct {
	applyFunc func(ctx context.Context, input services.ApplyLoanInput) (*services.ApplyLoanResult, error)
	calls     int
	lastInput services.ApplyLoanInput
}

func (m *loanServiceMock) Apply(ctx context.Context, input services.ApplyLoanInput) (*services.ApplyLoanResult, error) {
	m.calls++
	m.lastInput = input
	if m.applyFunc != nil {
		return m.applyFunc(ctx, input)
	}
	return nil, nil
}

func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}

func TestLoanHandlerApply_Success(t *testing.T) {
	mockService := &loanServiceMock{
		applyFunc: func(ctx context.Context, input services.ApplyLoanInput) (*services.ApplyLoanResult, error) {
			return &services.ApplyLoanResult{
				ApplicationID: "loan-123",
				Eligible:      true,
				Reason:        "Eligible under base rules",
				Timestamp:     "2026-02-04T00:00:00Z",
			}, nil
		},
	}
	handler := NewLoanHandler(newTestLogger(), mockService)
	e := echo.New()

	reqBody := ApplyLoanRequest{
		FullName:      "John Doe",
		MonthlyIncome: 20000,
		LoanAmount:    50000,
		LoanPurpose:   "home",
		Age:           30,
		PhoneNumber:   "0851234554",
		Email:         "john@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/loans", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.Apply(c); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp ApplyLoanResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ApplicationID != "loan-123" || !resp.Eligible || resp.Reason != "Eligible under base rules" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if mockService.calls != 1 {
		t.Fatalf("expected service Apply to be called once, got %d", mockService.calls)
	}
}

func TestLoanHandlerApply_Ineligible(t *testing.T) {
	mockService := &loanServiceMock{
		applyFunc: func(ctx context.Context, input services.ApplyLoanInput) (*services.ApplyLoanResult, error) {
			return &services.ApplyLoanResult{
				Eligible:  false,
				Reason:    "Monthly income is insufficient",
				Timestamp: "2026-02-04T00:00:00Z",
			}, &services.IneligibleError{Reason: "Monthly income is insufficient"}
		},
	}
	handler := NewLoanHandler(newTestLogger(), mockService)
	e := echo.New()

	reqBody := ApplyLoanRequest{
		FullName:      "John Doe",
		MonthlyIncome: 8000,
		LoanAmount:    10000,
		LoanPurpose:   "home",
		Age:           30,
		PhoneNumber:   "0851234554",
		Email:         "john@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/loans", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.Apply(c); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp ApplyLoanResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Eligible {
		t.Fatalf("expected eligible=false, got true")
	}
	if resp.Reason != "Monthly income is insufficient" {
		t.Fatalf("unexpected reason: %s", resp.Reason)
	}
}

func TestLoanHandlerApply_InvalidJSON(t *testing.T) {
	mockService := &loanServiceMock{
		applyFunc: func(ctx context.Context, input services.ApplyLoanInput) (*services.ApplyLoanResult, error) {
			t.Fatalf("service should not be called")
			return nil, nil
		},
	}
	handler := NewLoanHandler(newTestLogger(), mockService)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/loans", bytes.NewReader([]byte(`{bad json`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.Apply(c); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if mockService.calls != 0 {
		t.Fatalf("service should not be called on invalid JSON")
	}
}

func TestLoanHandlerApply_InvalidEmail(t *testing.T) {
	mockService := &loanServiceMock{}
	handler := NewLoanHandler(newTestLogger(), mockService)
	e := echo.New()

	reqBody := ApplyLoanRequest{
		FullName:      "John Doe",
		MonthlyIncome: 20000,
		LoanAmount:    50000,
		LoanPurpose:   "home",
		Age:           30,
		PhoneNumber:   "0851234554",
		Email:         "invalid-email",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/loans", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.Apply(c); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if mockService.calls != 0 {
		t.Fatalf("service should not be called when validation fails")
	}
}

func TestLoanHandlerApply_ServiceError(t *testing.T) {
	mockService := &loanServiceMock{
		applyFunc: func(ctx context.Context, input services.ApplyLoanInput) (*services.ApplyLoanResult, error) {
			return nil, errors.New("db down")
		},
	}
	handler := NewLoanHandler(newTestLogger(), mockService)
	e := echo.New()

	reqBody := ApplyLoanRequest{
		FullName:      "John Doe",
		MonthlyIncome: 20000,
		LoanAmount:    50000,
		LoanPurpose:   "home",
		Age:           30,
		PhoneNumber:   "0851234554",
		Email:         "john@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/loans", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.Apply(c); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
