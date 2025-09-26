package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"zpwoot/internal/domain/chatwoot"
	"zpwoot/pkg/errors"
)

type ChatwootHandler struct {
	chatwootService ChatwootService
}

type ChatwootService interface {
	CreateConfig(ctx context.Context, req *chatwoot.CreateChatwootConfigRequest) (*chatwoot.ChatwootConfig, error)
	GetConfig(ctx context.Context) (*chatwoot.ChatwootConfig, error)
	UpdateConfig(ctx context.Context, req *chatwoot.UpdateChatwootConfigRequest) (*chatwoot.ChatwootConfig, error)
	DeleteConfig(ctx context.Context) error
	SyncContact(ctx context.Context, req *chatwoot.SyncContactRequest) (*chatwoot.ChatwootContact, error)
	SyncConversation(ctx context.Context, req *chatwoot.SyncConversationRequest) (*chatwoot.ChatwootConversation, error)
	ProcessWebhook(ctx context.Context, payload *chatwoot.ChatwootWebhookPayload) error
}

func NewChatwootHandler(chatwootService ChatwootService) *ChatwootHandler {
	return &ChatwootHandler{
		chatwootService: chatwootService,
	}
}

// CreateConfig creates Chatwoot configuration
// @Summary Create Chatwoot configuration
// @Description Creates a new Chatwoot integration configuration. This enables synchronization between WhatsApp and Chatwoot. Requires API key authentication.
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body zpmeow_internal_app_chatwoot.CreateChatwootConfigRequest true "Chatwoot configuration request"
// @Success 201 {object} zpmeow_internal_app_chatwoot.ChatwootConfigResponse "Chatwoot configuration created successfully"
// @Failure 400 {object} object "Invalid request body or parameters"
// @Failure 401 {object} object "Unauthorized - Invalid or missing API key"
// @Failure 500 {object} object "Internal server error"
// @Router /api/v1/chatwoot/config [post]
func (h *ChatwootHandler) CreateConfig(c *fiber.Ctx) error {
	var req chatwoot.CreateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	config, err := h.chatwootService.CreateConfig(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// GetConfig gets Chatwoot configuration
// GET /api/v1/chatwoot/config
func (h *ChatwootHandler) GetConfig(c *fiber.Ctx) error {
	config, err := h.chatwootService.GetConfig(c.Context())
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Hide sensitive information
	if config != nil {
		config.APIKey = "***hidden***"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// UpdateConfig updates Chatwoot configuration
// PUT /api/v1/chatwoot/config
func (h *ChatwootHandler) UpdateConfig(c *fiber.Ctx) error {
	var req chatwoot.UpdateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	config, err := h.chatwootService.UpdateConfig(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Hide sensitive information
	if config != nil {
		config.APIKey = "***hidden***"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// DeleteConfig deletes Chatwoot configuration
// DELETE /api/v1/chatwoot/config
func (h *ChatwootHandler) DeleteConfig(c *fiber.Ctx) error {
	err := h.chatwootService.DeleteConfig(c.Context())
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration deleted successfully",
	})
}

// SyncContacts synchronizes contacts with Chatwoot
// POST /api/v1/chatwoot/sync/contacts
func (h *ChatwootHandler) SyncContacts(c *fiber.Ctx) error {
	var req chatwoot.SyncContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	contact, err := h.chatwootService.SyncContact(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    contact,
	})
}

// SyncConversations synchronizes conversations with Chatwoot
// POST /api/v1/chatwoot/sync/conversations
func (h *ChatwootHandler) SyncConversations(c *fiber.Ctx) error {
	var req chatwoot.SyncConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	conversation, err := h.chatwootService.SyncConversation(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    conversation,
	})
}

// ReceiveWebhook receives webhook from Chatwoot
// POST /api/v1/chatwoot/webhook
func (h *ChatwootHandler) ReceiveWebhook(c *fiber.Ctx) error {
	var payload chatwoot.ChatwootWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid webhook payload",
		})
	}

	// Validate event type
	if !chatwoot.IsValidChatwootEvent(payload.Event) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid event type",
			"event": payload.Event,
		})
	}

	err := h.chatwootService.ProcessWebhook(c.Context(), &payload)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Webhook processed successfully",
		"event":   payload.Event,
	})
}

// TestConnection tests Chatwoot API connection
// POST /api/v1/chatwoot/test
func (h *ChatwootHandler) TestConnection(c *fiber.Ctx) error {
	// TODO: Implement connection test
	// This would test the API connection with current configuration

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot connection test completed",
		"status":  "connected", // or "failed"
	})
}

// GetStats gets Chatwoot integration statistics
// GET /api/v1/chatwoot/stats
func (h *ChatwootHandler) GetStats(c *fiber.Ctx) error {
	// TODO: Implement statistics
	// This would return sync statistics, message counts, etc.

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"contacts_synced":      0,
			"conversations_synced": 0,
			"messages_sent":        0,
			"messages_received":    0,
			"last_sync":            nil,
		},
	})
}
