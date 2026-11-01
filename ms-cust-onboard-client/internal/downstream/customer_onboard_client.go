package downstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"

	"ms-cust-onboard-client/internal/observability"
)

// CustomerOnboardClient defines methods to talk to ms-customer-onboard
type CustomerOnboardClient interface {
	StartSession(ctx context.Context, req StartSessionRequest) (*DetailResponse, error)
	GetApplication(ctx context.Context, id uuid.UUID) (*DetailResponse, error)
	GetActiveByPhone(ctx context.Context, phone string) (*DetailResponse, error)
	ResumeSession(ctx context.Context, id uuid.UUID) (*DetailResponse, error)
	CancelSession(ctx context.Context, id uuid.UUID, reason string) error
	SubmitApplication(ctx context.Context, id uuid.UUID) (*DetailResponse, error)
	PatchPersonalInfo(ctx context.Context, id uuid.UUID, req PatchPersonalInfoRequest) (*DetailResponse, error)
	PatchAddress(ctx context.Context, id uuid.UUID, req PatchAddressRequest) (*DetailResponse, error)
	PatchIdentityDocs(ctx context.Context, id uuid.UUID, req PatchIdentityDocsRequest) (*DetailResponse, error)
	PatchFinancialInfo(ctx context.Context, id uuid.UUID, req PatchFinancialRequest) (*DetailResponse, error)
	PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*DetailResponse, error)
	ResubmitDocuments(ctx context.Context, id uuid.UUID, req ResubmitDocsRequest) error
	GetAuditLogs(ctx context.Context, id uuid.UUID) ([]AuditLog, error)
	GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*WorkflowProgressResponse, error)
	PatchDraft(ctx context.Context, id uuid.UUID, body interface{}) (*DetailResponse, error)
	HealthCheck(ctx context.Context) (bool, error)
}

type httpClient struct {
	baseURL    string
	client     *http.Client
}

func NewCustomerOnboardClient(baseURL string) CustomerOnboardClient {
	return &httpClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *httpClient) doRequest(ctx context.Context, method, path string, body interface{}, out interface{}) (int, error) {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	observability.InjectTraceContext(ctx, req)

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("downstream request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBytes, &errResp) == nil && errResp.Error != "" {
			return resp.StatusCode, fmt.Errorf("%s", errResp.Error)
		}
		return resp.StatusCode, fmt.Errorf("downstream error status %d: %s", resp.StatusCode, string(respBytes))
	}

	if out != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return resp.StatusCode, fmt.Errorf("failed to unmarshal downstream response: %w", err)
		}
	}

	return resp.StatusCode, nil
}

func (c *httpClient) StartSession(ctx context.Context, req StartSessionRequest) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPost, "/api/v1/onboarding/start", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) GetApplication(ctx context.Context, id uuid.UUID) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/onboarding/%s", id.String()), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) GetActiveByPhone(ctx context.Context, phone string) (*DetailResponse, error) {
	var out DetailResponse
	path := fmt.Sprintf("/api/v1/onboarding/active?phone=%s", url.QueryEscape(phone))
	_, err := c.doRequest(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) ResumeSession(ctx context.Context, id uuid.UUID) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/onboarding/%s/resume", id.String()), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) CancelSession(ctx context.Context, id uuid.UUID, reason string) error {
	req := CancelSessionRequest{Reason: reason}
	_, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/onboarding/%s/cancel", id.String()), req, nil)
	return err
}

func (c *httpClient) SubmitApplication(ctx context.Context, id uuid.UUID) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/onboarding/%s/submit", id.String()), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) PatchPersonalInfo(ctx context.Context, id uuid.UUID, req PatchPersonalInfoRequest) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/onboarding/%s/personal-info", id.String()), req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) PatchAddress(ctx context.Context, id uuid.UUID, req PatchAddressRequest) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/onboarding/%s/address", id.String()), req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) PatchIdentityDocs(ctx context.Context, id uuid.UUID, req PatchIdentityDocsRequest) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/onboarding/%s/identity-documents", id.String()), req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) PatchFinancialInfo(ctx context.Context, id uuid.UUID, req PatchFinancialRequest) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/onboarding/%s/financial-info", id.String()), req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*DetailResponse, error) {
	var out DetailResponse
	req := PatchWalletTierRequest{WalletTier: tier}
	_, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/onboarding/%s/wallet-tier", id.String()), req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) ResubmitDocuments(ctx context.Context, id uuid.UUID, req ResubmitDocsRequest) error {
	_, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/onboarding/%s/resubmit-documents", id.String()), req, nil)
	return err
}

func (c *httpClient) GetAuditLogs(ctx context.Context, id uuid.UUID) ([]AuditLog, error) {
	var logs []AuditLog
	_, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/onboarding/%s/audit-logs", id.String()), nil, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (c *httpClient) GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*WorkflowProgressResponse, error) {
	var progress WorkflowProgressResponse
	_, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/onboarding/%s/workflow-progress", id.String()), nil, &progress)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (c *httpClient) PatchDraft(ctx context.Context, id uuid.UUID, body interface{}) (*DetailResponse, error) {
	var out DetailResponse
	_, err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/onboarding/%s/draft", id.String()), body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *httpClient) HealthCheck(ctx context.Context) (bool, error) {
	fullURL := fmt.Sprintf("%s/healthz", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return false, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
