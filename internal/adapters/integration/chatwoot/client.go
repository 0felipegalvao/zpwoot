package chatwoot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"zpwoot/internal/core/ports/output"
)

type Client struct {
	baseURL    string
	token      string
	accountID  string
	httpClient *http.Client
	logger     output.Logger
}

func NewClient(baseURL, token, accountID string, logger output.Logger) *Client {
	return &Client{
		baseURL:   baseURL,
		token:     token,
		accountID: accountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

type InboxRequest struct {
	Name        string `json:"name"`
	Channel     string `json:"channel"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Provider    string `json:"provider,omitempty"`
	WebhookURL  string `json:"webhook_url,omitempty"`
}

type InboxResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Channel     string `json:"channel"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Provider    string `json:"provider,omitempty"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ContactRequest struct {
	Name             string                 `json:"name,omitempty"`
	Email            string                 `json:"email,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	Identifier       string                 `json:"identifier,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}

type ContactResponse struct {
	ID               int                    `json:"id"`
	Name             string                 `json:"name"`
	Email            string                 `json:"email"`
	PhoneNumber      string                 `json:"phone_number"`
	Identifier       string                 `json:"identifier"`
	CustomAttributes map[string]interface{} `json:"custom_attributes"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

type ConversationRequest struct {
	SourceID   string `json:"source_id"`
	InboxID    int    `json:"inbox_id"`
	ContactID  int    `json:"contact_id,omitempty"`
	Status     string `json:"status,omitempty"`
	AssigneeID int    `json:"assignee_id,omitempty"`
}

type ConversationResponse struct {
	ID         int                    `json:"id"`
	SourceID   string                 `json:"source_id"`
	InboxID    int                    `json:"inbox_id"`
	ContactID  int                    `json:"contact_id"`
	Status     string                 `json:"status"`
	AssigneeID int                    `json:"assignee_id"`
	Messages   []MessageResponse      `json:"messages,omitempty"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}

type MessageRequest struct {
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"`
	Private     bool                   `json:"private,omitempty"`
	ContentType string                 `json:"content_type,omitempty"`
	Attachments []AttachmentRequest    `json:"attachments,omitempty"`
	Echo        map[string]interface{} `json:"echo,omitempty"`
}

type AttachmentRequest struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name,omitempty"`
}

type MessageResponse struct {
	ID          int                    `json:"id"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"`
	ContentType string                 `json:"content_type"`
	Private     bool                   `json:"private"`
	SourceID    string                 `json:"source_id"`
	Sender      map[string]interface{} `json:"sender"`
	Echo        map[string]interface{} `json:"echo"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type ErrorResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s/api/v1/accounts/%s%s", c.baseURL, c.accountID, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_access_token", c.token)

	c.logger.Debug().
		Str("method", method).
		Str("url", url).
		Msg("Making Chatwoot API request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug().
		Int("status_code", resp.StatusCode).
		Str("response_body", string(body)).
		Msg("Chatwoot API response")

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Message)
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *Client) CreateInbox(ctx context.Context, req *InboxRequest) (*InboxResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/inboxes", req)
	if err != nil {
		return nil, err
	}

	var result InboxResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetInbox(ctx context.Context, inboxID int) (*InboxResponse, error) {
	endpoint := fmt.Sprintf("/inboxes/%d", inboxID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result InboxResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) UpdateInbox(ctx context.Context, inboxID int, req *InboxRequest) (*InboxResponse, error) {
	endpoint := fmt.Sprintf("/inboxes/%d", inboxID)
	resp, err := c.makeRequest(ctx, "PATCH", endpoint, req)
	if err != nil {
		return nil, err
	}

	var result InboxResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CreateContact(ctx context.Context, req *ContactRequest) (*ContactResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/contacts", req)
	if err != nil {
		return nil, err
	}

	var result ContactResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetContactByIdentifier(ctx context.Context, identifier string) (*ContactResponse, error) {
	endpoint := fmt.Sprintf("/contacts/search?q=%s", identifier)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var results []ContactResponse
	if err := c.handleResponse(resp, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("contact not found")
	}

	return &results[0], nil
}

func (c *Client) CreateConversation(ctx context.Context, req *ConversationRequest) (*ConversationResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/conversations", req)
	if err != nil {
		return nil, err
	}

	var result ConversationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetConversation(ctx context.Context, conversationID int) (*ConversationResponse, error) {
	endpoint := fmt.Sprintf("/conversations/%d", conversationID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result ConversationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) SendMessage(ctx context.Context, conversationID int, req *MessageRequest) (*MessageResponse, error) {
	endpoint := fmt.Sprintf("/conversations/%d/messages", conversationID)
	resp, err := c.makeRequest(ctx, "POST", endpoint, req)
	if err != nil {
		return nil, err
	}

	var result MessageResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) UpdateConversationStatus(ctx context.Context, conversationID int, status string) error {
	endpoint := fmt.Sprintf("/conversations/%d/toggle_status", conversationID)
	body := map[string]string{"status": status}

	resp, err := c.makeRequest(ctx, "POST", endpoint, body)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}
