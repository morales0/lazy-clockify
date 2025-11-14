# Lazy Clockify

A command-line utility to quickly create time entries in Clockify from your terminal.

## Features

- Quick time entry creation from the command line
- Automatic ticket number detection from Git branch names
- Support for logging time to different dates
- Optional custom messages for time entries
- Configuration file support for default values

## Installation

```bash
go install github.com/morales0/lazy-clockify/lazyclock@latest
```

Or build from source:

```bash
git clone https://github.com/morales0/lazy-clockify.git
cd lazy-clockify
go build -o lazyclock
```

## Configuration

Create a `config.yaml` file in one of the following locations:
- Specifiy config path: `lazyclock --config ./config.yaml`
- Current directory: `./config.yaml`
- Home directory: `~/.lazy-clockify/config.yaml`

### Required Configuration

```yaml
api_key: "your-clockify-api-key"
start_time: "9:00"
end_time: "17:00"
ticket_prefix: "JIRA"
```

You can get your API key from your Clockify profile settings.

## Usage

### Create a Time Entry

Basic usage (uses today's date and configured times):

```bash
lazyclock new
```

The tool will:
1. Try to detect the ticket number from your current Git branch
2. If not found, use custom message or fail
3. Show a preview of the time entry
4. Ask for confirmation before submitting

### Custom Message

Add a custom message to your time entry (replaced ticket number):

```bash
lazyclock new -m "Implemented user authentication"
lazyclock new --message "Fixed bug in payment processing"
```

### Different Date

Log time for a different date:

```bash
lazyclock new -d 2025-01-15
lazyclock new --date 2025-01-15
```

## Git Branch Integration

The tool automatically extracts ticket numbers from your Git branch name based on the configured prefix.

For example, with `ticket_prefix: "EL"`:
- Branch: `feature/EL-1234-user-auth` → Ticket: `EL-1234`
- Branch: `bugfix/EL-5678-fix-payment` → Ticket: `EL-5678`

## Commands

### `new`

Create a new time entry:

```bash
lazyclock new [flags]
```

**Flags:**
- `-m, --message string` - Optional custom message to append to the time entry description
- `-d, --date string` - Date to log time entry (YYYY-MM-DD format, defaults to today)
- `--start_time string` - Override configured start time (default: from config)
- `--end_time string` - Override configured end time (default: from config)
- `--ticket_prefix string` - Override configured ticket prefix (default: from config)
