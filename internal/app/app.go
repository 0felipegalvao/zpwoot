// Package app provides the Application Layer following Clean Architecture principles.
//
// This file serves as the main entry point and index for the application layer, providing:
// - Re-exports of all DTOs (Data Transfer Objects) from domain packages
// - Re-exports of all Use Case interfaces and constructors
// - Centralized access to conversion functions
// - Dependency injection container setup
//
// The application layer orchestrates business logic and serves as the contract
// between the presentation layer (HTTP handlers) and the domain layer.
package app

// Re-export common DTOs for easier imports
import (
	"zpwoot/internal/app/chatwoot"
	"zpwoot/internal/app/common"
	"zpwoot/internal/app/session"
	"zpwoot/internal/app/webhook"
)

// Common response types
type (
	SuccessResponse         = common.SuccessResponse
	ErrorResponse           = common.ErrorResponse
	HealthResponse          = common.HealthResponse
	PaginationResponse      = common.PaginationResponse
	ValidationError         = common.ValidationError
	ValidationErrorResponse = common.ValidationErrorResponse
	APIKeyResponse          = common.APIKeyResponse
	StatusResponse          = common.StatusResponse
	MessageResponse         = common.MessageResponse
)

// Session DTOs
type (
	CreateSessionRequest  = session.CreateSessionRequest
	CreateSessionResponse = session.CreateSessionResponse
	UpdateSessionRequest  = session.UpdateSessionRequest
	ListSessionsRequest   = session.ListSessionsRequest
	ListSessionsResponse  = session.ListSessionsResponse
	SessionInfoResponse   = session.SessionInfoResponse
	SessionResponse       = session.SessionResponse
	DeviceInfoResponse    = session.DeviceInfoResponse
	PairPhoneRequest      = session.PairPhoneRequest
	QRCodeResponse        = session.QRCodeResponse
	SetProxyRequest       = session.SetProxyRequest
	ProxyResponse         = session.ProxyResponse
)

// Webhook DTOs
type (
	CreateWebhookRequest  = webhook.CreateWebhookRequest
	CreateWebhookResponse = webhook.CreateWebhookResponse
	UpdateWebhookRequest  = webhook.UpdateWebhookRequest
	ListWebhooksRequest   = webhook.ListWebhooksRequest
	ListWebhooksResponse  = webhook.ListWebhooksResponse
	WebhookResponse       = webhook.WebhookResponse
	WebhookEventResponse  = webhook.WebhookEventResponse
	TestWebhookRequest    = webhook.TestWebhookRequest
	TestWebhookResponse   = webhook.TestWebhookResponse
	WebhookEventsResponse = webhook.WebhookEventsResponse
	WebhookEventInfo      = webhook.WebhookEventInfo
)

// Chatwoot DTOs
type (
	CreateChatwootConfigRequest    = chatwoot.CreateChatwootConfigRequest
	CreateChatwootConfigResponse   = chatwoot.CreateChatwootConfigResponse
	UpdateChatwootConfigRequest    = chatwoot.UpdateChatwootConfigRequest
	ChatwootConfigResponse         = chatwoot.ChatwootConfigResponse
	SyncContactRequest             = chatwoot.SyncContactRequest
	SyncContactResponse            = chatwoot.SyncContactResponse
	SyncConversationRequest        = chatwoot.SyncConversationRequest
	SyncConversationResponse       = chatwoot.SyncConversationResponse
	SendMessageToChatwootRequest   = chatwoot.SendMessageToChatwootRequest
	ChatwootAttachment             = chatwoot.ChatwootAttachment
	SendMessageToChatwootResponse  = chatwoot.SendMessageToChatwootResponse
	ChatwootWebhookPayload         = chatwoot.ChatwootWebhookPayload
	ChatwootAccount                = chatwoot.ChatwootAccount
	ChatwootConversation           = chatwoot.ChatwootConversation
	ChatwootMessage                = chatwoot.ChatwootMessage
	TestChatwootConnectionResponse = chatwoot.TestChatwootConnectionResponse
	ChatwootStatsResponse          = chatwoot.ChatwootStatsResponse
)

// Helper functions - re-export from common
var (
	NewSuccessResponse         = common.NewSuccessResponse
	NewErrorResponse           = common.NewErrorResponse
	NewValidationErrorResponse = common.NewValidationErrorResponse
	NewPaginationResponse      = common.NewPaginationResponse
)

// Conversion functions - re-export from specific packages
var (
	// Session conversions
	FromSession        = session.FromSession
	FromSessionInfo    = session.FromSessionInfo
	FromQRCodeResponse = session.FromQRCodeResponse

	// Webhook conversions
	FromWebhook        = webhook.FromWebhook
	FromWebhookEvent   = webhook.FromWebhookEvent
	GetSupportedEvents = webhook.GetSupportedEvents

	// Chatwoot conversions
	FromChatwootConfig = chatwoot.FromChatwootConfig
)

// Use Cases - interfaces for business logic orchestration
type (
	// Common use cases
	CommonUseCase = common.UseCase

	// Session use cases
	SessionUseCase = session.UseCase

	// Webhook use cases
	WebhookUseCase = webhook.UseCase

	// Chatwoot use cases
	ChatwootUseCase = chatwoot.UseCase
)

// Use Case constructors
var (
	// Common use case constructor
	NewCommonUseCase = common.NewUseCase

	// Session use case constructor
	NewSessionUseCase = session.NewUseCase

	// Webhook use case constructor
	NewWebhookUseCase = webhook.NewUseCase

	// Chatwoot use case constructor
	NewChatwootUseCase = chatwoot.NewUseCase
)
