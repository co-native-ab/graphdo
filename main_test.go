package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/co-native-ab/graphdo/internal/cmd"
	"github.com/co-native-ab/graphdo/internal/config"
	"github.com/co-native-ab/graphdo/internal/graph"
	"github.com/co-native-ab/graphdo/internal/testutil"
)

// testEnv bundles the mock server, temp config dir, and captured output
// for a single integration test.
type testEnv struct {
	state     *testutil.MockState
	graphURL  string
	configDir string
	stdout    *bytes.Buffer
	stderr    *bytes.Buffer
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	state := testutil.NewMockState()
	state.User = graph.User{
		DisplayName:       "Test User",
		Mail:              "test@example.com",
		UserPrincipalName: "test@example.com",
	}
	state.TodoLists = []graph.TodoList{
		{ID: "list-1", DisplayName: "My Tasks"},
	}
	state.Todos["list-1"] = []graph.TodoItem{
		{ID: "task-1", Title: "Buy milk", Status: "notStarted"},
	}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	configDir := t.TempDir()

	return &testEnv{
		state:     state,
		graphURL:  srv.URL,
		configDir: configDir,
		stdout:    &bytes.Buffer{},
		stderr:    &bytes.Buffer{},
	}
}

// seedConfig writes a config file pointing at list-1.
func (e *testEnv) seedConfig(t *testing.T) {
	t.Helper()
	cfg := &config.Config{
		ClientID:     cmd.DefaultClientID,
		TodoListID:   "list-1",
		TodoListName: "My Tasks",
	}
	if err := config.Save(cfg, e.configDir); err != nil {
		t.Fatalf("seeding config: %v", err)
	}
}

// runCLI executes run() with the given arguments plus standard test flags
// (--access-token, --graph-url, --config-dir).
func (e *testEnv) runCLI(t *testing.T, args ...string) error {
	t.Helper()
	fullArgs := append([]string{
		"--access-token", "test-token",
		"--graph-url", e.graphURL,
		"--config-dir", e.configDir,
	}, args...)
	return runWithIO(context.Background(), fullArgs, e.stdout, e.stderr)
}

// runWithIO is like run() but injects custom stdout/stderr.
func runWithIO(ctx context.Context, args []string, stdout, stderr *bytes.Buffer) error {
	var cliArgs cmd.Args
	p, err := parseArgs(&cliArgs, args, stderr)
	if err != nil {
		return err
	}
	if p == nil {
		return nil // help/version was printed
	}

	setupLogger(cliArgs.Debug)

	// Resolve config dir so auth cache and config share the same path.
	configDir, err := config.Dir(cliArgs.ConfigDir)
	if err != nil {
		return fmt.Errorf("resolving config directory: %w", err)
	}

	cfg, _ := config.Load(configDir)
	if cfg == nil {
		cfg = &config.Config{}
	}
	clientID := config.ResolveClientID("", cfg.ClientID, cmd.DefaultClientID)
	authenticator := newAuthenticator(cliArgs.AccessToken, clientID, configDir, cliArgs.DeviceCode)

	deps := &cmd.Dependencies{
		Authenticator: authenticator,
		GraphURL:      cliArgs.GraphURL,
		ConfigDir:     configDir,
		SkillContent:  skillContent,
		Stdout:        stdout,
		Stderr:        stderr,
		Stdin:         strings.NewReader(""),
	}

	return cmd.Dispatch(ctx, p, &cliArgs, deps)
}

func TestHelp(t *testing.T) {
	env := newTestEnv(t)
	err := env.runCLI(t, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(env.stderr.String(), "graphdo") {
		t.Errorf("help output missing 'graphdo', got: %s", env.stderr.String())
	}
}

func TestLogin(t *testing.T) {
	env := newTestEnv(t)
	err := env.runCLI(t, "login")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(env.stderr.String(), "Logged in") {
		t.Errorf("expected login success message, got: %s", env.stderr.String())
	}
}

func TestConfigShow(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "config", "show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(env.stdout.Bytes(), &cfg); err != nil {
		t.Fatalf("decoding config output: %v", err)
	}
	if cfg.TodoListID != "list-1" {
		t.Errorf("got TodoListID %q, want %q", cfg.TodoListID, "list-1")
	}
	if cfg.ClientID != cmd.DefaultClientID {
		t.Errorf("got ClientID %q, want %q", cfg.ClientID, cmd.DefaultClientID)
	}
}

func TestMailSend(t *testing.T) {
	env := newTestEnv(t)
	err := env.runCLI(t, "mail", "send", "--subject", "Hello", "--body", "World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mails := env.state.GetSentMails()
	if len(mails) != 1 {
		t.Fatalf("expected 1 mail, got %d", len(mails))
	}
	if mails[0].Subject != "Hello" {
		t.Errorf("got subject %q, want %q", mails[0].Subject, "Hello")
	}
	if mails[0].Body != "World" {
		t.Errorf("got body %q, want %q", mails[0].Body, "World")
	}
	if mails[0].ContentType != "Text" {
		t.Errorf("got content type %q, want %q", mails[0].ContentType, "Text")
	}
	if mails[0].To != "test@example.com" {
		t.Errorf("got to %q, want %q", mails[0].To, "test@example.com")
	}
}

func TestMailSendHTML(t *testing.T) {
	env := newTestEnv(t)
	err := env.runCLI(t, "mail", "send", "--subject", "Hello", "--body", "<h1>World</h1>", "--html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mails := env.state.GetSentMails()
	if len(mails) != 1 {
		t.Fatalf("expected 1 mail, got %d", len(mails))
	}
	if mails[0].ContentType != "HTML" {
		t.Errorf("got content type %q, want %q", mails[0].ContentType, "HTML")
	}
}

func TestTodoList(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []graph.TodoItem
	if err := json.Unmarshal(env.stdout.Bytes(), &items); err != nil {
		t.Fatalf("decoding todo list output: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(items))
	}
	if items[0].Title != "Buy milk" {
		t.Errorf("got title %q, want %q", items[0].Title, "Buy milk")
	}
}

func TestTodoCreate(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "create", "--title", "New task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := env.state.GetTodos("list-1")
	if len(items) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(items))
	}

	found := false
	for _, item := range items {
		if item.Title == "New task" {
			found = true
			break
		}
	}
	if !found {
		t.Error("created todo not found in mock state")
	}
}

func TestTodoComplete(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "complete", "--id", "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := env.state.GetTodos("list-1")
	if len(items) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(items))
	}
	if items[0].Status != "completed" {
		t.Errorf("got status %q, want %q", items[0].Status, "completed")
	}

	if !strings.Contains(env.stderr.String(), "completed") {
		t.Errorf("expected completion message, got: %s", env.stderr.String())
	}
}

func TestTodoDelete(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "delete", "--id", "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := env.state.GetTodos("list-1")
	if len(items) != 0 {
		t.Fatalf("expected 0 todos, got %d", len(items))
	}

	if !strings.Contains(env.stderr.String(), "deleted") {
		t.Errorf("expected deletion message, got: %s", env.stderr.String())
	}
}

func TestMissingConfig(t *testing.T) {
	env := newTestEnv(t)
	// Don't seed config — should fail with helpful message.

	err := env.runCLI(t, "todo", "list")
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !strings.Contains(err.Error(), "config invalid") {
		t.Errorf("error should mention config invalid, got: %v", err)
	}
}

func TestNoSubcommand(t *testing.T) {
	env := newTestEnv(t)
	err := env.runCLI(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should print help to stderr.
	if !strings.Contains(env.stderr.String(), "graphdo") {
		t.Errorf("expected help output on no subcommand, got: %s", env.stderr.String())
	}
}

func TestConfigShowEmpty(t *testing.T) {
	env := newTestEnv(t)
	// No config seeded — should still show empty config.

	err := env.runCLI(t, "config", "show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(env.stdout.Bytes(), &cfg); err != nil {
		t.Fatalf("decoding config output: %v", err)
	}
	if cfg.TodoListID != "" {
		t.Errorf("expected empty todo_list_id, got %q", cfg.TodoListID)
	}
}

func TestTodoListPagination(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)
	// Add more todos to test pagination
	env.state.Todos["list-1"] = []graph.TodoItem{
		{ID: "t1", Title: "Task 1", Status: "notStarted"},
		{ID: "t2", Title: "Task 2", Status: "notStarted"},
		{ID: "t3", Title: "Task 3", Status: "notStarted"},
		{ID: "t4", Title: "Task 4", Status: "notStarted"},
		{ID: "t5", Title: "Task 5", Status: "notStarted"},
	}

	// First page
	err := env.runCLI(t, "todo", "list", "--top", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var items []graph.TodoItem
	if err := json.Unmarshal(env.stdout.Bytes(), &items); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "t1" || items[1].ID != "t2" {
		t.Errorf("wrong items: %v", items)
	}

	// Second page
	env.stdout.Reset()
	err = env.runCLI(t, "todo", "list", "--top", "2", "--skip", "2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := json.Unmarshal(env.stdout.Bytes(), &items); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "t3" || items[1].ID != "t4" {
		t.Errorf("wrong items: %v", items)
	}

	// Last page
	env.stdout.Reset()
	err = env.runCLI(t, "todo", "list", "--top", "2", "--skip", "4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := json.Unmarshal(env.stdout.Bytes(), &items); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestTodoShow(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "show", "--id", "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var item graph.TodoItem
	if err := json.Unmarshal(env.stdout.Bytes(), &item); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if item.ID != "task-1" {
		t.Errorf("got ID %q, want %q", item.ID, "task-1")
	}
	if item.Title != "Buy milk" {
		t.Errorf("got title %q, want %q", item.Title, "Buy milk")
	}
}

func TestTodoUpdate(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "update", "--id", "task-1", "--title", "Buy oat milk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output
	var item graph.TodoItem
	if err := json.Unmarshal(env.stdout.Bytes(), &item); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if item.Title != "Buy oat milk" {
		t.Errorf("got title %q, want %q", item.Title, "Buy oat milk")
	}

	// Verify state was updated
	items := env.state.GetTodos("list-1")
	if len(items) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(items))
	}
	if items[0].Title != "Buy oat milk" {
		t.Errorf("state title %q, want %q", items[0].Title, "Buy oat milk")
	}

	// Verify stderr
	if !strings.Contains(env.stderr.String(), "updated") {
		t.Errorf("expected update message, got: %s", env.stderr.String())
	}
}

func TestTodoUpdateBody(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "update", "--id", "task-1", "--body", "2% milk from corner store")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var item graph.TodoItem
	if err := json.Unmarshal(env.stdout.Bytes(), &item); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if item.Body == nil || item.Body.Content != "2% milk from corner store" {
		t.Errorf("unexpected body: %+v", item.Body)
	}
}

func TestTodoUpdateRequiresField(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "todo", "update", "--id", "task-1")
	if err == nil {
		t.Fatal("expected error when no --title or --body provided")
	}
	if !strings.Contains(err.Error(), "at least one of --title or --body") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConfigDir(t *testing.T) {
	env := newTestEnv(t)

	cfg := &config.Config{TodoListID: "custom-list", TodoListName: "Custom"}
	if err := config.Save(cfg, env.configDir); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	// Verify the config file exists in the expected location.
	expectedPath := filepath.Join(env.configDir, "config.json")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("config file not at expected path: %v", err)
	}
}

func TestLoginSavesClientID(t *testing.T) {
	env := newTestEnv(t)

	err := env.runCLI(t, "login")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Login should have saved the default client ID to config.
	cfg, err := config.Load(env.configDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	if cfg.ClientID != cmd.DefaultClientID {
		t.Errorf("got ClientID %q, want %q", cfg.ClientID, cmd.DefaultClientID)
	}
}

func TestLoginCustomClientID(t *testing.T) {
	env := newTestEnv(t)

	err := env.runCLI(t, "login", "--client-id", "custom-app-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := config.Load(env.configDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	if cfg.ClientID != "custom-app-id" {
		t.Errorf("got ClientID %q, want %q", cfg.ClientID, "custom-app-id")
	}
}

func TestLoginPreservesExistingConfig(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "login")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Login should preserve the todo list config.
	cfg, err := config.Load(env.configDir)
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	if cfg.TodoListID != "list-1" {
		t.Errorf("login overwrote TodoListID: got %q, want %q", cfg.TodoListID, "list-1")
	}
	if cfg.TodoListName != "My Tasks" {
		t.Errorf("login overwrote TodoListName: got %q, want %q", cfg.TodoListName, "My Tasks")
	}
}

func TestLogout(t *testing.T) {
	env := newTestEnv(t)

	// Create fake cache files.
	cacheFile := filepath.Join(env.configDir, "msal_cache.json")
	accountFile := filepath.Join(env.configDir, "account.json")
	if err := os.WriteFile(cacheFile, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("writing cache file: %v", err)
	}
	if err := os.WriteFile(accountFile, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("writing account file: %v", err)
	}

	err := env.runCLI(t, "logout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(env.stderr.String(), "Logged out") {
		t.Errorf("expected logout success message, got: %s", env.stderr.String())
	}

	// Verify files were removed.
	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		t.Error("msal_cache.json should have been removed")
	}
	if _, err := os.Stat(accountFile); !os.IsNotExist(err) {
		t.Error("account.json should have been removed")
	}
}

func TestLogoutNoCacheFiles(t *testing.T) {
	env := newTestEnv(t)

	// Logout when no cache files exist should not error.
	err := env.runCLI(t, "logout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(env.stderr.String(), "Logged out") {
		t.Errorf("expected logout success message, got: %s", env.stderr.String())
	}
}

func TestStatusReady(t *testing.T) {
	env := newTestEnv(t)
	env.seedConfig(t)

	err := env.runCLI(t, "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Ready     bool   `json:"ready"`
		LoggedIn  bool   `json:"logged_in"`
		User      string `json:"user"`
		TodoList  string `json:"todo_list"`
		TodoCount int    `json:"todo_count"`
	}
	if err := json.Unmarshal(env.stdout.Bytes(), &result); err != nil {
		t.Fatalf("decoding status: %v", err)
	}
	if !result.Ready {
		t.Error("expected ready=true")
	}
	if !result.LoggedIn {
		t.Error("expected logged_in=true")
	}
	if result.User != "test@example.com" {
		t.Errorf("got user %q, want %q", result.User, "test@example.com")
	}
	if result.TodoList != "My Tasks" {
		t.Errorf("got todo_list %q, want %q", result.TodoList, "My Tasks")
	}
	if result.TodoCount != 1 {
		t.Errorf("got todo_count %d, want 1", result.TodoCount)
	}
}

func TestStatusNoConfig(t *testing.T) {
	env := newTestEnv(t)
	// No config seeded — should report not configured.

	err := env.runCLI(t, "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Ready bool   `json:"ready"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(env.stdout.Bytes(), &result); err != nil {
		t.Fatalf("decoding status: %v", err)
	}
	if result.Ready {
		t.Error("expected ready=false")
	}
	if !strings.Contains(result.Error, "not configured") {
		t.Errorf("expected config error, got: %q", result.Error)
	}
}

func TestSkillInstallStdout(t *testing.T) {
	env := newTestEnv(t)

	err := env.runCLI(t, "skill", "install", "--stdout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := env.stdout.String()
	if !strings.Contains(output, "graphdo") {
		t.Errorf("expected skill content on stdout, got: %s", output[:min(len(output), 100)])
	}
	if !strings.Contains(output, "# graphdo CLI") {
		t.Errorf("expected SKILL.md header, got: %s", output[:min(len(output), 100)])
	}
}

func TestSkillInstallFile(t *testing.T) {
	env := newTestEnv(t)
	outPath := filepath.Join(t.TempDir(), "skill.md")

	err := env.runCLI(t, "skill", "install", "--output", outPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	if !strings.Contains(string(data), "graphdo") {
		t.Errorf("expected skill content in file, got: %s", string(data)[:min(len(data), 100)])
	}

	if !strings.Contains(env.stderr.String(), "installed") {
		t.Errorf("expected install confirmation, got: %s", env.stderr.String())
	}
}

func TestSkillInstallProject(t *testing.T) {
	env := newTestEnv(t)

	// Change to temp dir so the project path is predictable.
	projectDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("changing to project dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	err = env.runCLI(t, "skill", "install", "--agent", "claude", "--scope", "project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skillPath := filepath.Join(projectDir, ".agents", "skills", "graphdo", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("reading skill file: %v", err)
	}
	if !strings.Contains(string(data), "graphdo") {
		t.Errorf("expected skill content in project file")
	}
}

func TestSkillInstallCopilotProject(t *testing.T) {
	env := newTestEnv(t)

	projectDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("changing to project dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Copilot with project scope also uses .agents/skills/graphdo/SKILL.md
	err = env.runCLI(t, "skill", "install", "--agent", "copilot", "--scope", "project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skillPath := filepath.Join(projectDir, ".agents", "skills", "graphdo", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("skill file not found at expected path: %v", err)
	}
}
