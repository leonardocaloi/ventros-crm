package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Config holds Temporal configuration
type Config struct {
	Host      string
	Namespace string
}

// DBConfig holds PostgreSQL configuration for Temporal databases
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// NewTemporalClient creates a new Temporal client and ensures namespace exists
func NewTemporalClient(cfg Config) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.Host,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}

	// Ensure namespace exists (similar to RabbitMQ queue creation)
	if err := ensureNamespaceExists(c, cfg.Namespace); err != nil {
		return nil, fmt.Errorf("failed to ensure namespace exists: %w", err)
	}

	return c, nil
}

// ensureNamespaceExists creates the namespace if it doesn't exist
func ensureNamespaceExists(c client.Client, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get workflow service client
	workflowService := c.WorkflowService()

	// Try to describe the namespace
	_, err := workflowService.DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{
		Namespace: namespace,
	})
	if err == nil {
		// Namespace already exists
		return nil
	}

	// Namespace doesn't exist, create it
	retention := durationpb.New(168 * time.Hour) // 7 days
	_, err = workflowService.RegisterNamespace(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        namespace,
		Description:                      "Ventros CRM namespace - auto-created",
		WorkflowExecutionRetentionPeriod: retention,
	})
	if err != nil {
		return fmt.Errorf("failed to register namespace: %w", err)
	}

	return nil
}
