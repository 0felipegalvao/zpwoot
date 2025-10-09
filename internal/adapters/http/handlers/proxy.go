package handlers

import (
	"encoding/json"
	"net/http"

	"zpwoot/internal/core/application/dto"
	"zpwoot/internal/core/ports/input"
	"zpwoot/internal/core/ports/output"

	"github.com/go-chi/chi/v5"
)

type ProxyHandler struct {
	proxyUseCases input.ProxyUseCases
	logger        output.Logger
}

func NewProxyHandler(proxyUseCases input.ProxyUseCases, logger output.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyUseCases: proxyUseCases,
		logger:        logger,
	}
}

// @Summary		Set Proxy Configuration
// @Description	Create or update proxy configuration for a session (upsert operation). To disable/remove proxy, set enabled: false
// @Tags			Proxy
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string					true	"Session ID"
// @Param			request		body		dto.CreateProxyRequest	true	"Proxy configuration"
// @Success		200			{object}	map[string]interface{}	"Configuration updated successfully"
// @Success		201			{object}	map[string]interface{}	"Configuration created successfully"
// @Failure		400			{object}	dto.ErrorResponse		"Invalid request"
// @Failure		500			{object}	dto.ErrorResponse		"Internal server error"
// @Router			/sessions/{sessionId}/proxy/set [post]
// @Security		ApiKeyAuth
func (h *ProxyHandler) Set(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	var req dto.CreateProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode proxy request")
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check if configuration exists
	existing, err := h.proxyUseCases.Get().Execute(r.Context(), sessionID)
	isUpdate := err == nil && existing != nil

	var response *dto.ProxyResponse
	if isUpdate {
		// Convert CreateRequest to UpdateRequest for existing configuration
		updateReq := &dto.UpdateProxyRequest{
			Host:     &req.Host,
			Port:     &req.Port,
			Protocol: &req.Protocol,
			Username: req.Username,
			Password: req.Password,
			Enabled:  req.Enabled,
		}
		
		response, err = h.proxyUseCases.Update().Execute(r.Context(), sessionID, updateReq)
		if err != nil {
			h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to update proxy configuration")
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update proxy configuration")
			return
		}
		h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    response,
		})
	} else {
		// Create new configuration
		response, err = h.proxyUseCases.Create().Execute(r.Context(), sessionID, &req)
		if err != nil {
			h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to create proxy configuration")
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create proxy configuration")
			return
		}
		h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"data":    response,
		})
	}
}

// @Summary		Find Proxy Configuration
// @Description	Get proxy configuration for a specific session
// @Tags			Proxy
// @Accept			json
// @Produce		json
// @Param			sessionId	path		string				true	"Session ID"
// @Success		200			{object}	map[string]interface{}	"Configuration found"
// @Failure		400			{object}	dto.ErrorResponse	"Invalid request"
// @Failure		404			{object}	dto.ErrorResponse	"Configuration not found"
// @Failure		500			{object}	dto.ErrorResponse	"Internal server error"
// @Router			/sessions/{sessionId}/proxy/find [get]
// @Security		ApiKeyAuth
func (h *ProxyHandler) Find(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "session ID is required")
		return
	}

	response, err := h.proxyUseCases.Get().Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get proxy configuration")
		if err.Error() == "proxy configuration not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "proxy configuration not found for this session")
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get proxy configuration")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

func (h *ProxyHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error().Err(err).Msg("Failed to encode JSON response")
		}
	}
}

func (h *ProxyHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]interface{}{
		"success": false,
		"error":   message,
		"status":  statusCode,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}
