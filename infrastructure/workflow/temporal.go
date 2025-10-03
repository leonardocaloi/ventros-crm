package workflow

import (
	"fmt"

	"go.temporal.io/sdk/client"
)

// Config holds Temporal configuration
type Config struct {
	Host      string
	Namespace string
}

// NewTemporalClient creates a new Temporal client
func NewTemporalClient(cfg Config) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.Host,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}

	return c, nil
}
