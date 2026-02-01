package rest

type ApplyLoanResponse struct {
	ApplicationID string `json:"applicationId"`
	Eligible      bool   `json:"eligible"`
	Reason        string `json:"reason"`
	Timestamp     string `json:"timestamp"`
}

type errorResponse struct {
	Message string `json:"message"`
	Reason  string `json:"reason"`
}
