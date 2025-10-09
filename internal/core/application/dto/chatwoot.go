package dto

import "time"

type CreateChatwootRequest struct {
	URL       string  `json:"url" validate:"required,url"`
	Token     string  `json:"token" validate:"required,min=1"`
	AccountID string  `json:"accountId" validate:"required,min=1"`
	InboxID   *string `json:"inboxId,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`

	InboxName      *string  `json:"inboxName,omitempty"`
	AutoCreate     *bool    `json:"autoCreate,omitempty"`
	SignMsg        *bool    `json:"signMsg,omitempty"`
	SignDelimiter  *string  `json:"signDelimiter,omitempty"`
	ReopenConv     *bool    `json:"reopenConv,omitempty"`
	ConvPending    *bool    `json:"convPending,omitempty"`
	ImportContacts *bool    `json:"importContacts,omitempty"`
	ImportMessages *bool    `json:"importMessages,omitempty"`
	ImportDays     *int     `json:"importDays,omitempty" validate:"omitempty,min=1,max=365"`
	MergeBrazil    *bool    `json:"mergeBrazil,omitempty"`
	Organization   *string  `json:"organization,omitempty"`
	Logo           *string  `json:"logo,omitempty" validate:"omitempty,url"`
	Number         *string  `json:"number,omitempty"`
	IgnoreJids     []string `json:"ignoreJids,omitempty"`
}

type UpdateChatwootRequest struct {
	URL       *string `json:"url,omitempty" validate:"omitempty,url"`
	Token     *string `json:"token,omitempty" validate:"omitempty,min=1"`
	AccountID *string `json:"accountId,omitempty" validate:"omitempty,min=1"`
	InboxID   *string `json:"inboxId,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`

	InboxName      *string   `json:"inboxName,omitempty"`
	AutoCreate     *bool     `json:"autoCreate,omitempty"`
	SignMsg        *bool     `json:"signMsg,omitempty"`
	SignDelimiter  *string   `json:"signDelimiter,omitempty"`
	ReopenConv     *bool     `json:"reopenConv,omitempty"`
	ConvPending    *bool     `json:"convPending,omitempty"`
	ImportContacts *bool     `json:"importContacts,omitempty"`
	ImportMessages *bool     `json:"importMessages,omitempty"`
	ImportDays     *int      `json:"importDays,omitempty" validate:"omitempty,min=1,max=365"`
	MergeBrazil    *bool     `json:"mergeBrazil,omitempty"`
	Organization   *string   `json:"organization,omitempty"`
	Logo           *string   `json:"logo,omitempty" validate:"omitempty,url"`
	Number         *string   `json:"number,omitempty"`
	IgnoreJids     *[]string `json:"ignoreJids,omitempty"`
}

type ChatwootResponse struct {
	ID        string  `json:"id"`
	SessionID string  `json:"sessionId"`
	URL       string  `json:"url"`
	Token     string  `json:"token,omitempty"`
	AccountID string  `json:"accountId"`
	InboxID   *string `json:"inboxId,omitempty"`
	Enabled   bool    `json:"enabled"`

	InboxName      *string  `json:"inboxName,omitempty"`
	AutoCreate     bool     `json:"autoCreate"`
	SignMsg        bool     `json:"signMsg"`
	SignDelimiter  string   `json:"signDelimiter"`
	ReopenConv     bool     `json:"reopenConv"`
	ConvPending    bool     `json:"convPending"`
	ImportContacts bool     `json:"importContacts"`
	ImportMessages bool     `json:"importMessages"`
	ImportDays     int      `json:"importDays"`
	MergeBrazil    bool     `json:"mergeBrazil"`
	Organization   *string  `json:"organization,omitempty"`
	Logo           *string  `json:"logo,omitempty"`
	Number         *string  `json:"number,omitempty"`
	IgnoreJids     []string `json:"ignoreJids"`

	WebhookURL string `json:"webhookUrl"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChatwootListResponse struct {
	Configurations []ChatwootResponse `json:"configurations"`
	Total          int                `json:"total"`
	Limit          int                `json:"limit"`
	Offset         int                `json:"offset"`
}

type ChatwootWebhookRequest struct {
	Event        string               `json:"event"`
	Account      ChatwootAccount      `json:"account"`
	Conversation ChatwootConversation `json:"conversation,omitempty"`
	Message      *ChatwootMessage     `json:"message,omitempty"`
	Contact      *ChatwootContact     `json:"contact,omitempty"`
	Inbox        *ChatwootInbox       `json:"inbox,omitempty"`
	ChangedBy    *ChatwootUser        `json:"changed_by,omitempty"`
	Timestamp    int64                `json:"timestamp"`
}

type ChatwootAccount struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ChatwootConversation struct {
	ID                int                    `json:"id"`
	Messages          []ChatwootMessage      `json:"messages,omitempty"`
	Contact           ChatwootContact        `json:"contact"`
	Inbox             ChatwootInbox          `json:"inbox"`
	Status            string                 `json:"status"`
	AgentLastSeenAt   *int64                 `json:"agent_last_seen_at"`
	ContactLastSeenAt *int64                 `json:"contact_last_seen_at"`
	Timestamp         int64                  `json:"timestamp"`
	CreatedAt         int64                  `json:"created_at"`
	UpdatedAt         int64                  `json:"updated_at"`
	Labels            []string               `json:"labels,omitempty"`
	CustomAttributes  map[string]interface{} `json:"custom_attributes,omitempty"`
}

type ChatwootMessage struct {
	ID           int                  `json:"id"`
	Content      string               `json:"content"`
	MessageType  int                  `json:"message_type"`
	ContentType  string               `json:"content_type"`
	Private      bool                 `json:"private"`
	SourceID     string               `json:"source_id,omitempty"`
	CreatedAt    int64                `json:"created_at"`
	Inbox        ChatwootInbox        `json:"inbox"`
	Conversation ChatwootConversation `json:"conversation"`
	Sender       *ChatwootUser        `json:"sender,omitempty"`
	Contact      *ChatwootContact     `json:"contact,omitempty"`
	Attachments  []ChatwootAttachment `json:"attachments,omitempty"`
	Echo         bool                 `json:"echo,omitempty"`
}

type ChatwootContact struct {
	ID               int                    `json:"id"`
	Name             string                 `json:"name"`
	Email            string                 `json:"email,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	Identifier       string                 `json:"identifier,omitempty"`
	Thumbnail        string                 `json:"thumbnail,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
	ContactInboxes   []ChatwootContactInbox `json:"contact_inboxes,omitempty"`
}

type ChatwootContactInbox struct {
	SourceID string `json:"source_id"`
	InboxID  int    `json:"inbox_id"`
}

type ChatwootInbox struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	ChannelType          string `json:"channel_type"`
	GreetingEnabled      bool   `json:"greeting_enabled,omitempty"`
	GreetingMessage      string `json:"greeting_message,omitempty"`
	WorkingHoursEnabled  bool   `json:"working_hours_enabled,omitempty"`
	EnableAutoAssignment bool   `json:"enable_auto_assignment,omitempty"`
}

type ChatwootUser struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	Type               string `json:"type"`
	Thumbnail          string `json:"thumbnail,omitempty"`
	AvailabilityStatus string `json:"availability_status,omitempty"`
}

type ChatwootAttachment struct {
	ID       int    `json:"id"`
	FileType string `json:"file_type"`
	FileSize int    `json:"file_size"`
	DataURL  string `json:"data_url"`
	ThumbURL string `json:"thumb_url,omitempty"`
}

type ChatwootWebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
