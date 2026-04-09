package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/co-native-ab/graphdo/internal/auth"
	"github.com/co-native-ab/graphdo/internal/config"

	"github.com/charmbracelet/huh"
)

// LoginCmd is the argument type for the login subcommand.
type LoginCmd struct {
	ClientID string `arg:"--client-id,env:GRAPHDO_CLIENT_ID" help:"Azure AD app client ID (saved to config)"`
}

// LogoutCmd is the argument type for the logout subcommand.
type LogoutCmd struct{}

// StatusCmd is the argument type for the status subcommand.
type StatusCmd struct{}

// ConfigCmd is the argument type for the config subcommand.
type ConfigCmd struct {
	Show *ConfigShowCmd `arg:"subcommand:show" help:"show current configuration"`
}

// ConfigShowCmd is the argument type for the config show subcommand.
type ConfigShowCmd struct{}

func runLogin(ctx context.Context, cmd *LoginCmd, deps *Dependencies) error {
	slog.Debug("starting login")

	// Resolve client ID: flag → existing config → default.
	cfg, err := config.Load(deps.ConfigDir)
	if err != nil {
		slog.Debug("could not load config for client ID resolution", "error", err)
		cfg = &config.Config{}
	}
	clientID := config.ResolveClientID(cmd.ClientID, cfg.ClientID, DefaultClientID)

	// Create a one-off authenticator with the resolved client ID.
	var authenticator auth.Authenticator
	if deps.Authenticator != nil {
		// Use injected authenticator (e.g. StaticAuthenticator in tests).
		authenticator = deps.Authenticator
	} else {
		authenticator = auth.NewBrowserAuthenticator(clientID, deps.ConfigDir)
	}

	if err := authenticator.Login(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Persist client ID to config.
	cfg.ClientID = clientID
	if err := config.Save(cfg, deps.ConfigDir); err != nil {
		return fmt.Errorf("saving client ID to config: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Logged in successfully")
	return nil
}

func runLogout(deps *Dependencies) error {
	slog.Debug("clearing auth cache", "config_dir", deps.ConfigDir)

	if err := auth.ClearCache(deps.ConfigDir); err != nil {
		return fmt.Errorf("clearing auth cache: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Logged out successfully")
	return nil
}

// statusResult is the JSON output of the status command.
type statusResult struct {
	Ready     bool   `json:"ready"`
	LoggedIn  bool   `json:"logged_in"`
	User      string `json:"user,omitempty"`
	TodoList  string `json:"todo_list,omitempty"`
	TodoCount int    `json:"todo_count,omitempty"`
	Error     string `json:"error,omitempty"`
}

func runStatus(ctx context.Context, deps *Dependencies) error {
	result := statusResult{}

	// Check auth explicitly — status is a diagnostic command.
	if _, err := deps.Authenticator.Token(ctx); err != nil {
		result.Error = "not logged in — run 'graphdo login'"
		return outputStatus(result, deps)
	}
	result.LoggedIn = true

	user, err := deps.GraphClient.GetMe(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("failed to reach Microsoft Graph: %v", err)
		return outputStatus(result, deps)
	}
	result.User = user.Mail
	if result.User == "" {
		result.User = user.UserPrincipalName
	}

	// Check config.
	cfg, err := config.Load(deps.ConfigDir)
	if err != nil || cfg.TodoListID == "" {
		result.Error = "todo list not configured — run 'graphdo config'"
		return outputStatus(result, deps)
	}
	result.TodoList = cfg.TodoListName

	// Verify the todo list is accessible.
	items, err := deps.GraphClient.ListTodos(ctx, cfg.TodoListID, 1, 0)
	if err != nil {
		result.Error = fmt.Sprintf("cannot access todo list: %v", err)
		return outputStatus(result, deps)
	}
	result.TodoCount = len(items)
	result.Ready = true

	return outputStatus(result, deps)
}

func outputStatus(result statusResult, deps *Dependencies) error {
	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encoding status: %w", err)
	}
	return nil
}

func runConfig(ctx context.Context, cmd *ConfigCmd, deps *Dependencies) error {
	if cmd.Show != nil {
		return runConfigShow(deps)
	}

	return runConfigSelect(ctx, deps)
}

func runConfigShow(deps *Dependencies) error {
	cfg, err := config.Load(deps.ConfigDir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	return nil
}

func runConfigSelect(ctx context.Context, deps *Dependencies) error {
	lists, err := deps.GraphClient.ListTodoLists(ctx)
	if err != nil {
		return fmt.Errorf("listing todo lists: %w", err)
	}

	if len(lists) == 0 {
		return fmt.Errorf("no todo lists found — create one in Microsoft To Do first")
	}

	options := make([]huh.Option[int], len(lists))
	for i, l := range lists {
		options[i] = huh.NewOption(l.DisplayName, i)
	}

	var selected int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select a todo list").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("running select form: %w", err)
	}

	// Load existing config to preserve fields like client_id.
	cfg, err := config.Load(deps.ConfigDir)
	if err != nil {
		slog.Debug("could not load existing config, starting fresh", "error", err)
		cfg = &config.Config{}
	}

	cfg.TodoListID = lists[selected].ID
	cfg.TodoListName = lists[selected].DisplayName

	if err := config.Save(cfg, deps.ConfigDir); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	_, _ = fmt.Fprintf(deps.Stderr, "✓ Selected list: %s\n", cfg.TodoListName)
	return nil
}
