package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

// --- MCP tool input types ---

type mcpMailSendInput struct {
	Subject string `json:"subject" jsonschema:"the email subject"`
	Body    string `json:"body" jsonschema:"the email body"`
	HTML    bool   `json:"html,omitempty" jsonschema:"send body as HTML (default false)"`
}

// --- CLI handlers ---

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

	user, err := deps.GraphClient.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("getting user profile: %w", err)
	}

	email := user.Mail
	if email == "" {
		email = user.UserPrincipalName
	}

	slog.Debug("sending email", "to", email, "subject", cmd.Subject, "html", cmd.HTML)

	if err := deps.GraphClient.SendMail(ctx, email, cmd.Subject, body, cmd.HTML); err != nil {
		return fmt.Errorf("sending mail: %w", err)
	}

	_, _ = fmt.Fprintf(deps.Stderr, "✓ Email sent to %s\n", email)
	return nil
}

// --- MCP tool registration ---

func registerMailMCPTools(s *mcp.Server, deps *Dependencies) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mail_send",
		Description: "Send an email to yourself via Microsoft Graph",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input mcpMailSendInput) (*mcp.CallToolResult, any, error) {
		user, err := deps.GraphClient.GetMe(ctx)
		if err != nil {
			return mcpErrResult(fmt.Errorf("getting user profile: %w", err)), nil, nil
		}
		email := user.Mail
		if email == "" {
			email = user.UserPrincipalName
		}
		if err := deps.GraphClient.SendMail(ctx, email, input.Subject, input.Body, input.HTML); err != nil {
			return mcpErrResult(fmt.Errorf("sending mail: %w", err)), nil, nil
		}
		return mcpTextResult(fmt.Sprintf("✓ Email sent to %s", email)), nil, nil
	})
}
