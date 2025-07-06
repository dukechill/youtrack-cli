# youtrack-cli

A command-line interface for interacting with YouTrack, optimized for terminal-based workflows and Neovim/LazyVim integration.

---

## 🚀 Installation

```bash
# Clone the repository
git clone https://github.com/dukechiu/youtrack-cli.git

# Navigate to the project directory
cd youtrack-cli

# Build the executable
go build -o youtrack-cli

# (Optional) Move to global path for system-wide access
sudo mv youtrack-cli /usr/local/bin/
chmod +x /usr/local/bin/youtrack-cli
```

---

## ⚙️ Configuration

Before using the tool, configure your YouTrack URL and API Token:

```bash
youtrack-cli configure
```

1. **Enter your YouTrack URL**  
   - Use the base URL (e.g., `https://yourcompany.youtrack.cloud` or `https://youtrack.example.com`)  
   - Do **not** include `/api` or any path.

2. **Enter your API Token**  
   - Log in to YouTrack.  
   - Go to: `Profile > Account Security` (or `Hub > Authentication`)  
   - Click `New token`, name it (e.g., `youtrack-cli`), and select `YouTrack` scope.  
   - Copy the token (starts with `perm:`) and paste it in the prompt.

3. **Verify Configuration**  
   It will be saved to `~/.youtrack-cli.yaml`:

```yaml
youtrack_url: https://yourcompany.youtrack.cloud
api_token: perm:your-token
```

---

## 🧰 Usage

### List Issues

```bash
# List your assigned issues
youtrack-cli list

# Output in JSON (for Neovim integration)
youtrack-cli list --json
```

### Add Work Item

```bash
youtrack-cli add-work [issue-id] [minutes] [description]
```

Example:

```bash
youtrack-cli add-work DP-123 60 "Fixed a bug"
```

### Check Work

```bash
# Find issues assigned to you without work logged today
youtrack-cli check-work
```

---

## 🧠 Neovim/LazyVim Integration

1. Copy integration script:

```bash
cp neovim-integration.lua ~/.config/nvim/lua/plugins/
```

2. Sync LazyVim:

```vim
:Lazy sync
```

3. Use shortcuts:

- `<leader>yl`: List YouTrack issues
- `<leader>ya`: Add work item
- `:YouTrackList`: Display issues in Telescope (if configured)

4. **(Optional)** Enable reminders:  
   If using `nvim-notify`, it will show alerts when `youtrack-cli check-work` detects missing work logs.

---

## 🛠 Development

- Source: Edit `main.go`  
- Dependencies: Managed via `go.mod`  
- Build:

```bash
go build -o youtrack-cli
```

- Git setup (SSH):

```bash
git remote add origin git@github.com:dukechill/youtrack-cli.git
git branch -M main
git push -u origin main
```

---

## 🧪 Troubleshooting

### Configuration Issues

- Check `~/.youtrack-cli.yaml` contains the correct URL/token.
- Test API connectivity:

```bash
curl -H "Authorization: Bearer perm:your-token" https://yourcompany.youtrack.cloud/api/issues?query=assignee:me
```

### Neovim Integration Issues

- Confirm `telescope.nvim` and `nvim-notify` are installed:

```vim
:Lazy list
```

- Verify command paths in `neovim-integration.lua`

---

## 📬 Contact

For bugs or suggestions, open an issue:  
👉 [GitHub Repo](https://github.com/dukechill/youtrack-cli)
