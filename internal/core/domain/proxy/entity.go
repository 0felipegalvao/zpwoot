package proxy

import (
	"time"

	"github.com/google/uuid"
)

type Protocol string

const (
	ProtocolHTTP   Protocol = "http"
	ProtocolHTTPS  Protocol = "https"
	ProtocolSOCKS4 Protocol = "socks4"
	ProtocolSOCKS5 Protocol = "socks5"
)

func (p Protocol) IsValid() bool {
	switch p {
	case ProtocolHTTP, ProtocolHTTPS, ProtocolSOCKS4, ProtocolSOCKS5:
		return true
	default:
		return false
	}
}

type ProxyConfig struct {
	ID        string
	SessionID string
	Host      string
	Port      int
	Protocol  Protocol
	Username  string
	Password  string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewProxyConfig(sessionID, host string, port int, protocol Protocol) *ProxyConfig {
	now := time.Now()
	return &ProxyConfig{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Host:      host,
		Port:      port,
		Protocol:  protocol,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (p *ProxyConfig) SetCredentials(username, password string) {
	p.Username = username
	p.Password = password
	p.touch()
}

func (p *ProxyConfig) Enable() {
	p.Enabled = true
	p.touch()
}

func (p *ProxyConfig) Disable() {
	p.Enabled = false
	p.touch()
}

func (p *ProxyConfig) Update(host string, port int, protocol Protocol) {
	p.Host = host
	p.Port = port
	p.Protocol = protocol
	p.touch()
}

func (p *ProxyConfig) GetURL() string {
	if p.Username != "" && p.Password != "" {
		return string(p.Protocol) + "://" + p.Username + ":" + p.Password + "@" + p.Host + ":" + string(rune(p.Port))
	}
	return string(p.Protocol) + "://" + p.Host + ":" + string(rune(p.Port))
}

func (p *ProxyConfig) touch() {
	p.UpdatedAt = time.Now()
}
