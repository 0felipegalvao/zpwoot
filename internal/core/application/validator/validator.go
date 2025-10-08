package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// Singleton validator instance
	validate *validator.Validate

	// Regex patterns
	phoneRegex     = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`) // Minimum 8 digits (country code + number)
	jidRegex       = regexp.MustCompile(`^\d+@(s\.whatsapp\.net|g\.us|broadcast|newsletter)$`)
	urlRegex       = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	sessionIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Initialize the validator with custom validators
func init() {
	validate = validator.New()

	// Register custom validators
	_ = validate.RegisterValidation("phone", validatePhone)
	_ = validate.RegisterValidation("jid", validateJID)
	_ = validate.RegisterValidation("whatsapp_url", validateWhatsAppURL)
	_ = validate.RegisterValidation("session_id", validateSessionID)
	_ = validate.RegisterValidation("webhook_event", validateWebhookEvent)
	_ = validate.RegisterValidation("message_type", validateMessageType)
	_ = validate.RegisterValidation("presence_type", validatePresenceType)
}

// GetValidator returns the singleton validator instance
func GetValidator() *validator.Validate {
	return validate
}

// Validate validates a struct
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// Custom validator: phone number (E.164 format)
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Allow empty if not required
	}

	// Remove common formatting
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	return phoneRegex.MatchString(phone)
}

// Custom validator: WhatsApp JID (Java ID)
func validateJID(fl validator.FieldLevel) bool {
	jid := fl.Field().String()
	if jid == "" {
		return true // Allow empty if not required
	}

	return jidRegex.MatchString(jid)
}

// Custom validator: WhatsApp URL (http/https)
func validateWhatsAppURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true // Allow empty if not required
	}

	return urlRegex.MatchString(url)
}

// Custom validator: Session ID (alphanumeric, dash, underscore)
func validateSessionID(fl validator.FieldLevel) bool {
	sessionID := fl.Field().String()
	if sessionID == "" {
		return true // Allow empty if not required
	}

	return sessionIDRegex.MatchString(sessionID)
}

// Custom validator: Webhook event type
func validateWebhookEvent(fl validator.FieldLevel) bool {
	event := fl.Field().String()
	if event == "" {
		return true // Allow empty if not required
	}

	validEvents := []string{
		"Message", "MessageRevoked", "MessageReaction",
		"Connected", "Disconnected", "QRCode", "PairSuccess", "LoggedOut",
		"HistorySync", "Receipt", "ChatPresence",
		"GroupInfo", "JoinedGroup",
		"Picture", "IdentityChange", "PrivacySettings",
		"OfflineSyncPreview", "OfflineSyncCompleted",
		"AppState", "KeepAliveTimeout", "KeepAliveRestored",
		"Blocklist", "MediaRetry",
		"CallOffer", "CallAccept", "CallPreAccept", "CallTransport",
		"CallOfferNotice", "CallRelayLatency", "CallTerminate", "UnknownCallEvent",
		"NewsletterJoin", "NewsletterLeave", "NewsletterMuteChange",
		"NewsletterLiveUpdate", "NewsletterMessageMeta",
	}

	for _, validEvent := range validEvents {
		if event == validEvent {
			return true
		}
	}

	return false
}

// Custom validator: Message type
func validateMessageType(fl validator.FieldLevel) bool {
	msgType := fl.Field().String()
	if msgType == "" {
		return true // Allow empty if not required
	}

	validTypes := []string{
		"text", "image", "video", "audio", "document",
		"sticker", "location", "contact", "reaction",
		"poll", "buttons", "list", "template",
	}

	for _, validType := range validTypes {
		if msgType == validType {
			return true
		}
	}

	return false
}

// Custom validator: Presence type
func validatePresenceType(fl validator.FieldLevel) bool {
	presence := fl.Field().String()
	if presence == "" {
		return true // Allow empty if not required
	}

	validPresences := []string{
		"available", "unavailable", "composing", "recording", "paused",
	}

	for _, validPresence := range validPresences {
		if presence == validPresence {
			return true
		}
	}

	return false
}

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
}

// FormatValidationErrors converts validator errors to a friendly format
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
				Tag:     e.Tag(),
				Value:   fmt.Sprintf("%v", e.Value()),
			})
		}
	}

	return errors
}

// getErrorMessage returns a user-friendly error message based on the validation tag
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number (E.164 format)", field)
	case "jid":
		return fmt.Sprintf("%s must be a valid WhatsApp JID", field)
	case "whatsapp_url":
		return fmt.Sprintf("%s must be a valid HTTP/HTTPS URL", field)
	case "session_id":
		return fmt.Sprintf("%s must contain only alphanumeric characters, dashes, and underscores", field)
	case "webhook_event":
		return fmt.Sprintf("%s must be a valid webhook event type", field)
	case "message_type":
		return fmt.Sprintf("%s must be a valid message type", field)
	case "presence_type":
		return fmt.Sprintf("%s must be a valid presence type", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ValidateStruct is a helper function to validate a struct and return formatted errors
func ValidateStruct(s interface{}) ([]ValidationError, error) {
	err := Validate(s)
	if err != nil {
		return FormatValidationErrors(err), err
	}
	return nil, nil
}

