package validator

import (
	"testing"
)

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"Valid phone with country code", "5511999999999", false},
		{"Valid phone with plus", "+5511999999999", false},
		{"Valid international", "+1234567890", false},
		{"Invalid - too short", "123", true},
		{"Invalid - contains letters", "abc123", true},
		{"Invalid - starts with zero", "00000000000", true},
		{"Empty phone", "", false}, // Empty is valid if not required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.phone, "phone")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJID(t *testing.T) {
	tests := []struct {
		name    string
		jid     string
		wantErr bool
	}{
		{"Valid individual JID", "5511999999999@s.whatsapp.net", false},
		{"Valid group JID", "120363123456789012@g.us", false},
		{"Valid broadcast JID", "123456789@broadcast", false},
		{"Valid newsletter JID", "123456789@newsletter", false},
		{"Invalid - no suffix", "5511999999999", true},
		{"Invalid - wrong domain", "invalid@domain.com", true},
		{"Empty JID", "", false}, // Empty is valid if not required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.jid, "jid")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWhatsAppURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid HTTPS URL", "https://api.example.com/webhook", false},
		{"Valid HTTP URL", "http://localhost:3000/webhook", false},
		{"Invalid - FTP protocol", "ftp://example.com", true},
		{"Invalid - no protocol", "example.com", true},
		{"Invalid - not a URL", "not a url", true},
		{"Empty URL", "", false}, // Empty is valid if not required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.url, "whatsapp_url")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSessionID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
	}{
		{"Valid with dash", "my-session", false},
		{"Valid with underscore", "session_123", false},
		{"Valid alphanumeric", "MySession2024", false},
		{"Invalid - contains space", "my session", true},
		{"Invalid - special char @", "session@123", true},
		{"Invalid - special char #", "session#1", true},
		{"Empty session ID", "", false}, // Empty is valid if not required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.sessionID, "session_id")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWebhookEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   string
		wantErr bool
	}{
		{"Valid - Message", "Message", false},
		{"Valid - Connected", "Connected", false},
		{"Valid - QRCode", "QRCode", false},
		{"Valid - GroupInfo", "GroupInfo", false},
		{"Invalid event", "InvalidEvent", true},
		{"Empty event", "", false}, // Empty is valid if not required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.event, "webhook_event")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMessageType(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		wantErr bool
	}{
		{"Valid - text", "text", false},
		{"Valid - image", "image", false},
		{"Valid - video", "video", false},
		{"Valid - audio", "audio", false},
		{"Valid - document", "document", false},
		{"Valid - location", "location", false},
		{"Invalid type", "invalid", true},
		{"Empty type", "", false}, // Empty is valid if not required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.msgType, "message_type")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePresenceType(t *testing.T) {
	tests := []struct {
		name     string
		presence string
		wantErr  bool
	}{
		{"Valid - available", "available", false},
		{"Valid - unavailable", "unavailable", false},
		{"Valid - composing", "composing", false},
		{"Valid - recording", "recording", false},
		{"Valid - paused", "paused", false},
		{"Invalid presence", "invalid", true},
		{"Empty presence", "", false}, // Empty is valid if not required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVar(tt.presence, "presence_type")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required,min=3"`
		Email string `validate:"required,email"`
		Phone string `validate:"required,phone"`
	}

	// Test with invalid data
	testData := TestStruct{
		Name:  "ab",        // Too short
		Email: "not-email", // Invalid email
		Phone: "abc",       // Invalid phone
	}

	err := Validate(testData)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	errors := FormatValidationErrors(err)
	if len(errors) != 3 {
		t.Errorf("Expected 3 validation errors, got %d", len(errors))
	}

	// Check that errors have proper structure
	for _, e := range errors {
		if e.Field == "" {
			t.Error("Error field should not be empty")
		}
		if e.Message == "" {
			t.Error("Error message should not be empty")
		}
		if e.Tag == "" {
			t.Error("Error tag should not be empty")
		}
	}
}

func TestValidateStruct(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required,min=3,max=50"`
		Phone string `validate:"required,phone"`
	}

	tests := []struct {
		name    string
		data    TestStruct
		wantErr bool
	}{
		{
			name: "Valid data",
			data: TestStruct{
				Name:  "John Doe",
				Phone: "5511999999999",
			},
			wantErr: false,
		},
		{
			name: "Invalid - name too short",
			data: TestStruct{
				Name:  "ab",
				Phone: "5511999999999",
			},
			wantErr: true,
		},
		{
			name: "Invalid - phone format",
			data: TestStruct{
				Name:  "John Doe",
				Phone: "abc123",
			},
			wantErr: true,
		},
		{
			name: "Invalid - missing required fields",
			data: TestStruct{
				Name:  "",
				Phone: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := ValidateStruct(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && len(errors) == 0 {
				t.Error("Expected validation errors, got none")
			}
			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("Expected no validation errors, got %d", len(errors))
			}
		})
	}
}

