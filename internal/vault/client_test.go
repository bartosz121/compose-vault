package vault

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func TestNewClient(t *testing.T) {
	cases := []struct {
		name      string
		address   string
		token     string
		wantErr   bool
		wantErrIs error
	}{
		{
			"happy path",
			"http://127.0.0.1:8200",
			"token",
			false,
			nil,
		},
		{
			"address is empty string",
			"",
			"token",
			true,
			ErrEmptyVaultAddress,
		},
		{
			"token is empty string",
			"addr",
			"",
			true,
			ErrEmptyVaultToken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewClient(tc.address, tc.token)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected err, got nil")
				}
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("expected %v, got %v", tc.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == nil {
				t.Fatalf("expected client, got nil")
			}

		})
	}

}

func TestNewClientTrimsWhitespace(t *testing.T) {
	t.Parallel()

	var apiClient *vaultapi.Client
	var gotAddress string

	vaultClientFactory := func(cfg *vaultapi.Config) (*vaultapi.Client, error) {
		gotAddress = cfg.Address

		c, err := vaultapi.NewClient(vaultapi.DefaultConfig())
		if err != nil {
			return nil, err
		}

		apiClient = c
		return c, nil
	}

	_, err := newClient(" address    ", "   token ", vaultClientFactory)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAddress != "address" {
		t.Fatalf("expected trimmed address, got %q", gotAddress)
	}

	if apiClient.Token() != "token" {
		t.Fatalf("expected trimmed token, got %q", apiClient.Token())
	}
}

func TestNewClientFactoryError(t *testing.T) {
	t.Parallel()

	factoryErr := errors.New("factory failed")
	vaultClientFactory := func(cfg *vaultapi.Config) (*vaultapi.Client, error) {
		return nil, factoryErr
	}

	got, err := newClient("http://127.0.0.1:8200", "token", vaultClientFactory)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, factoryErr) {
		t.Fatalf("expected wrapped error %v, got %v", factoryErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil client, got %#v", got)
	}
}

type FakeLogicalReader struct {
	secret  *vaultapi.Secret
	err     error
	gotPath string
}

func (f *FakeLogicalReader) ReadWithContext(_ context.Context, path string) (*vaultapi.Secret, error) {
	f.gotPath = path
	return f.secret, f.err
}

func TestClientReadSecret(t *testing.T) {
	readErr := errors.New("vault unavailable")

	cases := []struct {
		name         string
		path         string
		secret       *vaultapi.Secret
		readErr      error
		want         map[string]any
		wantReadPath string
		wantErr      bool
		wantErrIs    error
		errContains  string
	}{
		{
			"reads secret successfully",
			"app1/db",
			&vaultapi.Secret{Data: map[string]any{"a": "b"}},
			nil,
			map[string]any{"a": "b"},
			"app1/db",
			false,
			nil,
			"",
		},
		{
			"trims path before read",
			"  app1/db  ",
			&vaultapi.Secret{Data: map[string]any{"a": "b"}},
			nil,
			map[string]any{"a": "b"},
			"app1/db",
			false,
			nil,
			"",
		},
		{
			"returns nested kv v2 data",
			"secret/data/app1/db",
			&vaultapi.Secret{Data: map[string]any{"data": map[string]any{"password": "secret"}}},
			nil,
			map[string]any{"password": "secret"},
			"secret/data/app1/db",
			false,
			nil,
			"",
		},
		{
			"returns not found for empty path",
			"   ",
			nil,
			nil,
			nil,
			"",
			true,
			ErrSecretNotFound,
			"empty path",
		},
		{
			"wraps logical reader error",
			"app1/db",
			nil,
			readErr,
			nil,
			"app1/db",
			true,
			readErr,
			`read secret "app1/db"`,
		},
		{
			"returns not found when secret is nil",
			"app1/db",
			nil,
			nil,
			nil,
			"app1/db",
			true,
			ErrSecretNotFound,
			"app1/db",
		},
		{
			"returns not found for empty secret data",
			"app1/db",
			&vaultapi.Secret{Data: map[string]any{}},
			nil,
			nil,
			"app1/db",
			true,
			ErrSecretNotFound,
			"app1/db",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := &FakeLogicalReader{secret: tc.secret, err: tc.readErr}
			client := &Client{logical: f}

			got, err := client.ReadSecret(context.Background(), tc.path)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected err, not nil")
				}

				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("expected %v, got %v", tc.wantErrIs, err)
				}

				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if f.gotPath != tc.wantReadPath {
				t.Fatalf("got read path %q, want %q", f.gotPath, tc.wantReadPath)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}

		})
	}
}

func TestClientReadField(t *testing.T) {
	readErr := errors.New("vault unavailable")

	cases := []struct {
		name         string
		path         string
		field        string
		secret       *vaultapi.Secret
		readErr      error
		want         string
		wantReadPath string
		wantErr      bool
		wantErrIs    error
		errContains  string
	}{
		{
			"reads field successfully",
			"app1/db",
			"password",
			&vaultapi.Secret{Data: map[string]any{"password": "secret"}},
			nil,
			"secret",
			"app1/db",
			false,
			nil,
			"",
		},
		{
			"reads field from nested kv v2 data",
			"secret/data/app1/db",
			"password",
			&vaultapi.Secret{Data: map[string]any{"data": map[string]any{"password": "secret"}}},
			nil,
			"secret",
			"secret/data/app1/db",
			false,
			nil,
			"",
		},
		{
			"returns error when field is missing",
			"app1/db",
			"password",
			&vaultapi.Secret{Data: map[string]any{"username": "admin"}},
			nil,
			"",
			"app1/db",
			true,
			ErrFieldNotFound,
			"app1/db#password",
		},
		{
			"returns error when field is not a string",
			"app1/db",
			"port",
			&vaultapi.Secret{Data: map[string]any{"port": 5432}},
			nil,
			"",
			"app1/db",
			true,
			ErrFieldNotString,
			"app1/db#port",
		},
		{
			"propagates read secret errors",
			"app1/db",
			"password",
			nil,
			readErr,
			"",
			"app1/db",
			true,
			readErr,
			`read secret "app1/db"`,
		},
		{
			"propagates not found from read secret",
			"app1/db",
			"password",
			nil,
			nil,
			"",
			"app1/db",
			true,
			ErrSecretNotFound,
			"app1/db",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := &FakeLogicalReader{secret: tc.secret, err: tc.readErr}
			client := &Client{logical: f}

			got, err := client.ReadField(context.Background(), tc.path, tc.field)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected err, got nil")
				}

				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("expected %v, got %v", tc.wantErrIs, err)
				}

				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if f.gotPath != tc.wantReadPath {
				t.Fatalf("got read path %q, want %q", f.gotPath, tc.wantReadPath)
			}

			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
