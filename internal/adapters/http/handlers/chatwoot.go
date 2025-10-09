package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func (h *ChatwootHandler) Create(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	var req dto.CreateChatwootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode create chatwoot request")
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.chatwootUseCases.Create().Execute(r.Context(), sessionID, &req)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to create chatwoot configuration")
		if err.Error() == "chatwoot configuration already exists for session" {
			h.writeErrorResponse(w, http.StatusConflict, err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create chatwoot configuration")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

func (h *ChatwootHandler) Get(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	response, err := h.chatwootUseCases.Get().Execute(r.Context(), sessionID, h.baseURL)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get chatwoot configuration")
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get chatwoot configuration")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *ChatwootHandler) Update(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	var req dto.UpdateChatwootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode update chatwoot request")
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.chatwootUseCases.Update().Execute(r.Context(), sessionID, &req, h.baseURL)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to update chatwoot configuration")
		if err.Error() == "chatwoot configuration not found" {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update chatwoot configuration")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *ChatwootHandler) Delete(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	err := h.chatwootUseCases.Delete().Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to delete chatwoot configuration")
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to delete chatwoot configuration")
		return
	}

	h.writeJSONResponse(w, http.StatusNoContent, nil)
}

func (h *ChatwootHandler) List(w http.ResponseWriter, r *http.Request) {

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	enabledStr := r.URL.Query().Get("enabled")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var response *dto.ChatwootListResponse
	var err error

	if enabledStr == "true" {
		response, err = h.chatwootUseCases.Get().ListEnabled(r.Context(), limit, offset, h.baseURL)
	} else {
		response, err = h.chatwootUseCases.Get().List(r.Context(), limit, offset, h.baseURL)
	}

	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list chatwoot configurations")
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to list chatwoot configurations")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

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
		"error":   true,
		"message": message,
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
