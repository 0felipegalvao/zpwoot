package dto

import "time"

type CreateProxyRequest struct {
	Host     string  `json:"host" validate:"required,min=1"`
	Port     int     `json:"port" validate:"required,min=1,max=65535"`
	Protocol string  `json:"protocol" validate:"required,oneof=http https socks4 socks5"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

type UpdateProxyRequest struct {
	Host     *string `json:"host,omitempty" validate:"omitempty,min=1"`
	Port     *int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
	Protocol *string `json:"protocol,omitempty" validate:"omitempty,oneof=http https socks4 socks5"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

type ProxyResponse struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"`
	Username  string    `json:"username,omitempty"`
	Password  string    `json:"password,omitempty"`
	Enabled   bool      `json:"enabled"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ProxyListResponse struct {
	Configurations []ProxyResponse `json:"configurations"`
	Total          int             `json:"total"`
	Limit          int             `json:"limit"`
	Offset         int             `json:"offset"`
}
