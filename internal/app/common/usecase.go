package common

import (
	"context"
	"runtime"
	"time"
)

// UseCase defines the common use case interface
type UseCase interface {
	GetHealth(ctx context.Context) (*HealthResponse, error)
	GetVersion(ctx context.Context) (*VersionResponse, error)
	GetStats(ctx context.Context) (*StatsResponse, error)
}

// VersionResponse represents version information
type VersionResponse struct {
	Version   string `json:"version" example:"1.0.0"`
	BuildTime string `json:"build_time" example:"2024-01-01T00:00:00Z"`
	GitCommit string `json:"git_commit,omitempty" example:"abc123"`
	GoVersion string `json:"go_version" example:"go1.21.0"`
} // @name VersionResponse

// StatsResponse represents system statistics
type StatsResponse struct {
	Uptime          string          `json:"uptime" example:"2h30m15s"`
	MemoryUsage     MemoryStats     `json:"memory_usage"`
	GoroutineCount  int             `json:"goroutine_count" example:"25"`
	RequestCount    int64           `json:"request_count" example:"1250"`
	ErrorCount      int64           `json:"error_count" example:"5"`
	ActiveSessions  int             `json:"active_sessions" example:"10"`
	ActiveWebhooks  int             `json:"active_webhooks" example:"3"`
	DatabaseStatus  string          `json:"database_status" example:"connected"`
	LastHealthCheck time.Time       `json:"last_health_check" example:"2024-01-01T00:00:00Z"`
	Features        map[string]bool `json:"features"`
} // @name StatsResponse

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc      uint64 `json:"alloc" example:"1048576"`
	TotalAlloc uint64 `json:"total_alloc" example:"5242880"`
	Sys        uint64 `json:"sys" example:"10485760"`
	NumGC      uint32 `json:"num_gc" example:"10"`
} // @name MemoryStats

// useCaseImpl implements the common use case
type useCaseImpl struct {
	startTime time.Time
	version   string
	buildTime string
	gitCommit string
}

// NewUseCase creates a new common use case
func NewUseCase(version, buildTime, gitCommit string) UseCase {
	return &useCaseImpl{
		startTime: time.Now(),
		version:   version,
		buildTime: buildTime,
		gitCommit: gitCommit,
	}
}

// GetHealth returns the health status of the application
func (uc *useCaseImpl) GetHealth(ctx context.Context) (*HealthResponse, error) {
	uptime := time.Since(uc.startTime)

	response := &HealthResponse{
		Status:  "ok",
		Service: "zpwoot",
		Version: uc.version,
		Uptime:  uptime.String(),
	}

	return response, nil
}

// GetVersion returns version information
func (uc *useCaseImpl) GetVersion(ctx context.Context) (*VersionResponse, error) {
	response := &VersionResponse{
		Version:   uc.version,
		BuildTime: uc.buildTime,
		GitCommit: uc.gitCommit,
		GoVersion: runtime.Version(),
	}

	return response, nil
}

// GetStats returns system statistics
func (uc *useCaseImpl) GetStats(ctx context.Context) (*StatsResponse, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(uc.startTime)

	response := &StatsResponse{
		Uptime:         uptime.String(),
		GoroutineCount: runtime.NumGoroutine(),
		MemoryUsage: MemoryStats{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
		DatabaseStatus:  "connected", // TODO: Check actual database status
		LastHealthCheck: time.Now(),
		Features: map[string]bool{
			"sessions":      true,
			"webhooks":      true,
			"chatwoot":      true,
			"swagger_docs":  true,
			"health_checks": true,
			"metrics":       true,
		},
		// TODO: Implement actual counters
		RequestCount:   0,
		ErrorCount:     0,
		ActiveSessions: 0,
		ActiveWebhooks: 0,
	}

	return response, nil
}
