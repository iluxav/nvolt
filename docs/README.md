# nvolt Documentation

Welcome to the nvolt documentation! This guide will help you master secure secret management for your development workflow.

## 📖 Table of Contents

### Getting Started

- **[Getting Started Guide](getting-started.md)** — Install, configure, and push your first secrets

### Command Reference

Complete documentation for all CLI commands:

- **[Authentication](commands/authentication.md)** — Login and machine authentication
  - `nvolt login`
  - Silent login for CI/CD
  - Challenge-response protocol
  
- **[Secrets Management](commands/secrets.md)** — Push, pull, and run commands
  - `nvolt push` — Upload encrypted secrets
  - `nvolt pull` — Download and decrypt secrets
  - `nvolt run` — Execute commands with secrets injected
  - `nvolt sync` — Synchronize encryption keys
  - `nvolt set` — Configure defaults
  
- **[Machines](commands/machines.md)** — Manage cryptographic machine identities
  - `nvolt machine add` — Generate keys for new machines
  - `nvolt machine list` — List all machines in organization
  
- **[Organizations](commands/organization.md)** — Multi-organization management
  - `nvolt org` — View active organization
  - `nvolt org set` — Switch organizations
  
- **[Users](commands/users.md)** — User management and permissions (admin only)
  - `nvolt user add` — Add users to organization
  - `nvolt user list` — List organization users
  - `nvolt user mod` — Modify user permissions
  - `nvolt user rm` — Remove users from organization

### Guides

- **[CI/CD Integration](ci-cd-integration.md)** — Integrate nvolt into your pipelines
  - GitHub Actions
  - GitLab CI
  - CircleCI
  - Jenkins
  - Bitbucket Pipelines
  - Docker

- **[Security Model](security-model.md)** — Deep dive into cryptography and security
  - Wrapped key encryption
  - Challenge-response authentication
  - Threat model
  - Best practices
  - Compliance (SOC 2, GDPR, HIPAA, PCI DSS)

- **[Troubleshooting](troubleshooting.md)** — Common issues and solutions
  - Installation issues
  - Authentication problems
  - Permission errors
  - CI/CD debugging

---

## 🚀 Quick Start

```bash
# Install nvolt
curl -fsSL https://install.nvolt.io/latest/install.sh | bash

# Login
nvolt login

# Push secrets
nvolt push -f .env.local -p my-app -e development

# Pull secrets
nvolt pull -p my-app -e development

# Run with secrets
nvolt run -p my-app -e development -c "npm start"
```

---

## 📋 Common Workflows

### Development Workflow

```bash
# One-time setup
nvolt login
nvolt pull -f .env.local -p my-app -e development

# Daily usage
nvolt run -p my-app -e development -c "npm run dev"
```

### Team Onboarding

```bash
# Admin adds new team member
nvolt machine add alice-laptop
nvolt user add alice@example.com

# Alice sets up
nvolt login
nvolt pull -p my-app -e development
```

### CI/CD Setup

```bash
# Generate CI machine key
nvolt machine add github-actions

# Store key in GitHub Secrets as NVOLT_PRIVATE_KEY

# Use in workflow
nvolt run -p my-app -e production -c "./deploy.sh"
```

### Production Deployment

```bash
# On production server
nvolt login --silent --machine prod-server --org org-xyz
nvolt run -p my-app -e production -c "./deploy.sh"
```

---

## 🔐 Security Highlights

### End-to-End Encryption

```
Your Machine                Server              Team Machine
    │                         │                      │
    │  1. Encrypt secrets     │                      │
    ├────────────────────────>│                      │
    │                         │                      │
    │                    2. Store encrypted          │
    │                         │                      │
    │                         │  3. Request secrets  │
    │                         │<─────────────────────┤
    │                         │                      │
    │                         │  4. Send encrypted   │
    │                         ├─────────────────────>│
    │                         │                      │
    │                         │         5. Decrypt   │
    │                         │                      │
```

**The server never sees plaintext secrets or private keys.**

### Zero-Trust Principles

✅ Client-side encryption (AES-256-GCM)  
✅ Per-machine key pairs (RSA-2048)  
✅ Wrapped key architecture  
✅ Challenge-response authentication  
✅ Role-based access control (RBAC)  
✅ Environment-level permissions  

---

## 🎯 Use Cases

### For Developers

- **Local Development:** Pull secrets for .env files
- **Team Collaboration:** Share secrets securely across team
- **Environment Management:** Separate dev/staging/production secrets

### For DevOps

- **CI/CD Pipelines:** Silent authentication for automated deployments
- **Infrastructure as Code:** Inject secrets into Terraform, Ansible
- **Container Deployments:** Provide secrets to Docker/Kubernetes

### For Organizations

- **Compliance:** Meet SOC 2, HIPAA, PCI DSS requirements
- **Audit Trail:** Track who accessed which secrets when
- **Access Control:** Granular permissions per user/environment
- **Zero-Trust Security:** No plaintext secrets on server

---

## 💡 Key Concepts

### Scoping Hierarchy

```
Organization
  └── User (admin or dev role)
      └── Machine (unique cryptographic identity)
          └── Project
              └── Environment (dev, staging, prod, etc.)
                  └── Secrets (encrypted with wrapped keys)
```

### Wrapped Keys

Each secret has:
- **Symmetric key** (AES-256) for encrypting the secret value
- **Wrapped keys** (RSA-encrypted) for each authorized machine

Only machines with the corresponding private key can decrypt.

### Permissions

Three-tier access control:
1. **Organization Role:** admin or dev
2. **Project Access:** (reserved for future)
3. **Environment Permissions:** read, write, delete

---

## 📚 Additional Resources

- **[Main README](../README.md)** — Project overview and quick start
- **[GitHub Issues](https://github.com/yourusername/nvolt-cli/issues)** — Bug reports and feature requests
- **[GitHub Discussions](https://github.com/yourusername/nvolt-cli/discussions)** — Community support
- **[nvolt.io](https://nvolt.io)** — Cloud-hosted version
- **[Documentation](https://docs.nvolt.io)** — Online documentation

---

## 🆘 Need Help?

### Before Asking

1. Check [Troubleshooting](troubleshooting.md)
2. Search [GitHub Issues](https://github.com/yourusername/nvolt-cli/issues)
3. Review relevant command documentation

### Get Support

- **Community:** [GitHub Discussions](https://github.com/yourusername/nvolt-cli/discussions)
- **Email:** support@nvolt.io
- **Bug Reports:** [GitHub Issues](https://github.com/yourusername/nvolt-cli/issues)

---

## 🔄 Documentation Updates

This documentation is updated regularly. Last updated: October 2024

Found an issue? [Submit a PR](https://github.com/yourusername/nvolt-cli/pulls) or [open an issue](https://github.com/yourusername/nvolt-cli/issues).

---

<div align="center">

**[⬆ Back to Top](#nvolt-documentation)**

**Built with ❤️ by developers, for developers**

</div>

