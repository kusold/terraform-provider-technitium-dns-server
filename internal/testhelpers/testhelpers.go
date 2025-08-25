package testhelpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

const (
	TechnitiumImage   = "technitium/dns-server:13.6.0"
	DefaultUsername   = "admin"
	DefaultPassword   = "admin"
	TechnitiumAPIPort = "5380/tcp"
)

// TechnitiumContainer represents a running Technitium DNS Server container
type TechnitiumContainer struct {
	testcontainers.Container
	Host     string
	Port     string
	Username string
	Password string
}

// StartTechnitiumContainer starts a new Technitium DNS Server container for testing
func StartTechnitiumContainer(ctx context.Context, t *testing.T) (*TechnitiumContainer, error) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        TechnitiumImage,
		ExposedPorts: []string{TechnitiumAPIPort},
		Env: map[string]string{
			"DNS_SERVER_DOMAIN":                           "dns-server",
			"DNS_SERVER_ADMIN_PASSWORD":                   DefaultPassword,
			"DNS_SERVER_ADMIN_PASSWORD_FILE":              "",
			"DNS_SERVER_PREFER_IPV6":                      "false",
			"DNS_SERVER_WEB_SERVICE_HTTP_PORT":            "5380",
			"DNS_SERVER_WEB_SERVICE_ENABLE_HTTPS":         "false",
			"DNS_SERVER_WEB_SERVICE_USE_SELF_SIGNED_CERT": "false",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(TechnitiumAPIPort),
			wait.ForHTTP("/api/user/login").WithPort(TechnitiumAPIPort).WithStartupTimeout(60*time.Second),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, TechnitiumAPIPort)
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	return &TechnitiumContainer{
		Container: container,
		Host:      host,
		Port:      port.Port(),
		Username:  DefaultUsername,
		Password:  DefaultPassword,
	}, nil
}

// GetAPIURL returns the complete API URL for the container
func (tc *TechnitiumContainer) GetAPIURL() string {
	return fmt.Sprintf("http://%s:%s", tc.Host, tc.Port)
}

// Cleanup terminates the container
func (tc *TechnitiumContainer) Cleanup(ctx context.Context) error {
	return tc.Terminate(ctx)
}

// CreateTestClient creates a client for testing against the container
func CreateTestClient(host, username, password string) (*client.Client, error) {
	clientConfig := client.Config{
		Host:               host,
		Username:           username,
		Password:           password,
		TimeoutSeconds:     30,
		RetryAttempts:      3,
		InsecureSkipVerify: false,
	}

	return client.NewClient(clientConfig)
}
