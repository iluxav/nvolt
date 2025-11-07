# nvolt

**GitHub-native, Zero-Trust CLI for managing encrypted environment variables**

nvolt is a cryptographically enforced secret manager built entirely around Git and local files. No server, no login, no organization model - just Git, encryption, and per-machine keypairs.

## Features

- **Zero-Trust Architecture**: All encryption/decryption happens locally
- **No Backend**: All data lives in Git repositories
- **No Authentication**: Uses Git for access control
- **Cryptographically Enforced**: Access control through wrapped keys
- **Git-Native**: `.nvolt/` directories act as encrypted, committed `.env` replacements

## Installation

```bash
go install github.com/nvolt/nvolt/cmd/nvolt@latest
```

Or build from source:

```bash
make build
```

## Quick Start

### Local Mode (Current Directory)

```bash
# Initialize vault in current directory
nvolt init

# Push secrets from .env file
nvolt push -f .env

# Pull and view secrets
nvolt pull

# Run a command with secrets loaded
nvolt run npm start
```

### Global Mode (Dedicated GitHub Repo)

```bash
# Initialize with a GitHub repository
nvolt init --repo org/secrets-repo

# Push secrets
nvolt push -f .env -e production

# Pull secrets
nvolt pull -e production
```

## Commands

- `nvolt init [--repo <url>]` - Initialize vault and generate machine keypair
- `nvolt push` - Encrypt and push secrets to vault
- `nvolt pull` - Decrypt and pull secrets from vault
- `nvolt run` - Run a command with decrypted secrets
- `nvolt machine add/rm` - Manage machine access
- `nvolt vault show/verify` - Display vault info and verify integrity
- `nvolt sync [--rotate]` - Re-wrap or rotate master keys

## Development

```bash
# Install dependencies
make deps

# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Build binary
make build

# Run all checks
make check
```

## Project Structure

```
nvolt/
├── cmd/nvolt/          # Main entry point
├── internal/
│   ├── cli/            # CLI commands
│   ├── crypto/         # Cryptographic operations
│   ├── vault/          # Vault management
│   ├── git/            # Git operations
│   └── config/         # Configuration management
└── pkg/
    └── types/          # Shared types
```

## Documentation

See [CLAUDE.md](CLAUDE.md) for detailed architecture and implementation specifications.

See [TASKS.md](TASKS.md) for development progress tracking.

## License

MIT
