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

**Zero-Trust Secret Management for Modern Teams**

[![License](https://img.shields.io/badge/License-MIT%20with%20Commons%20Clause-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)](https://golang.org/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[Features](#-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [CI/CD](#-cicd-integration) • [Security](#-security)

</div>

---

## 🎯 Overview

**nvolt** is a CLI-first environment variable manager that brings **end-to-end encryption** to your secrets workflow. Built on Zero-Trust principles with client-side cryptography, your secrets are encrypted before they ever leave your machine—the server only stores encrypted data and never has access to your private keys.

### Why nvolt?

```bash
# Push secrets from .env file
nvolt push -f .env.production -p my-app -e production

# Run commands with secrets injected (no .env files!)
nvolt run -p my-app -e production -c "npm start"

# Silent authentication for CI/CD
nvolt login --silent --machine ci-runner --org org-xyz
```

- **🔐 Zero-Trust Security** — Client-side encryption with wrapped keys
- **👥 Team Collaboration** — Granular RBAC with admin and developer roles
- **🚀 Zero Friction** — Auto-detects project from git, seamless CLI UX
- **🔄 Multi-Environment** — Dev, staging, production, or custom environments
- **🤖 CI/CD Native** — Silent login via challenge-response protocol
- **📦 Self-Hostable** — No vendor lock-in, works with any server

---

## ✨ Features

| Feature                           | Description                                     |
| --------------------------------- | ----------------------------------------------- |
| **Wrapped Key Encryption**        | RSA key pairs + AES-GCM for hybrid cryptography |
| **Per-Machine Identity**          | Each device has unique cryptographic identity   |
| **Organization Management**       | Multi-tenant with role-based access control     |
| **Project & Environment Scoping** | Organize secrets hierarchically                 |
| **Memory-Only Injection**         | Secrets never touch disk with `nvolt run`       |
| **Interactive TUI**               | Beautiful terminal UI for exploring secrets     |

---

## 🚀 Quick Start

### Installation

```bash
# Install via script
curl -fsSL https://install.nvolt.io/latest/install.sh | bash

# Verify installation
nvolt --version
```

### First Steps

```bash
# 1. Login (opens browser for OAuth)
nvolt login

# 2. Push your first secrets
nvolt push -f .env.local -p my-app -e development

# 3. Pull secrets to verify
nvolt pull -p my-app -e development

# 4. Run commands with secrets injected
nvolt run -p my-app -e development -c "node server.js"
```

**That's it!** Your secrets are now encrypted end-to-end and synced across your team.

---

## 📚 Documentation

Comprehensive guides and API references:

- **[Getting Started](docs/getting-started.md)** — Installation, login, and first secrets
- **[Commands Reference](docs/commands/)** — Complete CLI command documentation
  - [Authentication](docs/commands/authentication.md) — Login and machine setup
  - [Secrets Management](docs/commands/secrets.md) — Push, pull, and run
  - [Machines](docs/commands/machines.md) — Add and list machines
  - [Organizations](docs/commands/organization.md) — Org management
  - [Users](docs/commands/users.md) — User permissions (admin only)
- **[CI/CD Integration](docs/ci-cd-integration.md)** — GitHub Actions, GitLab, CircleCI
- **[Security Model](docs/security-model.md)** — Cryptography and threat model
- **[Troubleshooting](docs/troubleshooting.md)** — Common issues and solutions

---

## 🔐 Security

nvolt uses **wrapped key encryption** for end-to-end security:

```
┌─────────────┐          ┌─────────────┐          ┌─────────────┐
│  Machine A  │          │   Server    │          │  Machine B  │
│             │          │  (Encrypted │          │             │
│ Private Key ├─────────▶│   Storage)  ├─────────▶│ Private Key │
│             │  Wrap    │             │  Unwrap  │             │
└─────────────┘          └─────────────┘          └─────────────┘
```

**What's Protected:**

- ✅ Server compromise (only encrypted data stored)
- ✅ Network interception (TLS + encrypted payloads)
- ✅ Unauthorized access (per-machine keys + RBAC)
- ✅ Insider threats (granular environment permissions)

**Threat Model:**

- ❌ Compromised developer machines (private keys exposed)
- ❌ Malicious CLI binary (verify GPG signature)

[Read the full security model →](docs/security-model.md)

---

## 🤖 CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy
on: [push]

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
          chmod 600 ~/.nvolt/private_key.pem

      - name: Deploy with secrets
        run: nvolt run -p my-app -e production -c "./deploy.sh"
```

**Setup:** Generate a CI machine key with `nvolt machine add github-actions`, then save the private key to GitHub Secrets.

[More CI/CD examples →](docs/ci-cd-integration.md)

---

## 🏗️ Architecture

### Scoping Hierarchy

```
Organization
  └── User (admin or dev role)
      └── Machine (unique cryptographic identity)
          └── Project
              └── Environment (dev, staging, prod, etc.)
                  └── Secrets (encrypted with wrapped keys)
```

### Global Flags

| Flag            | Short | Description      | Default                |
| --------------- | ----- | ---------------- | ---------------------- |
| `--project`     | `-p`  | Project name     | Git repo name          |
| `--environment` | `-e`  | Environment name | `default`              |
| `--org`         | `-o`  | Organization ID  | Active org from config |

---

## 🛠️ Advanced Usage

### Set Default Values

```bash
# Set default environment and organization
nvolt set -e production -o org-xyz789

# Configure custom server URL (for self-hosting)
nvolt set -s https://nvolt.mycompany.com
```

### Organization Switching

```bash
# Interactive org selector
nvolt org

# Switch to specific org
nvolt org set -o org-xyz789
```

### User Management (Admin Only)

```bash
# List users in organization
nvolt user list

# Add user with interactive permissions
nvolt user add john@example.com

# Modify user permissions
nvolt user mod john@example.com

# Remove user
nvolt user rm john@example.com
```

---

## 🗺️ Roadmap

- [ ] Homebrew formula and apt repository
- [ ] Secret versioning and rollback
- [ ] Automated key rotation policies
- [ ] HSM and hardware key support
- [ ] Time-limited secret sharing links
- [ ] Desktop GUI for non-technical users
- [ ] WASM-based web interface

---

## 🤝 Contributing

Contributions are welcome! For bug reports and feature requests, please use [GitHub Issues](https://github.com/yourusername/nvolt-cli/issues).

**Note:** nvolt CLI is licensed under MIT with Commons Clause (free for non-commercial use). For commercial licensing, contact: contact@nvolt.io

---

## 📄 License

Licensed under [MIT License with Commons Clause](LICENSE) — free for non-commercial use.

---

## 💬 Support

- **Documentation:** [docs/](docs/)
- **Issues:** [GitHub Issues](https://github.com/yourusername/nvolt-cli/issues)
- **Discussions:** [GitHub Discussions](https://github.com/yourusername/nvolt-cli/discussions)
- **Email:** support@nvolt.io

---

<div align="center">

**Built with ❤️ by developers, for developers**

[⭐ Star us on GitHub](https://github.com/yourusername/nvolt-cli) • [🐦 Follow on Twitter](https://twitter.com/nvolt_io)

</div>
