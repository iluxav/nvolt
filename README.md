# nvolt CLI

<div align="center">

**Ultra-Secured, Zero-Trust Environment Variable Manager**

[![License](https://img.shields.io/badge/License-MIT%20with%20Commons%20Clause-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)](https://golang.org/)

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Security](#security)

</div>

---

## Overview

**nvolt** is a CLI-first environment variable manager designed for developers who prioritize security without compromising workflow efficiency. Built on Zero-Trust principles, nvolt ensures that your secrets are encrypted locally before they ever leave your machine—the server only stores encrypted data and never has access to your private keys.

### Why nvolt?

- **🔐 Zero-Trust Security**: All encryption/decryption happens locally using asymmetric cryptography with wrapped keys
- **👥 Team Collaboration**: Share encrypted secrets across teams with granular role-based access control
- **🚀 Developer Experience**: Simple CLI commands that integrate seamlessly into your workflow
- **🔄 Multi-Environment Support**: Manage secrets across development, staging, production, and custom environments
- **🤖 CI/CD Ready**: Silent authentication for automated pipelines using challenge-response protocol
- **📦 No Lock-In**: Works with any server implementation—self-host or use the cloud version

---

## Features

### Core Capabilities

- **Wrapped Key Encryption**: Hybrid cryptography using RSA key pairs with AES-encrypted secrets
- **Per-Machine Authentication**: Each machine has its own cryptographic identity
- **Organization Management**: Multi-tenant support with admin and developer roles
- **Project & Environment Scoping**: Organize secrets by project and environment
- **Interactive Permissions**: Fine-grained access control (read/write/delete) at project and environment levels
- **Seamless Secret Injection**: Run commands with secrets loaded directly into memory—no `.env` files needed

### Security Model

nvolt implements a **wrapped key approach**:

1. CLI generates a unique symmetric key for each secret
2. The symmetric key is encrypted (wrapped) with each authorized machine's public key
3. Secrets are encrypted with the symmetric key
4. Server stores encrypted secrets and wrapped keys—never plaintext or private keys
5. Only machines with the corresponding private key can unwrap and decrypt secrets

This ensures **end-to-end encryption** where the server acts purely as encrypted storage.

---

## Installation

```bash
curl -fsSL https://install.nvolt.io/install.sh | bash
```

### Verify Installation

```bash
nvolt --version
```

### Build from Source

For developers who want to build from source:

```bash
git clone https://github.com/yourusername/nvolt-cli.git
cd nvolt-cli
go build -o nvolt cli/main.go
```

---

## Quick Start

### 1. Login

Authenticate using OAuth (opens browser):

```bash
nvolt login
```

This will:
- Generate RSA key pair (2048-bit)
- Store private key in `~/.nvolt/config.json`
- Send public key to server
- Create your default organization

### 2. Push Secrets

Push from a file:

```bash
nvolt push -f .env.production -p my-app -e production
```

Or push individual keys:

```bash
nvolt push -k DATABASE_URL=postgres://... -k API_KEY=secret123 -p my-app -e production
```

### 3. Pull Secrets

Retrieve and decrypt secrets:

```bash
# Print to console
nvolt pull -p my-app -e production

# Save to file
nvolt pull -f .env.production -p my-app -e production

# Pull specific key
nvolt pull -k DATABASE_URL -p my-app -e production
```

### 4. Run Commands with Secrets

Inject secrets directly into command execution (no files created):

```bash
nvolt run -p my-app -e production -c "npm start"
```

This is the recommended workflow—secrets are loaded into memory only and never touch disk.

---

## Documentation

### Commands Reference

#### Authentication

**`nvolt login`**
Interactive OAuth login via browser.

**`nvolt login --silent --machine ci-runner`**
Silent login for CI/CD using pre-provisioned private key. Requires `~/.nvolt/private_key.pem` to exist.

#### Machine Management

**`nvolt machine add <machine-name>`**
Generate key pair for a new machine (e.g., CI/CD runner). Displays private key **once**—securely transfer it to the target machine as `~/.nvolt/private_key.pem`.

#### Secret Management

**`nvolt push`**
```bash
# From file
nvolt push -f .env -p my-app -e staging

# Individual keys
nvolt push -k FOO=bar -k BAZ=qux -p my-app -e staging
```

**`nvolt pull`**
```bash
# All secrets to console
nvolt pull -p my-app -e staging

# All secrets to file
nvolt pull -f .env -p my-app -e staging

# Specific key
nvolt pull -k FOO -p my-app -e staging
```

**`nvolt run`**
```bash
nvolt run -p my-app -e staging -c "python manage.py runserver"
```

#### Synchronization

**`nvolt sync -p my-app -e production`**
Re-wrap keys for a specific project/environment after adding machines.

**`nvolt sync --all`**
Re-wrap all secrets across all projects and environments (recommended after adding machines).

#### Organization Management

**`nvolt org set -o <org-id>`**
Switch active organization. Without `-o`, shows interactive selector.

#### User Management (Admin Only)

**`nvolt user add <email>`**
Invite user to organization with interactive permission setup.

```bash
# With flags
nvolt user add user@example.com \
  -pp read=true,write=true,delete=false \
  -pe read=true,write=false,delete=false
```

**`nvolt user mod <email>`**
Modify user permissions (interactive with pre-selected current state).

**`nvolt user rm <email>`**
Remove user from organization.

### Flags

Global flags available for most commands:

- `-p, --project`: Project name (defaults to git repo name)
- `-e, --environment`: Environment name (defaults to `default`)
- `-o, --org`: Organization ID (defaults to active org in config)

---

## Architecture

### Scoping

All secrets are uniquely scoped by:

- **Organization ID**: Your team/company
- **User ID**: The authenticated user
- **Machine ID**: Unique machine identifier
- **Project Name**: Application/repository name
- **Environment**: deployment stage (e.g., `production`, `staging`)

### Encryption Flow Example

```
Machine A wants to share DB_PASSWORD with Machine B:

1. Machine A generates symmetric key: secret-key-123
2. Fetches public keys for Machine A and Machine B from server
3. Encrypts secret-key-123 with each public key:
   - wrapped-key-for-machine-A
   - wrapped-key-for-machine-B
4. Encrypts DB_PASSWORD value with secret-key-123
5. Sends encrypted value + wrapped keys to server

Machine B retrieves secret:

1. Downloads encrypted value + wrapped-key-for-machine-B
2. Decrypts wrapped-key-for-machine-B with private key → gets secret-key-123
3. Decrypts DB_PASSWORD value with secret-key-123
```

### Permissions Model

Three-tier RBAC system:

1. **Organization Level**: `admin` or `dev` role
2. **Project Level**: `{read, write, delete}` permissions
3. **Environment Level**: `{read, write, delete}` permissions

Auto-provisioning: First user to push to a new project/environment gets full access.

---

## Security

### Zero-Trust Principles

- Server never receives private keys or plaintext secrets
- All cryptographic operations performed client-side
- Challenge-response authentication for CI/CD (server never sees private key)
- Private keys stored locally at `~/.nvolt/config.json` and `~/.nvolt/private_key.pem`

### Best Practices

- Never commit `~/.nvolt/` to version control
- Transfer private keys for CI/CD machines using encrypted secret managers
- Rotate machine keys periodically using `nvolt machine add` + `nvolt sync --all`
- Use environment-specific permissions (restrict production write access)
- Audit user access regularly with `nvolt user mod`

### Threat Model

nvolt protects against:

- ✅ Server compromise (encrypted data only)
- ✅ Network interception (TLS + encrypted payloads)
- ✅ Unauthorized access (per-machine keys + RBAC)
- ✅ Insider threats (granular permissions)

nvolt does NOT protect against:

- ❌ Compromised developer machines (private key exposure)
- ❌ Malicious CLI binary (verify source before installing)

---

## Configuration

Configuration stored at `~/.nvolt/config.json`:

```json
{
  "private_key": "-----BEGIN RSA PRIVATE KEY-----...",
  "jwt_token": "eyJhbGciOiJIUzI1NiIs...",
  "machine_id": "m-abc123",
  "active_org": "org-xyz789",
  "server_url": "https://api.nvolt.io"
}
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup nvolt
        run: |
          curl -sL https://install.nvolt.io/cli | bash
          mkdir -p ~/.nvolt
          echo "${{ secrets.NVOLT_PRIVATE_KEY }}" > ~/.nvolt/private_key.pem

      - name: Deploy with secrets
        run: nvolt run -p my-app -e production -c "./deploy.sh"
```

**Setup**:
1. On your local machine: `nvolt machine add github-actions`
2. Copy the displayed private key to GitHub Secrets as `NVOLT_PRIVATE_KEY`
3. The CI machine will use silent login automatically

---

## Self-Hosting

nvolt CLI works with any compatible server implementation. See the [nvolt-server](https://github.com/yourusername/nvolt-server) repository for self-hosting instructions.

Cloud-hosted version available at [nvolt.io](https://nvolt.io) with additional features:
- Advanced audit logging
- SSO/SAML integration
- Multi-region replication
- Compliance certifications

---

## Roadmap

- [ ] Homebrew/apt package distribution
- [ ] Secret versioning and rollback
- [ ] Automated key rotation
- [ ] Hardware security module (HSM) support
- [ ] Secret sharing via time-limited tokens
- [ ] Desktop GUI for non-technical users

---

## Contributing

Contributions are welcome! Please note:

- This project is licensed under MIT with Commons Clause (non-commercial use)
- For commercial licensing inquiries: contact@nvolt.io
- Bug reports and feature requests: [GitHub Issues](https://github.com/yourusername/nvolt-cli/issues)

---

## License

Licensed under [MIT License with Commons Clause](LICENSE) — free for non-commercial use.

For commercial licensing or the cloud-hosted version with additional features, visit [nvolt.io](https://nvolt.io).

---

## Support

- **Documentation**: [docs.nvolt.io](https://docs.nvolt.io)
- **Community**: [GitHub Discussions](https://github.com/yourusername/nvolt-cli/discussions)
- **Issues**: [GitHub Issues](https://github.com/yourusername/nvolt-cli/issues)
- **Email**: support@nvolt.io

---

<div align="center">

**Built with ❤️ for developers who care about security**

</div>
