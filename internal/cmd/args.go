// Package cmd defines the CLI argument types and command handlers for graphdo.
package cmd

// DefaultClientID is the built-in Azure AD application (client) ID used when
// neither the --client-id flag nor the config file provides one.
const DefaultClientID = "b073490b-a1a2-4bb8-9d83-00bb5c15fcfd"

// Version is set by main.go from the build-time ldflags value.
var Version = "dev"

// Args defines the top-level CLI arguments and subcommands.
type Args struct {
	Login  *LoginCmd  `arg:"subcommand:login"  help:"sign in with your Microsoft account"`
	Logout *LogoutCmd `arg:"subcommand:logout" help:"clear cached credentials"`
	Status *StatusCmd `arg:"subcommand:status" help:"check if graphdo is ready to use"`
	Config *ConfigCmd `arg:"subcommand:config" help:"configure graphdo (select todo list)"`
	Mail   *MailCmd   `arg:"subcommand:mail"   help:"send an email to yourself"`
	Todo   *TodoCmd   `arg:"subcommand:todo"   help:"manage your todo items"`
	Skill  *SkillCmd  `arg:"subcommand:skill"  help:"manage the graphdo agent skill file"`
	Mcp    *McpCmd    `arg:"subcommand:mcp"    help:"run or install the MCP server"`

	GraphURL    string `arg:"--graph-url,env:GRAPHDO_GRAPH_URL" help:"Microsoft Graph API base URL" default:"https://graph.microsoft.com/v1.0"`
	AccessToken string `arg:"--access-token,env:GRAPHDO_ACCESS_TOKEN" help:"access token (skips login)"`
	ConfigDir   string `arg:"--config-dir,env:GRAPHDO_CONFIG_DIR" help:"config directory path"`
	DeviceCode  bool   `arg:"--device-code" help:"use device code flow instead of browser"`
	Debug       bool   `arg:"--debug" help:"enable debug logging"`
}

// Description returns a one-line description of the CLI for help output.
func (Args) Description() string {
	return "graphdo — Send emails to yourself and manage todos via Microsoft Graph"
}

// Version returns the version string for --version output.
func (Args) Version() string {
	return "graphdo " + Version
}
