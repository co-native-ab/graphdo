---
name: graphdo
description: Send emails and manage Microsoft To Do tasks via the graphdo CLI. Use this skill whenever the user wants to send themselves an email summary, reminder, or report, create or manage tasks in Microsoft To Do, do daily planning or wrap-ups involving email and task management, or work with their Microsoft 365 productivity tools from the command line. Also use it when the user mentions graphdo by name, or asks about emailing themselves or managing their todo list programmatically.
---

# graphdo CLI

Send emails to yourself and manage Microsoft To Do tasks via Microsoft Graph.

All machine-readable output is JSON on stdout. Human messages go to stderr. Parse stdout only.

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
