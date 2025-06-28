package primary

import "context"

// HealthChecker defines the primary port for health checking.
type HealthChecker interface {
	CheckHealth(ctx context.Context) (map[string]string, error)
}
