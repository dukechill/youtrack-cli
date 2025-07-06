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

## 📂 Project Structure

This project follows a standard Go application layout, leveraging the Cobra framework for command-line interface management. The codebase is structured to separate CLI concerns from core business logic, enhancing maintainability and scalability.

```
youtrack-cli/
├─ cmd/                  # Cobra Commands
│  ├─ root.go            # Defines the root command and initializes all subcommands.
│  ├─ list.go            # Implements the 'youtrack-cli list' command for listing issues.
│  ├─ board.go           # Implements the 'youtrack-cli board' commands (e.g., 'list').
│  ├─ sprint.go          # Implements the 'youtrack-cli sprint' commands (e.g., 'list').
│  ├─ config/            # Commands for managing CLI configuration.
│  │  ├─ set.go          # Implements 'youtrack-cli config set'.
│  │  ├─ view.go         # Implements 'youtrack-cli config view' (raw config).
│  │  └─ show.go         # Implements 'youtrack-cli config show' (masked config).
│  ├─ work/              # Commands for managing work items.
│  │  ├─ add.go          # Implements 'youtrack-cli work add'.
│  │  └─ check.go        # Implements 'youtrack-cli work check'.
│  └─ helpers.go         # Shared flags or utility functions specific to Cobra commands.
├─ internal/             # Internal application logic (not exposed as a public API).
│  ├─ youtrack/          # Core logic for interacting with YouTrack API.
│  │  ├─ client.go       # Handles HTTP requests to YouTrack, including common GET/POST methods.
│  │  ├─ models.go       # Defines Go structs for YouTrack API data models (e.g., Issue, Sprint).
│  │  └─ sprint.go       # Contains algorithms for determining the current/latest sprint.
│  └─ config/            # Handles reading from and writing to the ~/.youtrack-cli.yaml configuration file.
│     └─ file.go         # Implements configuration loading, saving, and value setting.
├─ go.mod                # Go module definition and dependency management.
└─ main.go               # The application's entry point, simply calls cmd.Execute().
```

---

## ⚙️ Configuration

Before using the tool, configure your YouTrack URL and API Token:

### Initial Configuration

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

### View Configuration

Display your current YouTrack configuration. The API Token will be partially masked for security.

```bash
youtrack-cli config show
```

### Set Configuration Values

Set specific configuration values like your default board or sprint.

```bash
youtrack-cli config set [key] [value]
```

Examples:

```bash
youtrack-cli config set board "My Agile Board"
youtrack-cli config set sprint "Sprint 26"
```

### List Agile Boards

List all available Agile Boards in your YouTrack instance. This is useful for finding the exact board name to set as your default.

```bash
youtrack-cli config list-boards
```

### List Sprints for a Board

List all sprints for a specified board. If no board is specified, it uses the configured default board.

```bash
youtrack-cli sprint list --board "My Agile Board"
# Or, if a default board is configured:
youtrack-cli sprint list
```

---

## 🧰 Usage

### List Issues

```bash
# List your assigned issues
youtrack-cli list

# List issues for a specific sprint (wrap sprint name in quotes if it contains spaces)
youtrack-cli list -s "Sprint 26"

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
- `<leader>yc`: Show YouTrack configuration
- `<leader>yb`: List YouTrack boards
- `:TelescopeYoutrack`: Display issues in Telescope (if configured)

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

### API Connectivity and Query Issues

If `youtrack-cli` commands are not returning expected results, especially for `list` or `sprint list`, it might be due to incorrect configuration, API token issues, or incorrect board/sprint names.

1.  **Verify Configuration**: Use `youtrack-cli config show` to ensure your YouTrack URL, API Token, and configured board/sprint names are correct.

2.  **Test API Connectivity with `curl`**:  
    You can directly test the YouTrack API using `curl` to confirm connectivity and token validity. Replace `YOUR_YOUTRACK_URL` and `YOUR_API_TOKEN` with your actual values.

    *   **List your assigned issues:**
        ```bash
        curl -H "Authorization: Bearer YOUR_API_TOKEN" "YOUR_YOUTRACK_URL/api/issues?query=for:me&fields=idReadable,summary"
        ```

    *   **List issues in a specific board and sprint (e.g., "CRM促案管理" and "Sprint 26"):**
        ```bash
        curl -H "Authorization: Bearer YOUR_API_TOKEN" "YOUR_YOUTRACK_URL/api/issues?query=Board%20%22CRM%E4%BF%83%E6%A1%88%E7%AE%A1%E7%90%86%22%3A%20%7B%22Sprint%2026%22%7D&fields=idReadable,summary"
        ```
        *Note: The board and sprint names are URL-encoded. `CRM促案管理` becomes `%22CRM%E4%BF%83%E6%A1%88%E7%AE%A1%E7%90%86%22` and `Sprint 26` becomes `%22Sprint%2026%22`.*

    *   **List Agile Boards:**
        ```bash
        curl -H "Authorization: Bearer YOUR_API_TOKEN" "YOUR_YOUTRACK_URL/api/agiles?fields=id,name"
        ```

    *   **List Sprints for a specific Board (e.g., Board ID `121-114` for "CRM促案管理"):**
        ```bash
        curl -H "Authorization: Bearer YOUR_API_TOKEN" "YOUR_YOUTRACK_URL/api/agiles/121-114/sprints?fields=id,name"
        ```
        *You can find the Board ID using the `list-boards` command or the `curl` command above.*

3.  **Check Board and Sprint Names in YouTrack**:  
    Ensure the board and sprint names you are using in `youtrack-cli` commands exactly match those in your YouTrack instance.  
    *   **For Board Names**: Navigate to `Agile Boards` in YouTrack and verify the exact spelling and casing.  
    *   **For Sprint Names**: Go to your specific Agile Board, and check the names of the sprints. Pay close attention to spaces or special characters.

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

<xaiArtifact version_id="1.1" artifact_id="1f39190e-b939-458a-85d3-c6954b919bdf"/>
