package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/co-native-ab/graphdo/internal/auth"
	"github.com/co-native-ab/graphdo/internal/config"
	"github.com/co-native-ab/graphdo/internal/graph"

	"github.com/alexflint/go-arg"
)

// Dependencies holds shared resources injected into command handlers.
type Dependencies struct {
	Authenticator auth.Authenticator
	GraphClient   *graph.Client
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
	case cliArgs.Mcp != nil:
		return runMcp(ctx, cliArgs.Mcp, deps)
	default:
		p.WriteHelp(deps.Stderr)
		return nil
	}
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
