package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/huh"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// McpCmd is the argument type for the mcp subcommand.
type McpCmd struct {
	Run     *McpRunCmd     `arg:"subcommand:run"     help:"start the graphdo MCP stdio server"`
	Install *McpInstallCmd `arg:"subcommand:install" help:"install graphdo as an MCP server"`
}

// McpRunCmd is the argument type for the mcp run subcommand.
type McpRunCmd struct{}

// McpInstallCmd is the argument type for the mcp install subcommand.
type McpInstallCmd struct {
	Target string `arg:"--target" help:"install target: claude-code, claude-desktop, vscode, copilot"`
}

// --- Handler dispatch ---

func runMcp(ctx context.Context, cmd *McpCmd, deps *Dependencies) error {
	switch {
	case cmd.Run != nil:
		return runMcpRun(ctx, deps)
	case cmd.Install != nil:
		return runMcpInstall(cmd.Install, deps)
	default:
		return fmt.Errorf("missing subcommand — run 'graphdo mcp --help' for usage")
	}
}

// --- MCP server ---

func runMcpRun(ctx context.Context, deps *Dependencies) error {
	s := BuildMCPServer(deps)
	transport := &mcp.IOTransport{
		Reader: io.NopCloser(deps.Stdin),
		Writer: writeNopCloser{deps.Stdout},
	}
	return s.Run(ctx, transport)
}

// writeNopCloser wraps an io.Writer with a no-op Close method.
type writeNopCloser struct{ io.Writer }

func (writeNopCloser) Close() error { return nil }

// BuildMCPServer creates the MCP server and registers all tools. Exported for testing.
func BuildMCPServer(deps *Dependencies) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "graphdo", Version: Version}, nil)
	registerMailMCPTools(s, deps)
	registerTodoMCPTools(s, deps)
	return s
}

// --- MCP result helpers ---

func mcpTextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func mcpJSONResult(v any) *mcp.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return mcpErrResult(fmt.Errorf("encoding result: %w", err))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}

func mcpErrResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}

// --- MCP install ---

type mcpInstallTarget string

const (
	mcpTargetClaudeCode    mcpInstallTarget = "claude-code"
	mcpTargetClaudeDesktop mcpInstallTarget = "claude-desktop"
	mcpTargetVSCode        mcpInstallTarget = "vscode"
	mcpTargetCopilot       mcpInstallTarget = "copilot"
)

// mcpServerEntry is the MCP server config entry written to config files.
type mcpServerEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func runMcpInstall(cmd *McpInstallCmd, deps *Dependencies) error {
	target, err := resolveMcpTarget(cmd.Target)
	if err != nil {
		return err
	}

	execPath := resolvedExecutable()

	entry := mcpServerEntry{
		Command: execPath,
		Args:    []string{"mcp", "run"},
	}

	configPath, jsonKey, err := mcpConfigPath(target)
	if err != nil {
		return err
	}

	if err := writeMcpConfig(configPath, jsonKey, entry); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(deps.Stderr, "✓ MCP server installed to %s\n", configPath)
	return nil
}

func resolveMcpTarget(target string) (mcpInstallTarget, error) {
	if target != "" {
		t := mcpInstallTarget(target)
		switch t {
		case mcpTargetClaudeCode, mcpTargetClaudeDesktop, mcpTargetVSCode, mcpTargetCopilot:
			return t, nil
		default:
			return "", fmt.Errorf("unknown target %q — use: claude-code, claude-desktop, vscode, copilot", target)
		}
	}
	return resolveMcpTargetInteractive()
}

func resolveMcpTargetInteractive() (mcpInstallTarget, error) {
	options := []huh.Option[mcpInstallTarget]{
		huh.NewOption("Claude Code (~/.claude.json)", mcpTargetClaudeCode),
		huh.NewOption("Claude Desktop (OS-specific config file)", mcpTargetClaudeDesktop),
		huh.NewOption("VS Code — workspace (.vscode/mcp.json)", mcpTargetVSCode),
		huh.NewOption("GitHub Copilot CLI (~/.copilot/mcp.json)", mcpTargetCopilot),
	}

	var selected mcpInstallTarget
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[mcpInstallTarget]().
				Title("Where should the MCP server be installed?").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("running MCP target picker: %w", err)
	}

	return selected, nil
}

// mcpConfigPath returns (configFilePath, jsonTopLevelKey) for the given target.
func mcpConfigPath(target mcpInstallTarget) (path, jsonKey string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("getting home directory: %w", err)
	}

	switch target {
	case mcpTargetClaudeCode:
		return filepath.Join(home, ".claude.json"), "mcpServers", nil
	case mcpTargetClaudeDesktop:
		p, err := claudeDesktopConfigPath(home)
		return p, "mcpServers", err
	case mcpTargetVSCode:
		return filepath.Join(".vscode", "mcp.json"), "servers", nil
	case mcpTargetCopilot:
		return filepath.Join(home, ".copilot", "mcp.json"), "servers", nil
	default:
		return "", "", fmt.Errorf("unknown target: %s", target)
	}
}

func claudeDesktopConfigPath(home string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	default:
		return filepath.Join(home, ".config", "claude", "claude_desktop_config.json"), nil
	}
}

// writeMcpConfig reads the existing JSON config at path, upserts graphdo under
// cfg[jsonKey]["graphdo"], and writes back atomically.
func writeMcpConfig(path, jsonKey string, entry mcpServerEntry) error {
	path = filepath.Clean(path)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Read existing config (start with empty map if file doesn't exist).
	existing := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("parsing existing config %s: %w", path, err)
		}
	}

	// Upsert the graphdo server entry under the target key.
	servers, _ := existing[jsonKey].(map[string]any)
	if servers == nil {
		servers = make(map[string]any)
	}
	servers["graphdo"] = entry
	existing[jsonKey] = servers

	// Write atomically via temp file + rename.
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, ".mcp-config-*.json")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("installing config file: %w", err)
	}

	return nil
}

// resolvedExecutable returns the path to the current binary, resolving symlinks.
// Falls back gracefully if the path cannot be determined.
func resolvedExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return "graphdo"
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}
	return resolved
}
