package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

type Client struct {
	baseURL string
	userID  string
	http    *http.Client
}

func NewClient(baseURL, userID string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid baseURL: %w", err)
	}
	return &Client{
		baseURL: baseURL,
		userID:  userID,
		http:    http.DefaultClient,
	}, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-user-id", c.userID)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, target interface{}) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var wrapper struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if wrapper.Code != 0 {
		return fmt.Errorf("API error code %d", wrapper.Code)
	}

	if target != nil {
		if err := json.Unmarshal(wrapper.Data, target); err != nil {
			return fmt.Errorf("failed to decode data: %w", err)
		}
	}

	return nil
}

func (c *Client) GetProfile(ctx context.Context) (*models.UserProfile, error) {
	req, err := c.newRequest(ctx, "GET", "/v1/profile", nil)
	if err != nil {
		return nil, err
	}

	var profile models.UserProfile
	if err := c.do(req, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (c *Client) UpdateProfile(ctx context.Context, profile *models.UserProfile) error {
	body, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}
	req, err := c.newRequest(ctx, "PUT", "/v1/profile", bytes.NewReader(body))
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

func (c *Client) QueryPolicies(ctx context.Context, regionCode, policyType, status string) ([]models.PolicyClaim, error) {
	path := "/v1/policies?"
	params := ""
	if regionCode != "" {
		params += "region_code=" + regionCode + "&"
	}
	if policyType != "" {
		params += "policy_type=" + policyType + "&"
	}
	if status != "" {
		params += "status=" + status + "&"
	}
	if params != "" {
		path += params[:len(params)-1]
	}

	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var claims []models.PolicyClaim
	if err := c.do(req, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func (c *Client) GeneratePlan(ctx context.Context, jsonBody string) (*models.PlanSnapshot, error) {
	req, err := c.newRequest(ctx, "POST", "/v1/plans/generate", bytes.NewBufferString(jsonBody))
	if err != nil {
		return nil, err
	}

	var plan models.PlanSnapshot
	if err := c.do(req, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}
