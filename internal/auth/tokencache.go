package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	msalcache "github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const (
	cacheFileName   = "msal_cache.json"
	accountFileName = "account.json"
)

// fileCache implements the MSAL ExportReplace interface, persisting token cache
// data to a JSON file in the config directory.
type fileCache struct {
	mu   sync.Mutex
	path string
}

func newFileCache(configDir string) *fileCache {
	return &fileCache{
		path: filepath.Join(configDir, cacheFileName),
	}
}

func (c *fileCache) Replace(_ context.Context, u msalcache.Unmarshaler, _ msalcache.ReplaceHints) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.path)
	if errors.Is(err, os.ErrNotExist) {
		slog.Debug("token cache file not found, starting fresh", "path", c.path)
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading token cache: %w", err)
	}

	slog.Debug("loaded token cache", "path", c.path, "bytes", len(data))
	return u.Unmarshal(data)
}

func (c *fileCache) Export(_ context.Context, m msalcache.Marshaler, _ msalcache.ExportHints) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := m.Marshal()
	if err != nil {
		return fmt.Errorf("marshaling token cache: %w", err)
	}

	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0o600); err != nil {
		return fmt.Errorf("writing token cache: %w", err)
	}

	slog.Debug("exported token cache", "path", c.path, "bytes", len(data))
	return nil
}

// saveAccount persists an MSAL Account as JSON in the config directory.
func saveAccount(account *public.Account, configDir string) error {
	data, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling account: %w", err)
	}

	p := filepath.Join(configDir, accountFileName)
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("writing account: %w", err)
	}

	slog.Debug("saved account", "path", p)
	return nil
}

// loadAccount loads a previously saved MSAL Account from the config directory.
// Returns a zero value and nil error if the file does not exist.
func loadAccount(configDir string) (public.Account, error) {
	p := filepath.Join(configDir, accountFileName)
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		slog.Debug("account file not found", "path", p)
		return public.Account{}, nil
	}
	if err != nil {
		return public.Account{}, fmt.Errorf("reading account: %w", err)
	}

	var account public.Account
	if err := json.Unmarshal(data, &account); err != nil {
		return public.Account{}, fmt.Errorf("unmarshaling account: %w", err)
	}

	slog.Debug("loaded account", "path", p, "username", account.PreferredUsername)
	return account, nil
}
