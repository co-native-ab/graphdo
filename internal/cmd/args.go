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

// LoginCmd is the argument type for the login subcommand.
type LoginCmd struct {
	ClientID string `arg:"--client-id,env:GRAPHDO_CLIENT_ID" help:"Azure AD app client ID (saved to config)"`
}

// ConfigCmd is the argument type for the config subcommand.
type ConfigCmd struct {
	Show *ConfigShowCmd `arg:"subcommand:show" help:"show current configuration"`
}

// ConfigShowCmd is the argument type for the config show subcommand.
type ConfigShowCmd struct{}

// MailCmd is the argument type for the mail subcommand.
type MailCmd struct {
	Send *MailSendCmd `arg:"subcommand:send" help:"send an email to yourself"`
}

// MailSendCmd is the argument type for the mail send subcommand.
type MailSendCmd struct {
	Subject string `arg:"--subject,required" help:"email subject"`
	Body    string `arg:"--body,required" help:"email body (use - for stdin)"`
	HTML    bool   `arg:"--html" help:"send body as HTML"`
}

// TodoCmd is the argument type for the todo subcommand.
type TodoCmd struct {
	List     *TodoListCmd     `arg:"subcommand:list" help:"list todos"`
	Show     *TodoShowCmd     `arg:"subcommand:show" help:"show a single todo"`
	Create   *TodoCreateCmd   `arg:"subcommand:create" help:"create a todo"`
	Update   *TodoUpdateCmd   `arg:"subcommand:update" help:"update a todo"`
	Complete *TodoCompleteCmd `arg:"subcommand:complete" help:"mark a todo as completed"`
	Delete   *TodoDeleteCmd   `arg:"subcommand:delete" help:"delete a todo"`
}

// TodoListCmd is the argument type for the todo list subcommand.
type TodoListCmd struct {
	Top  int `arg:"--top" help:"maximum number of items to return" default:"20"`
	Skip int `arg:"--skip" help:"number of items to skip (for pagination)" default:"0"`
}

// TodoShowCmd is the argument type for the todo show subcommand.
type TodoShowCmd struct {
	ID string `arg:"--id,required" help:"task ID"`
}

// TodoUpdateCmd is the argument type for the todo update subcommand.
type TodoUpdateCmd struct {
	ID    string `arg:"--id,required" help:"task ID"`
	Title string `arg:"--title" help:"new task title"`
	Body  string `arg:"--body" help:"new task body"`
}

// TodoCreateCmd is the argument type for the todo create subcommand.
type TodoCreateCmd struct {
	Title string `arg:"--title,required" help:"task title"`
	Body  string `arg:"--body" help:"task body"`
}

// TodoCompleteCmd is the argument type for the todo complete subcommand.
type TodoCompleteCmd struct {
	ID string `arg:"--id,required" help:"task ID"`
}

// TodoDeleteCmd is the argument type for the todo delete subcommand.
type TodoDeleteCmd struct {
	ID string `arg:"--id,required" help:"task ID"`
}

// LogoutCmd is the argument type for the logout subcommand.
type LogoutCmd struct{}

// StatusCmd is the argument type for the status subcommand.
type StatusCmd struct{}

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
