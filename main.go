// graphdo sends emails to yourself and manages todos via Microsoft Graph.
package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/co-native-ab/graphdo/internal/auth"
	"github.com/co-native-ab/graphdo/internal/cmd"
	"github.com/co-native-ab/graphdo/internal/config"
	"github.com/co-native-ab/graphdo/internal/graph"

	"github.com/alexflint/go-arg"
)

// version is set at build time via -ldflags.
var version = "dev"

//go:embed SKILL.md
var skillContent string

func main() {
	cmd.Version = version
	os.Exit(mainRun())
}

func mainRun() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func run(ctx context.Context, args []string) error {
	var cliArgs cmd.Args
	p, err := parseArgs(&cliArgs, args, os.Stderr)
	if err != nil {
		return err
	}
	if p == nil {
		return nil // help/version was printed
	}

	setupLogger(cliArgs.Debug)

	// Resolve config dir early so auth cache files go to the right place.
	configDir, err := config.Dir(cliArgs.ConfigDir)
	if err != nil {
		return fmt.Errorf("resolving config directory: %w", err)
	}

	// Resolve client ID from config for the authenticator.
	cfg, err := config.Load(configDir)
	if err != nil {
		slog.Debug("could not load config for client ID resolution", "error", err)
		cfg = &config.Config{}
	}
	clientID := config.ResolveClientID("", cfg.ClientID, cmd.DefaultClientID)
	authenticator := newAuthenticator(cliArgs.AccessToken, clientID, configDir, cliArgs.DeviceCode)

	deps := &cmd.Dependencies{
		Authenticator: authenticator,
		GraphClient:   graph.NewClient(cliArgs.GraphURL, authenticator),
		ConfigDir:     configDir,
		SkillContent:  skillContent,
		Stdout:        os.Stdout,
		Stderr:        os.Stderr,
		Stdin:         os.Stdin,
	}

	return cmd.Dispatch(ctx, p, &cliArgs, deps)
}

// parseArgs parses CLI arguments into cliArgs. Returns the parser for dispatch
// use, or nil if help/version was printed (not an error).
func parseArgs(cliArgs *cmd.Args, args []string, stderr io.Writer) (*arg.Parser, error) {
	p, err := arg.NewParser(arg.Config{}, cliArgs)
	if err != nil {
		return nil, fmt.Errorf("creating parser: %w", err)
	}

	if err := p.Parse(args); err != nil {
		switch {
		case errors.Is(err, arg.ErrHelp):
			p.WriteHelp(stderr)
			return nil, nil
		case errors.Is(err, arg.ErrVersion):
			_, _ = fmt.Fprintln(stderr, cliArgs.Version())
			return nil, nil
		default:
			p.WriteUsage(stderr)
			return nil, fmt.Errorf("parsing arguments: %w", err)
		}
	}

	return p, nil
}

// newAuthenticator returns a StaticAuthenticator if a token is provided,
// a DeviceCodeAuthenticator if --device-code is set, or a BrowserAuthenticator by default.
func newAuthenticator(accessToken, clientID, configDir string, deviceCode bool) auth.Authenticator {
	if accessToken != "" {
		return auth.NewStaticAuthenticator(accessToken)
	}
	if deviceCode {
		return auth.NewDeviceCodeAuthenticator(clientID, configDir)
	}
	return auth.NewBrowserAuthenticator(clientID, configDir)
}

func setupLogger(debug bool) {
	level := slog.LevelWarn
	if debug {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
