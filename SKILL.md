---
name: graphdo
description: Send emails and manage Microsoft To Do tasks via the graphdo CLI. Use this skill whenever the user wants to send themselves an email summary, reminder, or report, create or manage tasks in Microsoft To Do, do daily planning or wrap-ups involving email and task management, or work with their Microsoft 365 productivity tools from the command line. Also use it when the user mentions graphdo by name, or asks about emailing themselves or managing their todo list programmatically.
---

# graphdo CLI

Send emails to yourself and manage Microsoft To Do tasks via Microsoft Graph.

All machine-readable output is JSON on stdout. Human messages go to stderr. Parse stdout only.

## MCP Server

graphdo includes a built-in MCP (Model Context Protocol) server. Use this when graphdo is registered as an MCP server with your AI tool (Claude Code, Claude Desktop, VS Code, GitHub Copilot CLI). In isolated environments without shell access (e.g., Claude Cowork's VM), MCP is required.

### Setup

If the MCP server is not already configured, tell the user to run:

```bash
graphdo mcp install
```

This is an interactive command (requires human input). It will ask which AI tool to configure. Supported targets: Claude Code, Claude Desktop, VS Code, and GitHub Copilot CLI.

For non-interactive use:

```bash
graphdo mcp install --target claude-code
graphdo mcp install --target claude-desktop
graphdo mcp install --target vscode
graphdo mcp install --target copilot
```

Once installed, the MCP server starts automatically when the AI tool invokes it. The server provides these tools:

| MCP Tool | Description |
|----------|-------------|
| `mail_send` | Send an email to yourself (params: `subject`, `body`, optional `html`) |
| `todo_list` | List todos (optional params: `top`, `skip`) |
| `todo_show` | Get a single todo (param: `id`) |
| `todo_create` | Create a todo (params: `title`, optional `body`) |
| `todo_update` | Update title/body (params: `id`, optional `title`, optional `body` — at least one required) |
| `todo_complete` | Mark todo as completed (param: `id`) |
| `todo_delete` | Delete a todo (param: `id`) |

### MCP Prerequisite

Before the MCP server can be used, graphdo must be authenticated and configured. Verify by running:

```bash
graphdo status
```

If `"ready": false`, tell the user to run `graphdo login` and `graphdo config` (both are interactive — do not run them from an agent).

## CLI Commands

Use these commands when you have direct shell access to the system where graphdo is installed (e.g., running locally with Claude Code, or any agent with terminal access).

## Before You Start

Verify the CLI is configured by running:

```bash
graphdo status
```

If it returns JSON with `"ready": true`, the CLI is fully configured. If `"ready": false`, read the `"error"` field — it will tell the user what to do (run `graphdo login` or `graphdo config`). Both are interactive and require human input. Do not run them yourself.

## Commands

### Send email (self only)

```bash
graphdo mail send --subject "SUBJECT" --body "BODY" [--html]
```

| Flag | Required | Description |
|------|----------|-------------|
| `--subject` | yes | Email subject line |
| `--body` | yes | Email body text. Use `-` to read from stdin |
| `--html` | no | Send body as HTML instead of plain text |

Stdout: nothing on success. Use `--body -` with stdin for long bodies to avoid shell escaping issues.

**Example:**
```bash
echo "$SUMMARY" | graphdo mail send --subject "Daily Summary — $(date +%Y-%m-%d)" --body -
```

### List todos

```bash
graphdo todo list [--top N] [--skip N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--top` | `20` | Maximum items to return |
| `--skip` | `0` | Items to skip (for pagination) |

Stdout: JSON array of todo objects.

**Example (paginated):**
```bash
graphdo todo list --top 10
graphdo todo list --top 10 --skip 10
```

### Show a single todo

```bash
graphdo todo show --id "ID"
```

Stdout: JSON object of the todo item.

### Create a todo

```bash
graphdo todo create --title "TITLE" [--body "BODY"]
```

Stdout: JSON object of the created todo (includes the `id` you need for subsequent operations).

### Update a todo

```bash
graphdo todo update --id "ID" [--title "TITLE"] [--body "BODY"]
```

At least one of `--title` or `--body` must be provided. Stdout: JSON object of the updated todo.

### Complete a todo

```bash
graphdo todo complete --id "ID"
```

### Delete a todo

```bash
graphdo todo delete --id "ID"
```

### Show configuration

```bash
graphdo config show
```

Stdout: JSON object with `client_id`, `todo_list_id`, `todo_list_name`.

### Check readiness

```bash
graphdo status
```

Stdout: JSON with `ready` (bool), `logged_in`, `user`, `todo_list`, `todo_count`, and `error` (if not ready).

## Interactive Commands (Human Only)

Do not run these from an agent — they require human interaction and will hang.

| Command | Purpose |
|---------|---------|
| `graphdo login` | Authenticate with Microsoft (browser flow) |
| `graphdo logout` | Clear cached credentials |
| `graphdo config` | Pick which todo list to use (interactive picker) |
| `graphdo skill install` | Install this skill file for an agent |
| `graphdo mcp install` | Install the MCP server for an AI tool |

## Output Format

Todo items have this shape:

```json
{
  "id": "AQMkAD...",
  "title": "Buy groceries",
  "status": "notStarted",
  "body": {"content": "Milk, eggs, bread", "contentType": "text"}
}
```

The `status` field is either `"notStarted"` or `"completed"`. The `id` is an opaque string from Microsoft Graph — never construct or modify IDs yourself; always use IDs from `todo list` or `todo create` output.

## Common Workflows

### Daily wrap-up

1. Summarize the user's emails and meetings (e.g., using the Claude Microsoft 365 Connector).
2. Send the summary as an email:
   ```bash
   echo "$SUMMARY" | graphdo mail send --subject "Daily Wrap-up — $(date +%Y-%m-%d)" --body -
   ```
3. Create follow-up tasks:
   ```bash
   graphdo todo create --title "Respond to Alice's proposal" --body "She needs feedback by Friday"
   graphdo todo create --title "Book travel for conference" --body "Dates: March 15-17"
   ```

### Task management

```bash
graphdo todo list
graphdo todo complete --id "AQMkAD..."
graphdo todo create --title "Prepare slides for standup"
graphdo todo delete --id "AQMkAE..."
```

## Error Handling

| Symptom | Fix |
|---------|-----|
| Auth or token errors | Tell the user to run `graphdo login` |
| Config or list-not-found errors | Tell the user to run `graphdo config` |
| Missing `--subject` or `--body` | Provide both flags |
| Graph API error (HTTP status + code + message) | Read the error from stderr for details |

## Tips

- Always quote flag values: `--title "My task"`, not `--title My task`
- Use `--body -` with stdin for long content to avoid shell escaping problems
- Check readiness with `graphdo status` before running other commands
- Parse JSON from stdout only; ignore stderr (it's for humans)
