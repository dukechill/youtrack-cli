# youtrack-cli

A command-line interface for interacting with YouTrack.

## Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/dukechiu/youtrack-cli.git
    ```
2.  Navigate to the project directory:
    ```bash
    cd youtrack-cli
    ```
3.  Build the executable:
    ```bash
    go build
    ```

## Configuration

Before using the tool, you need to configure your YouTrack URL and API token:

```bash
./youtrack-cli configure
```

This will prompt you to enter your YouTrack URL and a permanent token. The configuration will be saved to `~/.youtrack-cli.yaml`.

## Usage

### List Issues

To list your assigned issues:

```bash
./youtrack-cli list
```

To get the output in JSON format:

```bash
./youtrack-cli list --json
```

### Add Work Item

To add a work item to an issue:

```bash
./youtrack-cli add-work [issue-id] [minutes] [description]
```

Example:

```bash
./youtrack-cli add-work DP-123 60 "Fixed a bug"
```

### Check Work

To check for issues assigned to you that you haven't logged work for today:

```bash
./youtrack-cli check-work
```
