package webhook

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const minSecretLength = 16

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ValidateURL(webhookURL string) error {
	if webhookURL == "" {
		return ErrInvalidURL
	}

	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: must use http or https", ErrInvalidURL)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: missing host", ErrInvalidURL)
	}

	if !s.isLocalhostAllowed() && s.isLocalhost(parsedURL.Host) {
		return fmt.Errorf("%w: localhost not allowed", ErrInvalidURL)
	}

	return nil
}

func (s *Service) ValidateEvents(events []string) error {
	if len(events) == 0 {
		return nil
	}

	validEvents := s.buildValidEventMap()
	for _, event := range events {
		if !validEvents[event] {
			return fmt.Errorf("%w: %s", ErrInvalidEvent, event)
		}
	}

	return nil
}

func (s *Service) ValidateSecret(secret string) error {
	if secret == "" {
		return ErrInvalidSecret
	}
	if len(secret) < minSecretLength {
		return fmt.Errorf("%w: must be at least %d characters", ErrInvalidSecret, minSecretLength)
	}
	return nil
}
func (s *Service) GetValidEventTypes() []string {
	return []string{
		"Message", "MessageRevoked", "MessageReaction", "Receipt",
		"Connected", "Disconnected", "QRCode", "PairSuccess", "LoggedOut",
		"KeepAliveTimeout", "KeepAliveRestored",
		"GroupInfo", "JoinedGroup",
		"Picture", "IdentityChange", "PrivacySettings", "Blocklist", "ChatPresence",
		"HistorySync", "OfflineSyncPreview", "OfflineSyncCompleted", "AppState",
		"CallOffer", "CallAccept", "CallPreAccept", "CallTransport",
		"CallOfferNotice", "CallRelayLatency", "CallTerminate", "UnknownCallEvent",
		"NewsletterJoin", "NewsletterLeave", "NewsletterMuteChange",
		"NewsletterLiveUpdate", "NewsletterMessageMeta",
		"MediaRetry",
	}
}
func (s *Service) GetEventCategories() map[string][]string {
	return map[string][]string{
		"Messages":   {"Message", "MessageRevoked", "MessageReaction", "Receipt"},
		"Connection": {"Connected", "Disconnected", "QRCode", "PairSuccess", "LoggedOut", "KeepAliveTimeout", "KeepAliveRestored"},
		"Groups":     {"GroupInfo", "JoinedGroup"},
		"User":       {"Picture", "IdentityChange", "PrivacySettings", "Blocklist", "ChatPresence"},
		"Sync":       {"HistorySync", "OfflineSyncPreview", "OfflineSyncCompleted", "AppState"},
		"Calls":      {"CallOffer", "CallAccept", "CallPreAccept", "CallTransport", "CallOfferNotice", "CallRelayLatency", "CallTerminate", "UnknownCallEvent"},
		"Newsletter": {"NewsletterJoin", "NewsletterLeave", "NewsletterMuteChange", "NewsletterLiveUpdate", "NewsletterMessageMeta"},
		"Media":      {"MediaRetry"},
	}
}

func (s *Service) buildValidEventMap() map[string]bool {
	events := s.GetValidEventTypes()
	eventMap := make(map[string]bool, len(events))
	for _, e := range events {
		eventMap[e] = true
	}
	return eventMap
}

func (s *Service) isLocalhostAllowed() bool {
	env := os.Getenv("NODE_ENV")
	return env == "development" || env == ""
}

func (s *Service) isLocalhost(host string) bool {
	return strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1")
}
