# graphdo

## What is graphdo?

graphdo is a small, friendly command-line tool that lets you **send emails to yourself** and **manage your to-do list** — all from your terminal.

It's designed to work hand-in-hand with AI assistants like Claude. Here's a common workflow:

1. Your AI assistant summarizes your day — meetings, action items, things to follow up on.
2. It uses graphdo to **send you a summary email** so everything lands in your inbox.
3. It **creates follow-up tasks** in your Microsoft To Do list so nothing falls through the cracks.

Think of graphdo as a bridge between your AI assistant and your Microsoft 365 account. You stay organized without lifting a finger.

---

## Prerequisites

- A **Microsoft 365 account** (work, school, or personal — any will do)
- If you're using a **work or school** account, your IT administrator may need to approve graphdo for your organization (see [Organization Setup](#organization-setup) below)
- That's it! graphdo handles everything else.

---

## Installation

Download graphdo for your operating system, then follow the steps below.

👉 **Download the latest version here:** [github.com/co-native-ab/graphdo/releases/latest](https://github.com/co-native-ab/graphdo/releases/latest)

### Windows

1. Download **`graphdo-windows-amd64.exe`** from the link above.
2. Rename the file to **`graphdo.exe`**.
3. Create a folder called `C:\Tools\` (if it doesn't already exist) and move `graphdo.exe` into it.
4. Add `C:\Tools\` to your system PATH so you can run graphdo from anywhere:
   - Click the **Start menu** and search for **"Environment Variables"**.
   - Click **"Edit the system environment variables"**.
   - Click the **"Environment Variables…"** button.
   - Under **"User variables"**, find **Path** and click **Edit**.
   - Click **New** and type `C:\Tools\`.
   - Click **OK** on all the windows to save.
5. Open a **new** terminal window (the old one won't see the change) and type:
   ```
   graphdo --help
   ```
   If you see a list of commands, you're all set! 🎉

### macOS

1. Download the right file for your Mac:
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

### Linux

1. Download the right file for your system:
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

---

## Getting Started

Once graphdo is installed, follow these steps to get up and running for the first time.

### Step 1: Sign in to your Microsoft account

```
graphdo login
```

Your web browser will open automatically to Microsoft's sign-in page. Here's what to do:

1. Sign in with your Microsoft account.
2. Approve the permissions — graphdo needs access to send emails and manage your tasks.
3. Once you see the success page in your browser, you can close it.

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

### Step 3: Send your first email

```
graphdo mail send --subject "Hello from graphdo" --body "It works!"
```

Check your inbox — you should see an email from yourself. ✉️

### Step 4: Check your todos

```
graphdo todo list
```

This shows your current tasks from the list you picked in Step 2.

---

## Organization Setup

> **Personal Microsoft accounts** (like @outlook.com or @hotmail.com) can skip this section entirely. This is only relevant for work or school accounts managed by an organization.

### For regular users

When you first run `graphdo login`, Microsoft may tell you that you need admin approval. This means your IT administrator needs to grant graphdo permission to access your email and tasks on behalf of your organization.

**What to tell your IT admin:**

> "I'd like to use a tool called graphdo that helps me send emails to myself and manage my todo list from the command line. It needs admin consent for these permissions: User.Read, Mail.Send, and Tasks.ReadWrite. The application ID is `b073490b-a1a2-4bb8-9d83-00bb5c15fcfd` and it's published by Co-native AB."

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
   graphdo config --client-id "your-app-client-id"
   ```
   The client ID is saved to your configuration file, so you only need to pass it once.

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
| **API permissions (delegated)** | `User.Read`, `Mail.Send`, `Tasks.ReadWrite` |

---

## Commands

### `graphdo login`

Sign in to your Microsoft account. You only need to do this once (or again if your session expires).

```
graphdo login
```

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

## Global Flags

These flags work with any command.

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--client-id` | `GRAPHDO_CLIENT_ID` | Use a different Azure AD application. The client ID is saved to your config file so you only need to pass it once. Most users don't need this. |
| `--graph-url` | `GRAPHDO_GRAPH_URL` | Use a different Microsoft Graph API endpoint (for sovereign cloud environments like GCC or China). |
| `--device-code` | — | Use device code flow for login instead of opening a browser. Useful for remote servers or headless environments. |
| `--config-dir` | `GRAPHDO_CONFIG_DIR` | Use a custom configuration directory instead of the OS default. |
| `--access-token` | `GRAPHDO_ACCESS_TOKEN` | Provide a pre-obtained access token directly, skipping the login flow entirely. |
| `--debug` | — | Enable verbose debug logging (written to stderr). |

---

## For AI Assistants

If you're using graphdo with an AI assistant like Claude, see [SKILL.md](SKILL.md) for integration guidance.

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
