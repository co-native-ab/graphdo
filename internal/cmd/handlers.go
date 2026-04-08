package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/co-native-ab/graphdo/internal/auth"
	"github.com/co-native-ab/graphdo/internal/config"
	"github.com/co-native-ab/graphdo/internal/graph"

	"github.com/alexflint/go-arg"
	"github.com/charmbracelet/huh"
)

// Dependencies holds shared resources injected into command handlers.
type Dependencies struct {
	Authenticator auth.Authenticator
	GraphURL      string
	ConfigDir     string
	SkillContent  string
	Stdout        io.Writer
	Stderr        io.Writer
	Stdin         io.Reader
}

// Dispatch routes the parsed CLI args to the appropriate command handler.
func Dispatch(ctx context.Context, p *arg.Parser, cliArgs *Args, deps *Dependencies) error {
	switch {
	case cliArgs.Login != nil:
		return runLogin(ctx, cliArgs.Login, deps)
	case cliArgs.Logout != nil:
		return runLogout(deps)
	case cliArgs.Status != nil:
		return runStatus(ctx, deps)
	case cliArgs.Config != nil:
		return runConfig(ctx, cliArgs.Config, deps)
	case cliArgs.Mail != nil:
		return runMail(ctx, cliArgs.Mail, deps)
	case cliArgs.Todo != nil:
		return runTodo(ctx, cliArgs.Todo, deps)
	case cliArgs.Skill != nil:
		return runSkill(cliArgs.Skill, deps)
	default:
		p.WriteHelp(deps.Stderr)
		return nil
	}
}

// newGraphClient creates a Graph API client from the dependencies.
func newGraphClient(ctx context.Context, deps *Dependencies) (*graph.Client, error) {
	token, err := deps.Authenticator.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting access token: %w", err)
	}

	return graph.NewClient(deps.GraphURL, token), nil
}

// loadConfig loads and validates the config, returning a helpful error if unconfigured.
func loadConfig(configDir string) (*config.Config, error) {
	cfg, err := config.Load(configDir)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config invalid (run 'graphdo config' first): %w", err)
	}

	return cfg, nil
}

// --- Command handlers ---

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

	// Check auth by acquiring a token and calling /me.
	token, err := deps.Authenticator.Token(ctx)
	if err != nil {
		result.Error = "not logged in — run 'graphdo login'"
		return outputStatus(result, deps)
	}
	result.LoggedIn = true

	client := graph.NewClient(deps.GraphURL, token)

	user, err := client.GetMe(ctx)
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
	items, err := client.ListTodos(ctx, cfg.TodoListID, 1, 0)
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
	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	lists, err := client.ListTodoLists(ctx)
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

func runMail(ctx context.Context, cmd *MailCmd, deps *Dependencies) error {
	if cmd.Send == nil {
		return fmt.Errorf("missing subcommand — run 'graphdo mail --help' for usage")
	}

	return runMailSend(ctx, cmd.Send, deps)
}

func runMailSend(ctx context.Context, cmd *MailSendCmd, deps *Dependencies) error {
	body := cmd.Body
	if body == "-" {
		data, err := io.ReadAll(deps.Stdin)
		if err != nil {
			return fmt.Errorf("reading body from stdin: %w", err)
		}
		body = string(data)
	}

	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	user, err := client.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("getting user profile: %w", err)
	}

	email := user.Mail
	if email == "" {
		email = user.UserPrincipalName
	}

	slog.Debug("sending email", "to", email, "subject", cmd.Subject, "html", cmd.HTML)

	if err := client.SendMail(ctx, email, cmd.Subject, body, cmd.HTML); err != nil {
		return fmt.Errorf("sending mail: %w", err)
	}

	_, _ = fmt.Fprintf(deps.Stderr, "✓ Email sent to %s\n", email)
	return nil
}

func runTodo(ctx context.Context, cmd *TodoCmd, deps *Dependencies) error {
	switch {
	case cmd.List != nil:
		return runTodoList(ctx, cmd.List, deps)
	case cmd.Show != nil:
		return runTodoShow(ctx, cmd.Show, deps)
	case cmd.Create != nil:
		return runTodoCreate(ctx, cmd.Create, deps)
	case cmd.Update != nil:
		return runTodoUpdate(ctx, cmd.Update, deps)
	case cmd.Complete != nil:
		return runTodoComplete(ctx, cmd.Complete, deps)
	case cmd.Delete != nil:
		return runTodoDelete(ctx, cmd.Delete, deps)
	default:
		return fmt.Errorf("missing subcommand — run 'graphdo todo --help' for usage")
	}
}

func runTodoList(ctx context.Context, cmd *TodoListCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	items, err := client.ListTodos(ctx, cfg.TodoListID, cmd.Top, cmd.Skip)
	if err != nil {
		return fmt.Errorf("listing todos: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		return fmt.Errorf("encoding todos: %w", err)
	}

	return nil
}

func runTodoShow(ctx context.Context, cmd *TodoShowCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	slog.Debug("getting todo", "id", cmd.ID, "list", cfg.TodoListID)

	item, err := client.GetTodo(ctx, cfg.TodoListID, cmd.ID)
	if err != nil {
		return fmt.Errorf("getting todo: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		return fmt.Errorf("encoding todo: %w", err)
	}

	return nil
}

func runTodoUpdate(ctx context.Context, cmd *TodoUpdateCmd, deps *Dependencies) error {
	if cmd.Title == "" && cmd.Body == "" {
		return fmt.Errorf("at least one of --title or --body must be provided")
	}

	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	slog.Debug("updating todo", "id", cmd.ID, "list", cfg.TodoListID)

	item, err := client.UpdateTodo(ctx, cfg.TodoListID, cmd.ID, cmd.Title, cmd.Body)
	if err != nil {
		return fmt.Errorf("updating todo: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		return fmt.Errorf("encoding todo: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Task updated")
	return nil
}

func runTodoCreate(ctx context.Context, cmd *TodoCreateCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	slog.Debug("creating todo", "title", cmd.Title, "list", cfg.TodoListID)

	item, err := client.CreateTodo(ctx, cfg.TodoListID, cmd.Title, cmd.Body)
	if err != nil {
		return fmt.Errorf("creating todo: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		return fmt.Errorf("encoding todo: %w", err)
	}

	return nil
}

func runTodoComplete(ctx context.Context, cmd *TodoCompleteCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	slog.Debug("completing todo", "id", cmd.ID, "list", cfg.TodoListID)

	if err := client.CompleteTodo(ctx, cfg.TodoListID, cmd.ID); err != nil {
		return fmt.Errorf("completing todo: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Task marked as completed")
	return nil
}

func runTodoDelete(ctx context.Context, cmd *TodoDeleteCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	client, err := newGraphClient(ctx, deps)
	if err != nil {
		return err
	}

	slog.Debug("deleting todo", "id", cmd.ID, "list", cfg.TodoListID)

	if err := client.DeleteTodo(ctx, cfg.TodoListID, cmd.ID); err != nil {
		return fmt.Errorf("deleting todo: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Task deleted")
	return nil
}

// --- Skill handlers ---

// skillTarget identifies where the skill file should be written.
type skillTarget string

const (
	skillTargetProject     skillTarget = "project"
	skillTargetClaudeUser  skillTarget = "claude-user"
	skillTargetCopilotUser skillTarget = "copilot-user"
	skillTargetFile        skillTarget = "file"
	skillTargetStdout      skillTarget = "stdout"
)

func runSkill(cmd *SkillCmd, deps *Dependencies) error {
	if cmd.Install != nil {
		return runSkillInstall(cmd.Install, deps)
	}

	return fmt.Errorf("missing subcommand — run 'graphdo skill --help' for usage")
}

func runSkillInstall(cmd *SkillInstallCmd, deps *Dependencies) error {
	if deps.SkillContent == "" {
		return fmt.Errorf("skill content not available")
	}

	target, outputPath, err := resolveSkillTarget(cmd)
	if err != nil {
		return err
	}

	switch target {
	case skillTargetStdout:
		_, _ = fmt.Fprint(deps.Stdout, deps.SkillContent)
		return nil
	case skillTargetFile:
		return writeSkillFile(outputPath, deps.SkillContent, deps.Stderr)
	case skillTargetProject:
		return writeSkillFile(".agents/skills/graphdo/SKILL.md", deps.SkillContent, deps.Stderr)
	case skillTargetClaudeUser:
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		return writeSkillFile(filepath.Join(home, ".claude", "skills", "graphdo", "SKILL.md"), deps.SkillContent, deps.Stderr)
	case skillTargetCopilotUser:
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		return writeSkillFile(filepath.Join(home, ".copilot", "skills", "graphdo", "SKILL.md"), deps.SkillContent, deps.Stderr)
	default:
		return fmt.Errorf("unknown skill target: %s", target)
	}
}

// resolveSkillTarget determines the target from flags or runs an interactive picker.
func resolveSkillTarget(cmd *SkillInstallCmd) (skillTarget, string, error) {
	// Explicit stdout flag.
	if cmd.Stdout {
		return skillTargetStdout, "", nil
	}

	// Explicit file output.
	if cmd.Output != "" {
		return skillTargetFile, cmd.Output, nil
	}

	// Flag-based (non-interactive) resolution.
	if cmd.Agent != "" || cmd.Scope != "" {
		return resolveSkillTargetFromFlags(cmd.Agent, cmd.Scope)
	}

	// Interactive picker.
	return resolveSkillTargetInteractive()
}

func resolveSkillTargetFromFlags(agent, scope string) (skillTarget, string, error) {
	agent = strings.ToLower(agent)
	scope = strings.ToLower(scope)

	if scope == "" {
		return "", "", fmt.Errorf("--scope is required when --agent is specified (use 'project' or 'user')")
	}
	if agent == "" {
		return "", "", fmt.Errorf("--agent is required when --scope is specified (use 'claude' or 'copilot')")
	}

	switch {
	case scope == "project":
		return skillTargetProject, "", nil
	case agent == "claude" && scope == "user":
		return skillTargetClaudeUser, "", nil
	case agent == "copilot" && scope == "user":
		return skillTargetCopilotUser, "", nil
	default:
		return "", "", fmt.Errorf("unsupported combination: --agent %q --scope %q", agent, scope)
	}
}

func resolveSkillTargetInteractive() (skillTarget, string, error) {
	options := []huh.Option[skillTarget]{
		huh.NewOption("Current project (.agents/skills/graphdo/SKILL.md)", skillTargetProject),
		huh.NewOption("Claude Code — user profile (~/.claude/skills/graphdo/SKILL.md)", skillTargetClaudeUser),
		huh.NewOption("GitHub Copilot — user profile (~/.copilot/skills/graphdo/SKILL.md)", skillTargetCopilotUser),
		huh.NewOption("Print to stdout", skillTargetStdout),
	}

	var selected skillTarget
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[skillTarget]().
				Title("Where should the skill file be installed?").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", "", fmt.Errorf("running skill target picker: %w", err)
	}

	return selected, "", nil
}

func writeSkillFile(path, content string, stderr io.Writer) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing skill file: %w", err)
	}

	_, _ = fmt.Fprintf(stderr, "✓ Skill file installed to %s\n", path)
	return nil
}
