package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
)

// SkillCmd is the argument type for the skill subcommand.
type SkillCmd struct {
	Install *SkillInstallCmd `arg:"subcommand:install" help:"install the graphdo agent skill file"`
}

// SkillInstallCmd is the argument type for the skill install subcommand.
type SkillInstallCmd struct {
	Agent  string `arg:"--agent" help:"agent type: claude or copilot"`
	Scope  string `arg:"--scope" help:"installation scope: project or user"`
	Output string `arg:"--output" help:"write skill file to this path"`
	Stdout bool   `arg:"--stdout" help:"print skill file to stdout"`
}

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
