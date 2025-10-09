package chatwoot

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Chatwoot struct {
	ID        string  `json:"id" db:"id"`
	SessionID string  `json:"sessionId" db:"sessionId"`
	URL       string  `json:"url" db:"url"`
	Token     string  `json:"token" db:"token"`
	AccountID string  `json:"accountId" db:"accountId"`
	InboxID   *string `json:"inboxId,omitempty" db:"inboxId"`
	Enabled   bool    `json:"enabled" db:"enabled"`

	InboxName      *string  `json:"inboxName,omitempty" db:"inboxName"`
	AutoCreate     bool     `json:"autoCreate" db:"autoCreate"`
	SignMsg        bool     `json:"signMsg" db:"signMsg"`
	SignDelimiter  string   `json:"signDelimiter" db:"signDelimiter"`
	ReopenConv     bool     `json:"reopenConv" db:"reopenConv"`
	ConvPending    bool     `json:"convPending" db:"convPending"`
	ImportContacts bool     `json:"importContacts" db:"importContacts"`
	ImportMessages bool     `json:"importMessages" db:"importMessages"`
	ImportDays     int      `json:"importDays" db:"importDays"`
	MergeBrazil    bool     `json:"mergeBrazil" db:"mergeBrazil"`
	Organization   *string  `json:"organization,omitempty" db:"organization"`
	Logo           *string  `json:"logo,omitempty" db:"logo"`
	Number         *string  `json:"number,omitempty" db:"number"`
	IgnoreJids     []string `json:"ignoreJids,omitempty" db:"ignoreJids"`

	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedAt"`
}

func NewChatwoot(sessionID, url, token, accountID string) *Chatwoot {
	now := time.Now()
	return &Chatwoot{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		URL:            url,
		Token:          token,
		AccountID:      accountID,
		Enabled:        true,
		AutoCreate:     false,
		SignMsg:        false,
		SignDelimiter:  "\n\n",
		ReopenConv:     true,
		ConvPending:    false,
		ImportContacts: false,
		ImportMessages: false,
		ImportDays:     60,
		MergeBrazil:    true,
		IgnoreJids:     []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (c *Chatwoot) IsEnabled() bool {
	return c.Enabled
}

func (c *Chatwoot) Enable() {
	c.Enabled = true
	c.touch()
}

func (c *Chatwoot) Disable() {
	c.Enabled = false
	c.touch()
}

func (c *Chatwoot) UpdateConfiguration(url, token, accountID string, inboxID *string) {
	c.URL = url
	c.Token = token
	c.AccountID = accountID
	c.InboxID = inboxID
	c.touch()
}

func (c *Chatwoot) UpdateAdvancedSettings(settings *AdvancedSettings) {
	if settings.InboxName != nil {
		c.InboxName = settings.InboxName
	}
	if settings.AutoCreate != nil {
		c.AutoCreate = *settings.AutoCreate
	}
	if settings.SignMsg != nil {
		c.SignMsg = *settings.SignMsg
	}
	if settings.SignDelimiter != nil {
		c.SignDelimiter = *settings.SignDelimiter
	}
	if settings.ReopenConv != nil {
		c.ReopenConv = *settings.ReopenConv
	}
	if settings.ConvPending != nil {
		c.ConvPending = *settings.ConvPending
	}
	if settings.ImportContacts != nil {
		c.ImportContacts = *settings.ImportContacts
	}
	if settings.ImportMessages != nil {
		c.ImportMessages = *settings.ImportMessages
	}
	if settings.ImportDays != nil {
		c.ImportDays = *settings.ImportDays
	}
	if settings.MergeBrazil != nil {
		c.MergeBrazil = *settings.MergeBrazil
	}
	if settings.Organization != nil {
		c.Organization = settings.Organization
	}
	if settings.Logo != nil {
		c.Logo = settings.Logo
	}
	if settings.Number != nil {
		c.Number = settings.Number
	}
	if settings.IgnoreJids != nil {
		c.IgnoreJids = *settings.IgnoreJids
	}
	c.touch()
}

func (c *Chatwoot) ShouldIgnoreJID(jid string) bool {
	for _, ignoredJID := range c.IgnoreJids {
		if ignoredJID == jid {
			return true
		}
	}
	return false
}

func (c *Chatwoot) touch() {
	c.UpdatedAt = time.Now()
}

func (c *Chatwoot) GetWebhookURL(baseURL string) string {

	if strings.HasSuffix(baseURL, "/") {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}

	return fmt.Sprintf("%s/%s/webhook/chatwoot", baseURL, c.SessionID)
}

type AdvancedSettings struct {
	InboxName      *string   `json:"inboxName,omitempty"`
	AutoCreate     *bool     `json:"autoCreate,omitempty"`
	SignMsg        *bool     `json:"signMsg,omitempty"`
	SignDelimiter  *string   `json:"signDelimiter,omitempty"`
	ReopenConv     *bool     `json:"reopenConv,omitempty"`
	ConvPending    *bool     `json:"convPending,omitempty"`
	ImportContacts *bool     `json:"importContacts,omitempty"`
	ImportMessages *bool     `json:"importMessages,omitempty"`
	ImportDays     *int      `json:"importDays,omitempty"`
	MergeBrazil    *bool     `json:"mergeBrazil,omitempty"`
	Organization   *string   `json:"organization,omitempty"`
	Logo           *string   `json:"logo,omitempty"`
	Number         *string   `json:"number,omitempty"`
	IgnoreJids     *[]string `json:"ignoreJids,omitempty"`
}
