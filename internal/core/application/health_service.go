package application

import (
	"context"
	"go_hex/internal/core/ports/primary"
	"go_hex/internal/core/ports/secondary"
	"log/slog"
)

// HealthService provides health checking functionality.
type HealthService struct {
	friendRepo secondary.FriendRepository
	logger     *slog.Logger
}

// Ensure HealthService implements the HealthChecker interface.
var _ primary.HealthChecker = (*HealthService)(nil)

func NewHealthService(friendRepo secondary.FriendRepository, logger *slog.Logger) *HealthService {
	return &HealthService{
		friendRepo: friendRepo,
		logger:     logger,
	}
}

func (s *HealthService) CheckHealth(ctx context.Context) (map[string]string, error) {
	status := make(map[string]string)

	// Check application health
	status["status"] = "healthy"
	status["service"] = "go_hex"

	// Check repository health
	_, err := s.friendRepo.GetAllFriends()
	if err != nil {
		s.logger.Error("Repository health check failed", "error", err)
		status["repository"] = "unhealthy"
		status["repository_error"] = err.Error()
		status["status"] = "degraded"
	} else {
		status["repository"] = "healthy"
	}

	s.logger.Debug("Health check completed", "status", status["status"])
	return status, nil
}
