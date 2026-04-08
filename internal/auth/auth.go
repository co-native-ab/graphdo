// Package auth provides authentication flows for Microsoft identity.
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	authorityURL = "https://login.microsoftonline.com/common"
)

// Scopes defines the Microsoft Graph permissions required by graphdo.
var Scopes = []string{
	"Mail.Send",
	"Tasks.ReadWrite",
	"User.Read",
}

// Authenticator abstracts the login and token acquisition flow.
type Authenticator interface {
	Login(ctx context.Context) error
	Token(ctx context.Context) (string, error)
}

// BrowserAuthenticator implements Authenticator using the interactive browser flow
// with file-based token caching.
type BrowserAuthenticator struct {
	clientID  string
	configDir string
}

// NewBrowserAuthenticator creates an authenticator using the interactive browser flow.
func NewBrowserAuthenticator(clientID, configDir string) *BrowserAuthenticator {
	return &BrowserAuthenticator{clientID: clientID, configDir: configDir}
}

func (a *BrowserAuthenticator) newClient() (public.Client, error) {
	fc := newFileCache(a.configDir)
	return public.New(a.clientID,
		public.WithAuthority(authorityURL),
		public.WithCache(fc),
	)
}

// Login performs an interactive browser login and caches the resulting token.
func (a *BrowserAuthenticator) Login(ctx context.Context) error {
	slog.Debug("starting interactive browser login")

	client, err := a.newClient()
	if err != nil {
		return fmt.Errorf("creating MSAL client: %w", err)
	}

	result, err := client.AcquireTokenInteractive(ctx, Scopes)
	if err != nil {
		return fmt.Errorf("interactive browser login: %w", err)
	}

	if err := saveAccount(&result.Account, a.configDir); err != nil {
		return fmt.Errorf("saving account: %w", err)
	}

	slog.Info("login successful", "username", result.Account.PreferredUsername)
	return nil
}

// Token acquires a cached access token, refreshing silently if needed.
func (a *BrowserAuthenticator) Token(ctx context.Context) (string, error) {
	client, err := a.newClient()
	if err != nil {
		return "", fmt.Errorf("creating MSAL client: %w", err)
	}

	account, err := loadAccount(a.configDir)
	if err != nil {
		return "", fmt.Errorf("loading account: %w", err)
	}

	if account.IsZero() {
		return "", fmt.Errorf("not logged in — run 'graphdo login' first")
	}

	slog.Debug("acquiring token silently", "username", account.PreferredUsername)
	result, err := client.AcquireTokenSilent(ctx, Scopes,
		public.WithSilentAccount(account),
	)
	if err != nil {
		return "", fmt.Errorf("acquiring token (run 'graphdo login' to re-authenticate): %w", err)
	}

	slog.Debug("token acquired")
	return result.AccessToken, nil
}

// DeviceCodeAuthenticator implements Authenticator using the device code flow
// with file-based token caching.
type DeviceCodeAuthenticator struct {
	clientID  string
	configDir string
}

// NewDeviceCodeAuthenticator creates an authenticator using the device code flow.
func NewDeviceCodeAuthenticator(clientID, configDir string) *DeviceCodeAuthenticator {
	return &DeviceCodeAuthenticator{clientID: clientID, configDir: configDir}
}

func (a *DeviceCodeAuthenticator) newClient() (public.Client, error) {
	fc := newFileCache(a.configDir)
	return public.New(a.clientID,
		public.WithAuthority(authorityURL),
		public.WithCache(fc),
	)
}

// Login performs a device code login and caches the resulting token.
func (a *DeviceCodeAuthenticator) Login(ctx context.Context) error {
	slog.Debug("starting device code login")

	client, err := a.newClient()
	if err != nil {
		return fmt.Errorf("creating MSAL client: %w", err)
	}

	dc, err := client.AcquireTokenByDeviceCode(ctx, Scopes)
	if err != nil {
		return fmt.Errorf("initiating device code flow: %w", err)
	}

	fmt.Fprintln(os.Stderr, dc.Result.Message)

	result, err := dc.AuthenticationResult(ctx)
	if err != nil {
		return fmt.Errorf("device code login: %w", err)
	}

	if err := saveAccount(&result.Account, a.configDir); err != nil {
		return fmt.Errorf("saving account: %w", err)
	}

	slog.Info("login successful", "username", result.Account.PreferredUsername)
	return nil
}

// Token acquires a cached access token, refreshing silently if needed.
func (a *DeviceCodeAuthenticator) Token(ctx context.Context) (string, error) {
	client, err := a.newClient()
	if err != nil {
		return "", fmt.Errorf("creating MSAL client: %w", err)
	}

	account, err := loadAccount(a.configDir)
	if err != nil {
		return "", fmt.Errorf("loading account: %w", err)
	}

	if account.IsZero() {
		return "", fmt.Errorf("not logged in — run 'graphdo login' first")
	}

	slog.Debug("acquiring token silently", "username", account.PreferredUsername)
	result, err := client.AcquireTokenSilent(ctx, Scopes,
		public.WithSilentAccount(account),
	)
	if err != nil {
		return "", fmt.Errorf("acquiring token (run 'graphdo login' to re-authenticate): %w", err)
	}

	slog.Debug("token acquired")
	return result.AccessToken, nil
}

// ClearCache removes the MSAL token cache and account files from the config
// directory. It silently ignores files that do not exist.
func ClearCache(configDir string) error {
	files := []string{
		filepath.Join(configDir, cacheFileName),
		filepath.Join(configDir, accountFileName),
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing %s: %w", filepath.Base(f), err)
		}
		slog.Debug("removed cache file", "path", f)
	}

	return nil
}

// StaticAuthenticator implements Authenticator with a fixed token.
// Used for testing and the --access-token flag.
type StaticAuthenticator struct {
	token string
}

// NewStaticAuthenticator creates an authenticator with a pre-set access token.
func NewStaticAuthenticator(token string) *StaticAuthenticator {
	return &StaticAuthenticator{token: token}
}

// Login is a no-op for a static authenticator.
func (a *StaticAuthenticator) Login(_ context.Context) error {
	slog.Debug("static authenticator login (no-op)")
	return nil
}

// Token returns the static access token.
func (a *StaticAuthenticator) Token(_ context.Context) (string, error) {
	return a.token, nil
}
