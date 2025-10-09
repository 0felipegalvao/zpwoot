package handlers

import (
	"encoding/json"
	"net/http"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/ports/input"
	"zpwoot/internal/core/ports/output"

	"github.com/go-chi/chi/v5"
)

type ChatwootHandler struct {
	chatwootUseCases input.ChatwootUseCases
	webhookHandler   input.ChatwootWebhookHandler
	logger           output.Logger
	baseURL          string
}

func NewChatwootHandler(chatwootUseCases input.ChatwootUseCases, webhookHandler input.ChatwootWebhookHandler, logger output.Logger, baseURL string) *ChatwootHandler {
	return &ChatwootHandler{
		chatwootUseCases: chatwootUseCases,
		webhookHandler:   webhookHandler,
		logger:           logger,
		baseURL:          baseURL,
	}
}

// @Summary		Set Chatwoot Configuration
// @Description	Create or update Chatwoot configuration for a session (upsert operation). To disable/remove integration, set enabled: false
// @Tags			Chatwoot
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string						true	"Session ID"
// @Param			request		body		dto.CreateChatwootRequest	true	"Chatwoot configuration"
// @Success		200			{object}	dto.ChatwootResponse		"Configuration updated successfully"
// @Success		201			{object}	dto.ChatwootResponse		"Configuration created successfully"
// @Failure		400			{object}	dto.ErrorResponse			"Invalid request"
// @Failure		500			{object}	dto.ErrorResponse			"Internal server error"
// @Router			/sessions/{sessionId}/chatwoot/set [post]
// @Security		ApiKeyAuth
func (h *ChatwootHandler) Set(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	var req dto.CreateChatwootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode chatwoot request")
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check if configuration exists
	existing, err := h.chatwootUseCases.Get().Execute(r.Context(), sessionID, h.baseURL)
	isUpdate := err == nil && existing != nil

	var response *dto.ChatwootResponse
	if isUpdate {
		// Convert CreateRequest to UpdateRequest for existing configuration
		updateReq := &dto.UpdateChatwootRequest{
			URL:            &req.URL,
			Token:          &req.Token,
			AccountID:      &req.AccountID,
			InboxID:        req.InboxID,
			Enabled:        req.Enabled,
			InboxName:      req.InboxName,
			AutoCreate:     req.AutoCreate,
			SignMsg:        req.SignMsg,
			SignDelimiter:  req.SignDelimiter,
			ReopenConv:     req.ReopenConv,
			ConvPending:    req.ConvPending,
			ImportContacts: req.ImportContacts,
			ImportMessages: req.ImportMessages,
			ImportDays:     req.ImportDays,
			MergeBrazil:    req.MergeBrazil,
			Organization:   req.Organization,
			Logo:           req.Logo,
			Number:         req.Number,
			IgnoreJids:     &req.IgnoreJids,
		}

		response, err = h.chatwootUseCases.Update().Execute(r.Context(), sessionID, updateReq, h.baseURL)
		if err != nil {
			h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to update chatwoot configuration")
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update chatwoot configuration")
			return
		}
		h.writeJSONResponse(w, http.StatusOK, response)
	} else {
		// Create new configuration
		response, err = h.chatwootUseCases.Create().Execute(r.Context(), sessionID, &req)
		if err != nil {
			h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to create chatwoot configuration")
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create chatwoot configuration")
			return
		}
		h.writeJSONResponse(w, http.StatusCreated, response)
	}
}

// @Summary		Find Chatwoot Configuration
// @Description	Get Chatwoot configuration for a specific session
// @Tags			Chatwoot
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Success		200			{object}	dto.ChatwootResponse	"Configuration found"
// @Failure		400			{object}	dto.ErrorResponse		"Invalid request"
// @Failure		404			{object}	dto.ErrorResponse		"Configuration not found"
// @Failure		500			{object}	dto.ErrorResponse		"Internal server error"
// @Router			/sessions/{sessionId}/chatwoot/find [get]
// @Security		ApiKeyAuth
func (h *ChatwootHandler) Find(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	response, err := h.chatwootUseCases.Get().Execute(r.Context(), sessionID, h.baseURL)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get chatwoot configuration")
		if err.Error() == "chatwoot configuration not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "chatwoot configuration not found for this session")
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get chatwoot configuration")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}



// @Summary		Chatwoot Webhook
// @Description	Receive webhook events from Chatwoot and forward to WhatsApp
// @Tags			Chatwoot
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string							true	"Session ID"
// @Param			webhook		body		dto.ChatwootWebhookRequest		true	"Webhook payload from Chatwoot"
// @Success		200			{object}	dto.ChatwootWebhookResponse		"Webhook processed successfully"
// @Failure		400			{object}	dto.ErrorResponse				"Invalid request"
// @Failure		500			{object}	dto.ErrorResponse				"Internal server error"
// @Router			/sessions/{sessionId}/chatwoot/webhook [post]
// @Security		ApiKeyAuth
func (h *ChatwootHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	var req dto.ChatwootWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode chatwoot webhook request")
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.Info().
		Str("session_id", sessionID).
		Str("event", req.Event).
		Int64("timestamp", req.Timestamp).
		Msg("Received Chatwoot webhook")

	response, err := h.webhookHandler.HandleWebhook(r.Context(), sessionID, &req)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("session_id", sessionID).
			Str("event", req.Event).
			Msg("Failed to process Chatwoot webhook")
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to process webhook")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *ChatwootHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error().Err(err).Msg("Failed to encode JSON response")
		}
	}
}

func (h *ChatwootHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

func (h *ChatwootHandler) GetBaseURL() string {
	return h.baseURL
}

func (h *ChatwootHandler) SetBaseURL(baseURL string) {
	h.baseURL = baseURL
}
