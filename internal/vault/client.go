package vault

import (
	"context"
	"errors"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

var (
	ErrEmptyVaultAddress = errors.New("vault address is required")
	ErrEmptyVaultToken   = errors.New("vault token is required")
	ErrSecretNotFound    = errors.New("vault secret not found")
	ErrFieldNotFound     = errors.New("vault field not found")
	ErrFieldNotString    = errors.New("vault field must be a string")
)

type vaultClientFactory func(*vaultapi.Config) (*vaultapi.Client, error)

type logicalReader interface {
	ReadWithContext(context.Context, string) (*vaultapi.Secret, error)
}

type Client struct {
	logical logicalReader
}

func NewClient(address, token string) (*Client, error) {
	return newClient(address, token, vaultapi.NewClient)
}

func newClient(address, token string, newVaultClient vaultClientFactory) (*Client, error) {
	address = strings.TrimSpace(address)
	token = strings.TrimSpace(token)
	if address == "" {
		return nil, ErrEmptyVaultAddress
	}
	if token == "" {
		return nil, ErrEmptyVaultToken
	}

	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	client, err := newVaultClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create vault client: %w", err)
	}

	client.SetToken(token)
	return &Client{logical: client.Logical()}, nil
}

func (c *Client) ReadSecret(ctx context.Context, path string) (map[string]any, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("%w: empty path", ErrSecretNotFound)
	}

	secret, err := c.logical.ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("read secret %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("%w: %s", ErrSecretNotFound, path)
	}

	if data, ok := asMap(secret.Data["data"]); ok {
		return data, nil
	}

	if len(secret.Data) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrSecretNotFound, path)
	}

	return secret.Data, nil
}

func (c *Client) ReadField(ctx context.Context, path, field string) (string, error) {
	data, err := c.ReadSecret(ctx, path)
	if err != nil {
		return "", err
	}

	value, ok := data[field]
	if !ok {
		return "", fmt.Errorf("%w: %s#%s", ErrFieldNotFound, path, field)
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s#%s", ErrFieldNotString, path, field)
	}

	return stringValue, nil
}

// asMap normalizes a dynamic value returned by Vault into map form so we can
// safely read fields from both KV v2 nested `data` payloads and plain map data
func asMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	default:
		return nil, false
	}
}
