# nvolt

<div align="center">

```
   ███╗   ██╗██╗   ██╗ ██████╗ ██╗  ████████╗
   ████╗  ██║██║   ██║██╔═══██╗██║  ╚══██╔══╝
██╔██╗ ██║██║   ██║██║   ██║██║     ██║
██║╚██╗██║╚██╗ ██╔╝██║   ██║██║     ██║
██║ ╚████║ ╚████╔╝ ╚██████╔╝███████╗██║
╚═╝  ╚═══╝  ╚═══╝   ╚═════╝ ╚══════╝╚═╝
```

**GitHub-native, Zero-Trust CLI for managing encrypted environment variables**

[![Go Version](https://img.shields.io/github/go-mod/go-version/iluxav/nvolt)](https://golang.org/doc/devel/release.html)
[![Release](https://img.shields.io/github/v/release/iluxav/nvolt)](https://github.com/iluxav/nvolt/releases)
[![License](https://img.shields.io/github/license/iluxav/nvolt)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/iluxav/nvolt)](https://goreportcard.com/report/github.com/iluxav/nvolt)

[Website](https://nvolt.io) • [Documentation](https://nvolt.io/docs.html) • [Quick Start](#quick-start)

</div>

**`nvolt`** is a cryptographically enforced secret manager built entirely around Git and local files. No server, no login, no organization model - just Git, encryption, and per-machine keypairs.

## Features

- **Zero-Trust Architecture**: All encryption/decryption happens locally
- **No Backend**: All data lives in Git repositories
- **No Authentication**: Uses Git for access control
- **Cryptographically Enforced**: Access control through wrapped keys
- **Git-Native**: `.nvolt/` directories act as encrypted, committed `.env` replacements
- **$0/month**: Free forever, no usage limits

## Why nvolt?

| Feature                | nvolt    | HashiCorp Vault     | Doppler | git-crypt   | SOPS        |
| ---------------------- | -------- | ------------------- | ------- | ----------- | ----------- |
| **Monthly Cost**       | **free** | $$$                 | $$      | free        | free        |
| **Zero-Knowledge**     | ✅       | ⚠️ Self-hosted only | ❌      | ✅          | ✅          |
| **No Backend**         | ✅       | ❌                  | ❌      | ✅          | ✅          |
| **No Login/Auth**      | ✅       | ❌                  | ❌      | ✅          | ✅          |
| **Per-Machine Access** | ✅       | ✅                  | ✅      | ⚠️ GPG only | ⚠️ GPG only |
| **Environment-Based**  | ✅       | ✅                  | ✅      | ❌          | ❌          |
| **Multi-Project**      | ✅       | ✅                  | ✅      | ⚠️ Limited  | ⚠️ Limited  |

## Installation

### Quick Install (Recommended)

```bash
# macOS and Linux (also works in Git Bash on Windows)
curl -sSL https://nvolt.io/install.sh | sh
```

### Using Go

```bash
go install github.com/iluxav/nvolt/cmd/nvolt@latest
```

### From Source

```bash
git clone https://github.com/iluxav/nvolt.git
cd nvolt
make build
```

## Quick Start

### Local Mode (Current Directory)

```bash
# Initialize vault in current directory
$ nvolt init
✓ Machine keypair generated
✓ Vault initialized at .nvolt/

# Push secrets from .env file
$ nvolt push -f .env
✓ Encrypted 12 secrets
✓ Secrets pushed to vault

# Pull and view secrets
$ nvolt pull
API_KEY=abc123
DB_PASSWORD=secret

# Run a command with secrets loaded
$ nvolt run npm start
✓ Loaded 12 secrets
🚀 Server running on port 3000
```

### Global Mode (Dedicated GitHub Repo)

```bash
# Initialize with a GitHub repository
nvolt init --repo org/secrets-repo

# Push secrets to production environment
nvolt push -f .env.production -e production

# Pull secrets from production
nvolt pull -e production
```

## Use Cases

- 🚀 **Startups & Solo Developers**: No monthly costs, enterprise-grade security without the enterprise price tag
- 👥 **Small Teams**: Securely share secrets across laptops and CI/CD using tools you already know
- 🔒 **Security-Conscious Organizations**: Zero-Trust architecture with no single point of failure
- 🤖 **CI/CD Pipelines**: Grant servers access to specific environments, secrets loaded at runtime

## Commands

### `nvolt init`

Initialize vault and generate machine keypair.

```bash
# Local mode (current directory)
nvolt init

# Global mode (dedicated GitHub repo)
nvolt init --repo org/secrets-repo
```

**Flags:**

- `--repo` - GitHub repository URL for global vault

---

### `nvolt push`

Encrypt and push secrets to the vault.

```bash
# From .env file
nvolt push -f .env.production -e production

# Set individual secrets with -k flag
nvolt push -k API_KEY=abc123 -k DB_PASSWORD=secret

# Multiple secrets with custom project name
nvolt push -k API_KEY=abc123 -k DB_SECRET=xyz789 -p my-backend -e staging
```

**Flags:**

- `-f, --file` - Path to .env file
- `-k, --key` - Key=value pairs (can be specified multiple times)
- `-e, --env` - Environment name (default: "default")
- `-p, --project` - Project name (auto-detected if not specified)

---

### `nvolt pull`

Decrypt and retrieve secrets from the vault.

```bash
# View secrets for default environment
nvolt pull

# View secrets for specific environment
nvolt pull -e production

# Write to .env file
nvolt pull -e production > .env.local
```

**Flags:**

- `-e, --env` - Environment name (default: "default")
- `-p, --project` - Project name (auto-detected if not specified)

---

### `nvolt run`

Run a command with decrypted secrets loaded as environment variables.

```bash
# Run development server
nvolt run npm start

# Run with specific environment
nvolt run -e production npm start

# Run arbitrary commands
nvolt run python app.py
```

**Flags:**

- `-e, --env` - Environment name (default: "default")
- `-c, --command` - Command to run

---

### `nvolt machine add`

Generate a new keypair for CI or another device.

```bash
nvolt machine add ci-server
nvolt machine add alice-laptop
```

---

### `nvolt machine rm`

Revoke machine access and re-wrap master keys.

```bash
nvolt machine rm old-laptop
```

---

### `nvolt vault show`

Display vault information and machine access.

```bash
nvolt vault show
```

---

### `nvolt vault verify`

Verify integrity of encrypted files and keys.

```bash
nvolt vault verify
```

---

### `nvolt sync`

Re-wrap or rotate master keys.

```bash
# Re-wrap keys for all machines
nvolt sync

# Rotate master key
nvolt sync --rotate
```

**Flags:**

- `--rotate` - Rotate the master encryption key

## Security

nvolt uses industry-standard cryptography to protect your secrets:

- **Encryption**: AES-256-GCM for secret encryption
- **Key Wrapping**: RSA-4096 for wrapping master keys
- **Local-Only**: All cryptographic operations happen on your machine
- **Audit Trail**: Every change is tracked in Git history
- **Zero-Knowledge**: nvolt never sees your plaintext secrets

### Reporting Vulnerabilities

If you discover a security vulnerability, please email [security@nvolt.io](mailto:security@nvolt.io). We take security seriously and will respond promptly.

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

- 📖 [Full Documentation](https://nvolt.io/docs.html) - Complete guide with examples
- 🏗️ [CLAUDE.md](CLAUDE.md) - Detailed architecture and implementation specifications
- 📋 [TASKS.md](TASKS.md) - Development progress tracking

## Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run `make check` to ensure tests pass
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

Please ensure your code follows the existing style and includes tests for new functionality.

## Links

- 🌐 [Website](https://nvolt.io)
- 📖 [Documentation](https://nvolt.io/docs.html)
- 💬 [Discussions](https://github.com/iluxav/nvolt/discussions)
- 🐛 [Issue Tracker](https://github.com/iluxav/nvolt/issues)
- 📦 [Releases](https://github.com/iluxav/nvolt/releases)

## Status

nvolt is in **active development**. Current stable version: [v1.0.21](https://github.com/iluxav/nvolt/releases)

## License

MIT - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built with ❤️ for developers who value security and simplicity**

Star ⭐ this repo if you find nvolt useful!

</div>
