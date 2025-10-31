# Getting Started with nvolt

Welcome to nvolt! This guide will help you install, configure, and start using nvolt to manage your environment variables securely.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Initial Setup](#initial-setup)
- [Your First Secret](#your-first-secret)
- [Running Commands with Secrets](#running-commands-with-secrets)
- [Next Steps](#next-steps)

---

## Prerequisites

Before getting started with nvolt, ensure you have:

- **Operating System:** Linux, macOS, or WSL2 on Windows
- **Go:** Version 1.19+ (only if building from source)
- **Internet Connection:** For OAuth authentication and server communication
- **Git Repository:** (Optional) nvolt auto-detects project names from git repos

---

## Installation

### Using Install Script (Recommended)

The easiest way to install nvolt is using our installation script:

```bash
curl -fsSL https://install.nvolt.io/latest/install.sh | bash
```

This script will:
1. Download the latest nvolt binary for your platform
2. Install it to `/usr/local/bin` (or `~/.local/bin` if no sudo access)
3. Make it executable
4. Verify the installation

### Manual Installation

If you prefer manual installation:

1. **Download the binary** from the [releases page](https://github.com/yourusername/nvolt-cli/releases)
2. **Extract and move it:**
   ```bash
   tar -xzf nvolt-linux-amd64.tar.gz
   sudo mv nvolt /usr/local/bin/
   sudo chmod +x /usr/local/bin/nvolt
   ```

### Building from Source

For developers who want to build from source:

```bash
# Clone the repository
git clone https://github.com/yourusername/nvolt-cli.git
cd nvolt-cli
go build -o nvolt cli/main.go
sudo mv nvolt /usr/local/bin/
```

### Verify Installation

Confirm nvolt is installed correctly:

```bash
nvolt --version
```

You should see output like:
```
nvolt version dev
```

---

## Initial Setup

### 1. Login

Start by authenticating with nvolt:

```bash
nvolt login
```

**What happens:**
1. A browser window opens for OAuth authentication (Google, GitHub, etc.)
2. nvolt generates a unique RSA key pair (2048-bit) for your machine
3. Your **private key** is stored locally at `~/.nvolt/config.json`
4. Your **public key** is sent to the server
5. An organization is automatically created for you (or you join an existing one)

**Output:**
```
Logging in...
Waiting for authentication...
..................
✓ Successfully authenticated!
```

### 2. Verify Your Configuration

Check your nvolt configuration:

```bash
cat ~/.nvolt/config.json
```

You should see:
```json
{
  "jwt_token": "eyJhbGciOiJIUzI1NiIs...",
  "machine_id": "m-abc123",
  "active_org": "org-xyz789",
  "server_url": "https://api.nvolt.io"
}
```

Your private key is stored separately at:
```
~/.nvolt/private_key.pem
```

> ⚠️ **Security Warning:** Never commit `~/.nvolt/` to version control!

### 3. Check Your Organization

View your active organization:

```bash
nvolt org
```

Output:
```
→ Active Organization: My Personal Org (admin)
```

If you belong to multiple organizations, you can switch between them:

```bash
nvolt org set
```

---

## Your First Secret

Let's push your first environment variables to nvolt.

### Option 1: Push from a .env File

If you have an existing `.env` file:

```bash
# Example .env file
cat > .env.local << EOF
DATABASE_URL=postgres://localhost:5432/mydb
API_KEY=sk_test_123456789
REDIS_URL=redis://localhost:6379
EOF

# Push to nvolt
nvolt push -f .env.local -p my-app -e development
```

**Flags explained:**
- `-f .env.local`: Source file containing environment variables
- `-p my-app`: Project name (defaults to git repo name if omitted)
- `-e development`: Environment name (defaults to `default` if omitted)

**Output:**
```
→ Project: my-app
→ Environment: development
→ Active Organization: My Personal Org (admin) [org-xyz789]
→ Machine Key ID: m-abc123
→ Mode: Full replacement (all existing variables will be replaced)
→ Variables to push: 3

Pushing secrets to server...

✓ Successfully pushed secrets!
→ 3 variables are now securely stored
```

### Option 2: Push Individual Keys

For one-off secrets, use the `-k` flag:

```bash
nvolt push -k DATABASE_URL=postgres://prod.db:5432/app \
           -k API_KEY=sk_live_987654321 \
           -p my-app -e production
```

**Note:** Using `-k` performs a **partial update** (only specified keys are added/updated), while `-f` performs a **full replacement**.

---

## Pulling Secrets

Retrieve your encrypted secrets:

### Pull to Console

```bash
nvolt pull -p my-app -e development
```

Output:
```
✓ Successfully pulled 3 variable(s)!

Decrypted Variables:

API_KEY=sk_test_123456789
DATABASE_URL=postgres://localhost:5432/mydb
REDIS_URL=redis://localhost:6379
```

### Pull to File

Save secrets to a file:

```bash
nvolt pull -f .env.local -p my-app -e development
```

Output:
```
✓ Successfully pulled 3 variable(s)!
→ Saved to: .env.local
```

### Pull Specific Key

Retrieve a single secret:

```bash
nvolt pull -k DATABASE_URL -p my-app -e development
```

Output:
```
DATABASE_URL=postgres://localhost:5432/mydb
```

---

## Running Commands with Secrets

The recommended way to use nvolt is with the `run` command. This injects secrets directly into your command's environment **without creating any files**:

```bash
nvolt run -p my-app -e development -c "npm start"
```

**What happens:**
1. nvolt pulls and decrypts all secrets for the specified project/environment
2. Secrets are loaded into memory
3. Your command runs with secrets as environment variables
4. After execution, secrets are cleared from memory

### Examples

```bash
# Start Node.js server
nvolt run -p my-app -e production -c "node server.js"

# Run Python application
nvolt run -p my-app -e staging -c "python manage.py runserver"

# Execute shell script
nvolt run -p my-app -e development -c "./scripts/deploy.sh"

# Run Docker container with secrets
nvolt run -p my-app -e production -c "docker compose up"
```

### Why Use `nvolt run`?

✅ **Security:** Secrets never touch disk  
✅ **Simplicity:** No need to manage `.env` files  
✅ **Clean:** No secrets in shell history  
✅ **Audit-friendly:** All secret access is logged server-side  

---

## Next Steps

Congratulations! You've successfully set up nvolt and managed your first secrets. Here's what to explore next:

### Learn More Commands

- **[Authentication](commands/authentication.md)** — Silent login for CI/CD machines
- **[Secrets Management](commands/secrets.md)** — Advanced push/pull/run patterns
- **[Machines](commands/machines.md)** — Add machines for team members or CI/CD
- **[Organizations](commands/organization.md)** — Manage multi-org workflows
- **[Users](commands/users.md)** — Invite team members and set permissions (admin only)

### Set Up CI/CD

Learn how to use nvolt in automated pipelines:

- **[CI/CD Integration](ci-cd-integration.md)** — GitHub Actions, GitLab CI, CircleCI examples

### Understand the Security Model

Dive deep into how nvolt protects your secrets:

- **[Security Model](security-model.md)** — Wrapped keys, threat model, best practices

### Troubleshooting

Having issues? Check our troubleshooting guide:

- **[Troubleshooting](troubleshooting.md)** — Common problems and solutions

---

## Quick Reference Card

```bash
# Authentication
nvolt login                              # Interactive OAuth login
nvolt login --silent --machine ci --org org-xyz  # CI/CD silent login

# Push secrets
nvolt push -f .env -p my-app -e dev      # From file (full replacement)
nvolt push -k KEY=value -p my-app -e dev # Individual key (partial update)

# Pull secrets
nvolt pull -p my-app -e dev              # To console
nvolt pull -f .env -p my-app -e dev      # To file
nvolt pull -k KEY -p my-app -e dev       # Specific key

# Run commands
nvolt run -p my-app -e dev -c "npm start"

# Organization management
nvolt org                                # View active org
nvolt org set                            # Switch org (interactive)

# Machine management
nvolt machine add ci-runner              # Add new machine
nvolt machine list                       # List all machines

# User management (admin only)
nvolt user list                          # List users
nvolt user add user@example.com          # Add user
nvolt user mod user@example.com          # Modify permissions
nvolt user rm user@example.com           # Remove user

# Sync after adding machines
nvolt sync                               # Sync org-level master key
```

---

**Need help?** Join our [GitHub Discussions](https://github.com/yourusername/nvolt-cli/discussions) or email support@nvolt.io

