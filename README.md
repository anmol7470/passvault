# PassVault

A secure, command-line password manager built with Go. Store, manage, and quickly access your passwords with strong encryption, all from your terminal.

## Features

- **Strong Encryption**: AES-256-GCM encryption with Argon2id key derivation
- **Password Strength Analysis**: Powered by zxcvbn (Dropbox's password strength estimator)
- **Secure Password Generator**: Generate cryptographically secure passwords
- **Quick Access Aliases**: Instantly copy passwords with custom aliases for stored passwords
- **Search and List**: Browse and search passwords with an intuitive interface
- **Password Auditing**: Analyze all stored passwords for security weaknesses
- **Export Capabilities**: Export passwords to JSON or CSV formats
- **Clipboard Integration**: One-command password copying
- **Local Storage**: All data stored locally in encrypted SQLite database

## Installation

### Prerequisites

[Go](https://go.dev/dl/) needs to be installed in your system

### Build from Source

```bash
git clone https://github.com/anmol7470/passvault.git
cd passvault
go build
```

Run the created binary to add a password:

```bash
./passvault add
```

### Installation

```bash
go install github.com/anmol7470/passvault@latest
```

Run the installed binary to add a password:

```bash
passvault add
```

## Commands

On first run, you'll be prompted to create a master password. This is used to encrypt all your stored passwords. Choose a strong, memorable password.

### `add`

Add a new password entry to the vault.

```bash
passvault add [flags]

Flags:
  -s, --service string    Service name
  -u, --username string   Username
  -p, --password string   Password
  -n, --notes string      Notes (optional)
  -a, --alias string      Alias for quick access (optional)
```

### `list`

Browse all passwords with an interactive interface.

```bash
passvault list
```

Features:

- Type to search in real-time
- Navigate with arrow keys or j/k
- Press Enter to view password details

### `get`

Retrieve a specific password.

```bash
passvault get [alias] [flags]

Flags:
  -q, --query string   Search query for service or username
```

If an exact alias match is found, the password is instantly copied to clipboard.

### `update`

Update an existing password entry.

```bash
passvault update [flags]

Flags:
  -q, --query string   Search query for service or username
```

### `delete`

Delete a password entry.

```bash
passvault delete [flags]

Flags:
  -q, --query string   Search query for service or username
```

Requires confirmation before deletion.

### `export`

Export passwords to JSON or CSV format.

```bash
passvault export [flags]

Flags:
  --json   Export as JSON
  --csv    Export as CSV
```

Exports are saved with timestamps and can be optionally opened after creation.

### `audit`

Analyze all stored passwords for security weaknesses.

```bash
passvault audit
```

Provides:

- Overall security statistics
- List of weak passwords requiring immediate action
- List of moderate passwords to consider strengthening
- Crack time estimates for each password
- Actionable recommendations

### `change-master-password`

Change your master password.

```bash
passvault change-master-password
```

All stored passwords will be re-encrypted with the new master password.

### `reset`

Completely reset the password vault.

```bash
passvault reset
```

Deletes all passwords and the master password. This action cannot be undone.
