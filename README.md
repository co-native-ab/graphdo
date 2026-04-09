# graphdo

## What is graphdo?

graphdo is a small, friendly command-line tool that lets you **send emails to yourself** and **manage your to-do list** — all from your terminal.

It's designed to work hand-in-hand with AI assistants like [Claude Code](https://docs.anthropic.com/en/docs/claude-code). Here's a common workflow:

1. You ask Claude to **summarize your day** — meetings, emails, action items (using the Claude.ai Microsoft 365 Connector).
2. Claude uses graphdo to **send you a summary email** so everything lands in your inbox.
3. Claude **creates follow-up tasks** in your Microsoft To Do list so nothing falls through the cracks.

Think of graphdo as a bridge between your AI assistant and your Microsoft 365 account. You stay organized without lifting a finger.

**Getting started takes about 2 minutes:** [install graphdo](#installation), then run three commands — `login`, `config`, and `mcp install` (or `skill install`). That's it. See [Getting Started](#getting-started) for the full walkthrough.

---

## Prerequisites

- A **Microsoft 365 account** (work, school, or personal — any will do)
- If you're using a **work or school** account, your IT administrator may need to approve graphdo for your organization (see [Organization Setup](#organization-setup) below)
- That's it! graphdo handles everything else.

---

## Installation

Choose your operating system below. Each section provides an **automatic install script** (recommended) and **manual steps** if you prefer.

👉 **All releases:** [github.com/co-native-ab/graphdo/releases/latest](https://github.com/co-native-ab/graphdo/releases/latest)

### Windows

**Automatic install (PowerShell):**

```powershell
$LatestTag = (Invoke-RestMethod -Uri "https://api.github.com/repos/co-native-ab/graphdo/releases/latest").tag_name
$Version = $LatestTag -replace "^v", ""
$Arch = (Get-CimInstance Win32_ComputerSystem).SystemType
switch ($Arch) {
    "x64-based PC" { $Arch = "amd64" }
    "ARM64-based PC" { $Arch = "arm64" }
    default { throw "Unsupported architecture: $Arch" }
}
$TempDir = New-TemporaryFile | % { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
Invoke-WebRequest "https://github.com/co-native-ab/graphdo/releases/download/v${Version}/graphdo-windows-${Arch}.exe" -OutFile "${TempDir}\graphdo.exe"
Move-Item "${TempDir}\graphdo.exe" "${ENV:LOCALAPPDATA}\Microsoft\WindowsApps\"
```

Open a **new** terminal window and type `graphdo --help` to verify. If you see a list of commands, you're all set! 🎉

<details>
<summary>Manual steps</summary>

1. Download **`graphdo-windows-amd64.exe`** (or `graphdo-windows-arm64.exe` for ARM) from the [releases page](https://github.com/co-native-ab/graphdo/releases/latest).
2. Rename the file to **`graphdo.exe`**.
3. Move `graphdo.exe` into `%LOCALAPPDATA%\Microsoft\WindowsApps\` (this folder is already in your PATH).
4. Open a **new** terminal window and type:
   ```
   graphdo --help
   ```
   If you see a list of commands, you're all set! 🎉

</details>

### macOS

**Automatic install:**

```shell
LATEST_TAG=$(curl -s https://api.github.com/repos/co-native-ab/graphdo/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${LATEST_TAG#"v"}
ARCH=$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')
TEMP_DIR=$(mktemp -d)
curl -L "https://github.com/co-native-ab/graphdo/releases/download/v${VERSION}/graphdo-darwin-${ARCH}" -o "${TEMP_DIR}/graphdo"
chmod +x "${TEMP_DIR}/graphdo"
sudo mv "${TEMP_DIR}/graphdo" /usr/local/bin/graphdo
```

Type `graphdo --help` to verify. If you see a list of commands, you're all set! 🎉

> **If macOS blocks it** with a message like "cannot be opened because the developer cannot be verified":
> Open **System Settings** → **Privacy & Security**, scroll down and look for a message about graphdo being blocked, then click **"Allow Anyway"** and try again.

<details>
<summary>Manual steps</summary>

1. Download the right file for your Mac from the [releases page](https://github.com/co-native-ab/graphdo/releases/latest):
   - **Intel Mac** → `graphdo-darwin-amd64`
   - **Apple Silicon (M1, M2, M3, M4)** → `graphdo-darwin-arm64`

   Not sure which you have? Click the Apple menu → "About This Mac". If it says "Apple M1" (or M2, M3, M4), you have Apple Silicon. Otherwise, you have Intel.

2. Open **Terminal** (search for it in Spotlight with Cmd+Space).
3. Go to your Downloads folder:
   ```
   cd ~/Downloads
   ```
4. Rename the file to `graphdo` (replace the filename below with the one you downloaded):
   ```
   mv graphdo-darwin-arm64 graphdo
   ```
5. Make it runnable (this gives your computer permission to execute it):
   ```
   chmod +x graphdo
   ```
6. Move it to a system folder so you can use it from anywhere:
   ```
   sudo mv graphdo /usr/local/bin/
   ```
   You'll be asked for your password — that's normal.
7. **If macOS blocks it** with a message like "cannot be opened because the developer cannot be verified":
   - Open **System Settings** → **Privacy & Security**.
   - Scroll down and look for a message about graphdo being blocked.
   - Click **"Allow Anyway"**.
   - Try the next step again.
8. Verify it works:
   ```
   graphdo --help
   ```
   If you see a list of commands, you're all set! 🎉

</details>

### Linux

**Automatic install:**

```shell
LATEST_TAG=$(curl -s https://api.github.com/repos/co-native-ab/graphdo/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${LATEST_TAG#"v"}
ARCH=$(uname -m | sed -e 's/x86_64/amd64/' -e 's/armv[0-9]*/&/' -e 's/aarch64/arm64/')
TEMP_DIR=$(mktemp -d)
curl -L "https://github.com/co-native-ab/graphdo/releases/download/v${VERSION}/graphdo-linux-${ARCH}" -o "${TEMP_DIR}/graphdo"
chmod +x "${TEMP_DIR}/graphdo"
sudo mv "${TEMP_DIR}/graphdo" /usr/local/bin/graphdo
```

Type `graphdo --help` to verify. If you see a list of commands, you're all set! 🎉

<details>
<summary>Manual steps</summary>

1. Download the right file for your system from the [releases page](https://github.com/co-native-ab/graphdo/releases/latest):
   - **Most desktops/laptops** → `graphdo-linux-amd64`
   - **ARM devices (Raspberry Pi, etc.)** → `graphdo-linux-arm64`
2. Open a terminal and go to your Downloads folder:
   ```
   cd ~/Downloads
   ```
3. Make it runnable:
   ```
   chmod +x graphdo-linux-amd64
   ```
4. Move it to a system folder so you can use it from anywhere:
   ```
   sudo mv graphdo-linux-amd64 /usr/local/bin/graphdo
   ```
5. Verify it works:
   ```
   graphdo --help
   ```
   If you see a list of commands, you're all set! 🎉

</details>

---

## Getting Started

Once graphdo is installed, there are just **three steps** to get everything working. Each step only needs to be done once.

### Step 1: Sign in to your Microsoft account

```
graphdo login
```

Your web browser will open automatically to Microsoft's sign-in page. Sign in with your Microsoft account and approve the permissions. Once you see the success page in your browser, you can close it.

You'll see a "✓ Logged in successfully" message in your terminal.

> **Trouble opening the browser?** If graphdo can't open a browser (for example, on a remote server), use device code flow instead:
> ```
> graphdo login --device-code
> ```
> This will show a link and a code — open the link on any device, enter the code, and sign in there.

### Step 2: Pick your todo list

```
graphdo config
```

This will show you all your Microsoft To Do lists and let you pick which one graphdo should use. Just follow the prompts.

### Step 3: Connect graphdo to your AI assistant

There are two ways to connect graphdo to your AI assistant. Choose the one that fits your setup:

#### Option A: MCP server

If your AI tool supports the [Model Context Protocol](https://modelcontextprotocol.io) (Claude Code, Claude Desktop, VS Code, GitHub Copilot CLI), install the MCP server:

```
graphdo mcp install
```

This is an interactive command — it will ask you which tool to configure. The MCP server lets your AI assistant call graphdo's tools directly, which is faster and more reliable than running CLI commands.

For non-interactive use:

```bash
graphdo mcp install --target claude-code     # Claude Code
graphdo mcp install --target claude-desktop  # Claude Desktop
graphdo mcp install --target vscode          # VS Code (current workspace)
graphdo mcp install --target copilot         # GitHub Copilot CLI
```

#### Option B: Skill file

If your AI tool doesn't support MCP, or you prefer the agent to use CLI commands directly, install the skill file instead. This teaches the assistant how to use graphdo's CLI commands:

```
graphdo skill install
```

This will ask you which AI assistant you use and where to install the skill file. If you're using **Claude Code**, choose **"Claude Code — user profile"**. This installs the skill for all your projects automatically.

> **Not sure what to pick?** Use the **MCP server** (Option A) if your AI tool runs in an isolated environment without shell access (e.g., Claude Cowork). Use the **skill file** (Option B) if your agent has direct shell access and you want it to use CLI commands. Both work well — choose what fits your setup.

### ✅ You're all set!

Run this to confirm everything is working:

```
graphdo status
```

If you see `"ready": true`, you're good to go! 🎉

---

## Using graphdo with Claude Code

This is the most common way to use graphdo — and the reason it was built.

If you have the **Claude.ai Microsoft 365 Connector** active (the one that can read your emails and calendar), you can ask Claude to summarize your day and send the summary to your inbox. Here's how:

1. Make sure you've completed the three setup steps above (login, config, skill install).
2. Open Claude Code in your terminal.
3. Ask Claude something like:

> "Summarize the important emails and meetings I had in the last 24 hours. Then send the summary to me using graphdo, and create todo items for any follow-up actions."

That's it. Claude will:
- Use the Microsoft 365 Connector to read your recent emails and calendar events.
- Write a summary of what happened.
- Use `graphdo mail send` to email the summary to you.
- Use `graphdo todo create` to add follow-up tasks to your todo list.

You wake up the next morning with a clean summary in your inbox and actionable tasks in Microsoft To Do. ☕

### More things you can ask Claude

- *"Check my todos and mark anything related to the Q3 report as completed."*
- *"Create a task to review the budget proposal by Friday."*
- *"Send me an email with a list of all the meetings I have tomorrow."*

As long as the skill is installed and graphdo is set up, Claude knows how to use it.

---

## Organization Setup

> **Personal Microsoft accounts** (like @outlook.com or @hotmail.com) can skip this section entirely. This is only relevant for work or school accounts managed by an organization.

### For regular users

When you first run `graphdo login`, Microsoft may tell you that you need admin approval. This means your IT administrator needs to grant graphdo permission to access your email and tasks on behalf of your organization.

**What to tell your IT admin:**

> "I'd like to use a tool called graphdo that helps me send emails to myself and manage my todo list from the command line. It needs admin consent for these permissions: User.Read, Mail.Send, Tasks.ReadWrite, and offline_access. The application ID is `b073490b-a1a2-4bb8-9d83-00bb5c15fcfd` and it's published by Co-native AB."

### For IT administrators

graphdo is a multi-tenant application published by Co-native AB. To grant consent for your organization:

1. Go to the [Azure Portal](https://portal.azure.com) → **Microsoft Entra ID** → **Enterprise applications**.
2. Click **New application** → **All applications** → search for the application ID: `b073490b-a1a2-4bb8-9d83-00bb5c15fcfd`.
3. If the app doesn't appear, users can trigger the consent flow by running `graphdo login` — this will create a service principal in your tenant.
4. Go to **Permissions** and click **Grant admin consent for [your organization]**.
5. Review and approve the following delegated permissions:

   | Permission | Type | Description |
   |-----------|------|-------------|
   | `User.Read` | Delegated | Read the signed-in user's basic profile |
   | `Mail.Send` | Delegated | Send mail as the signed-in user |
   | `Tasks.ReadWrite` | Delegated | Read and write the signed-in user's tasks |
   | `offline_access` | Delegated | Maintain access to data you have given it access to (enables refresh tokens) |

6. Once consent is granted, all users in your organization can use `graphdo login` without further approval.

**Security notes:**
- graphdo can **only send emails to the signed-in user themselves** — it cannot send to other recipients.
- graphdo only accesses the user's **own tasks** in Microsoft To Do.
- The source code is open at [github.com/co-native-ab/graphdo](https://github.com/co-native-ab/graphdo).

---

## Using Your Own App Registration

If your organization prefers to use its own Azure AD application instead of the shared one, you can create your own app registration and configure graphdo to use it.

### Why use your own app?

- Your organization's security policy doesn't allow third-party multi-tenant apps
- You want to control the app registration lifecycle
- You want to restrict which users can use graphdo

### Quick setup

1. **Create an app registration** — follow the step-by-step guide in [docs/azure-app-registration.md](docs/azure-app-registration.md)
2. **Tell graphdo to use it:**
   ```
   graphdo login --client-id "your-app-client-id"
   ```
   The client ID is saved to your configuration file, so you only need to pass it once. All future commands will use it automatically.

3. **Or set it as an environment variable:**
   ```
   export GRAPHDO_CLIENT_ID="your-app-client-id"
   ```

### App registration requirements

If you're creating your own app registration, make sure it has:

| Setting | Value |
|---------|-------|
| **Supported account types** | "Accounts in any organizational directory and personal Microsoft accounts" (multi-tenant + personal) — or "Accounts in this organizational directory only" (single-tenant) if you only need it for your org |
| **Platform** | Mobile and desktop applications |
| **Redirect URI** | `http://localhost` |
| **Allow public client flows** | Yes |
| **API permissions (delegated)** | `User.Read`, `Mail.Send`, `Tasks.ReadWrite`, `offline_access` |

---

## Commands

### `graphdo login`

Sign in to your Microsoft account. You only need to do this once (or again if your session expires).

```
graphdo login
```

---

### `graphdo logout`

Clear your cached credentials. Use this if you want to sign in with a different account or if you're having authentication issues.

```
graphdo logout
```

---

### `graphdo status`

Check that everything is set up correctly. Returns a JSON summary showing whether you're logged in, your user, your configured todo list, and whether graphdo is ready to use.

```
graphdo status
```

Example output:

```json
{
  "ready": true,
  "logged_in": true,
  "user": "you@example.com",
  "todo_list": "My Tasks",
  "todo_count": 5
}
```

If something isn't configured, `ready` will be `false` and an `error` field will explain what's missing.

---

### `graphdo config`

Pick which todo list graphdo should use. This is interactive — it will show your lists and let you choose.

```
graphdo config
```

### `graphdo config show`

See your current settings (which todo list is selected, etc.).

```
graphdo config show
```

---

### `graphdo mail send`

Send an email to yourself.

| Flag | Description | Required? |
|------|-------------|-----------|
| `--subject "..."` | The email subject line | Yes |
| `--body "..."` | The email body. Use `-` to read from piped input. | Yes |
| `--html` | Send the body as HTML instead of plain text | No |

**Examples:**

Send a simple email:
```
graphdo mail send --subject "Meeting Notes" --body "Remember to follow up with team"
```

Send a longer email by piping text in (the `-` tells graphdo to read the body from the pipe):
```
echo "Long email body here" | graphdo mail send --subject "Notes" --body -
```

Send an HTML email:
```
graphdo mail send --subject "Formatted Report" --body "<h1>Daily Report</h1><p>Everything looks good.</p>" --html
```

---

### `graphdo todo list`

Show your todo items from the selected list. The output is in JSON format, which makes it easy for AI assistants to read and work with.

```
graphdo todo list
```

For long lists, you can paginate:

| Flag | Description | Default |
|------|-------------|---------|
| `--top N` | How many items to show | 20 |
| `--skip N` | How many items to skip (for the next page) | 0 |

**Example — show 10 items at a time:**
```
graphdo todo list --top 10
graphdo todo list --top 10 --skip 10
```

---

### `graphdo todo show`

Show the details of a single task.

| Flag | Description | Required? |
|------|-------------|-----------|
| `--id "..."` | The ID of the task (from `todo list`) | Yes |

```
graphdo todo show --id "AAMkAD..."
```

---

### `graphdo todo create`

Create a new task in your selected todo list.

| Flag | Description | Required? |
|------|-------------|-----------|
| `--title "..."` | The task title | Yes |
| `--body "..."` | A description or notes for the task | No |

**Examples:**

Create a simple task:
```
graphdo todo create --title "Buy groceries"
```

Create a task with a description:
```
graphdo todo create --title "Prepare presentation" --body "Include Q3 metrics and team updates"
```

---

### `graphdo todo update`

Update a task's title or description (or both).

| Flag | Description | Required? |
|------|-------------|-----------|
| `--id "..."` | The ID of the task to update (from `todo list`) | Yes |
| `--title "..."` | New title for the task | No (but at least one of title/body required) |
| `--body "..."` | New description for the task | No (but at least one of title/body required) |

**Examples:**

Change a task's title:
```
graphdo todo update --id "AAMkAD..." --title "Buy organic groceries"
```

Add notes to a task:
```
graphdo todo update --id "AAMkAD..." --body "Remember to check the sale section"
```

---

### `graphdo todo complete`

Mark a task as done.

| Flag | Description | Required? |
|------|-------------|-----------|
| `--id "..."` | The ID of the task to complete (from `todo list`) | Yes |

```
graphdo todo complete --id "AAMkAD..."
```

---

### `graphdo todo delete`

Remove a task entirely.

| Flag | Description | Required? |
|------|-------------|-----------|
| `--id "..."` | The ID of the task to delete (from `todo list`) | Yes |

```
graphdo todo delete --id "AAMkAD..."
```

---

### `graphdo skill install`

Install the graphdo agent skill file so AI agents (Claude Code, GitHub Copilot) know how to use graphdo. This is an **interactive** command — it will ask where to install the skill file.

```
graphdo skill install
```

You can also use flags for non-interactive installation:

| Flag | Description |
|------|-------------|
| `--agent` | Agent type: `claude` or `copilot` |
| `--scope` | Where to install: `project` or `user` |
| `--output "..."` | Write the skill file to a specific file path |
| `--stdout` | Print the skill file content to the terminal |

**Installation targets:**

| Option | Path |
|--------|------|
| Current project | `.agents/skills/graphdo/SKILL.md` |
| Claude Code (user) | `~/.claude/skills/graphdo/SKILL.md` |
| GitHub Copilot (user) | `~/.copilot/skills/graphdo/SKILL.md` |

**Examples:**

```bash
# Interactive — choose where to install
graphdo skill install

# Install to the current project (works with both Claude and Copilot)
graphdo skill install --agent claude --scope project

# Install to your Claude Code user profile
graphdo skill install --agent claude --scope user

# Just print the skill file to the terminal
graphdo skill install --stdout
```

---

### `graphdo mcp run`

Start graphdo as a **stdio MCP server**, exposing all mail and todo operations as MCP tools. AI agents that support the [Model Context Protocol](https://modelcontextprotocol.io) can call graphdo's tools directly without using the CLI.

```
graphdo mcp run
```

The server reads from stdin and writes to stdout (stdio transport). Logging goes to stderr and does not interfere with the protocol.

**Exposed MCP tools:**

| Tool | Description |
|------|-------------|
| `mail_send` | Send an email to yourself |
| `todo_list` | List todos from your configured list |
| `todo_show` | Get a single todo by ID |
| `todo_create` | Create a new todo |
| `todo_update` | Update a todo's title and/or body |
| `todo_complete` | Mark a todo as completed |
| `todo_delete` | Delete a todo |

> **Note:** graphdo must be [configured](#step-2-pick-your-todo-list) before todo tools will work. Run `graphdo config` first.

---

### `graphdo mcp install`

Write the graphdo MCP server entry to an AI tool's config file so it can call `graphdo mcp run` automatically. This is an **interactive** command — it will ask which tool to configure.

```
graphdo mcp install
```

You can also use a flag for non-interactive installation:

| Flag | Description |
|------|-------------|
| `--target` | Target tool: `claude-code`, `claude-desktop`, `vscode`, or `copilot` |

**Installation targets:**

| Target | Config file |
|--------|-------------|
| `claude-code` | `~/.claude.json` |
| `claude-desktop` | macOS: `~/Library/Application Support/Claude/claude_desktop_config.json` · Linux: `~/.config/claude/claude_desktop_config.json` · Windows: `%APPDATA%\Claude\claude_desktop_config.json` |
| `vscode` | `.vscode/mcp.json` (workspace-relative) |
| `copilot` | `~/.copilot/mcp.json` |

The install is **additive** — existing server entries in the config file are preserved. Only the `graphdo` entry is added or updated.

**Examples:**

```bash
# Interactive — choose which tool to configure
graphdo mcp install

# Install for Claude Code
graphdo mcp install --target claude-code

# Install for VS Code (in the current workspace)
graphdo mcp install --target vscode
```

---

## Global Flags

These flags work with any command.

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--graph-url` | `GRAPHDO_GRAPH_URL` | Use a different Microsoft Graph API endpoint (for sovereign cloud environments like GCC or China). |
| `--device-code` | — | Use device code flow for login instead of opening a browser. Useful for remote servers or headless environments. |
| `--config-dir` | `GRAPHDO_CONFIG_DIR` | Use a custom configuration directory instead of the OS default. |
| `--access-token` | `GRAPHDO_ACCESS_TOKEN` | Provide a pre-obtained access token directly, skipping the login flow entirely. |
| `--debug` | — | Enable verbose debug logging (written to stderr). |

The `--client-id` flag is available on the `login` command only. See [Using Your Own App Registration](#using-your-own-app-registration) for details.

---

## For AI Assistants

graphdo provides two ways for AI assistants to interact with it:

1. **MCP server** — Run `graphdo mcp install` to register graphdo as an MCP server. Your AI tool calls graphdo's tools directly via the [Model Context Protocol](https://modelcontextprotocol.io), which avoids shell escaping issues. Required for isolated environments without shell access (e.g., Claude Cowork). See [Step 3](#step-3-connect-graphdo-to-your-ai-assistant) in Getting Started.

2. **Skill file** — Run `graphdo skill install` to install a skill file that teaches the AI assistant how to use graphdo's CLI commands. Works great when the agent has direct shell access (e.g., running locally with Claude Code).

For the raw skill file, see [SKILL.md](SKILL.md).

---

## Troubleshooting

**"I get a permission error"**
Make sure you've signed in first by running `graphdo login`.

**"The login code expired"**
No worries — just run `graphdo login` again. The code is valid for about 15 minutes, so try to complete the sign-in promptly.

**"No todo lists found"**
You'll need to create a list in Microsoft To Do first. Open [to-do.office.com](https://to-do.office.com), create a list, then come back and run `graphdo config`.

**"You need admin approval"**
Your organization's IT administrator needs to approve graphdo. See [Organization Setup](#organization-setup) for what to tell them.

**"The browser didn't open"**
Try running `graphdo login --device-code` instead. This gives you a link and code you can use on any device.

**"I want to use a different Azure AD app"**
Run `graphdo login --client-id "your-client-id"` with your own app's client ID. See [Using Your Own App Registration](#using-your-own-app-registration) for details.

**"Command not found"**
This means your computer can't find graphdo. Make sure you followed the Installation steps above — specifically the part about adding graphdo to your PATH.

**"Cannot be opened because the developer cannot be verified" (macOS)**
This is a macOS security feature. Go to **System Settings → Privacy & Security**, find the message about graphdo, and click **"Allow Anyway"**.

---

## Privacy & Security

Your privacy matters. Here's what you should know:

- 🔒 graphdo only accesses **your own** email and tasks. It cannot access anyone else's.
- 📧 It can **only send emails to yourself** — it is not possible to send emails to other people.
- 💻 Your login credentials are cached **locally on your computer** and nowhere else.
- 🌐 No data is sent anywhere except to **Microsoft's official servers** (the same ones Outlook and To Do use).
- 📖 The source code is **fully open** at [github.com/co-native-ab/graphdo](https://github.com/co-native-ab/graphdo) — anyone can review exactly what it does.
