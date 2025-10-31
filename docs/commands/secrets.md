# Secrets Management Commands

Manage your environment variables securely with nvolt's push, pull, and run commands.

## Table of Contents

- [nvolt push](#nvolt-push)
- [nvolt pull](#nvolt-pull)
- [nvolt run](#nvolt-run)
- [nvolt sync](#nvolt-sync)
- [nvolt set](#nvolt-set)
- [Best Practices](#best-practices)

---

## nvolt push

Push environment variables to the nvolt server, encrypted end-to-end.

### Usage

```bash
# From file (full replacement)
nvolt push -f <file> -p <project> -e <environment>

# Individual keys (partial update)
nvolt push -k KEY1=value1 -k KEY2=value2 -p <project> -e <environment>
```

### Flags

| Flag | Short | Required | Description | Default |
|------|-------|----------|-------------|---------|
| `--file` | `-f` | * | Path to .env file | - |
| `--key` | `-k` | * | Individual key-value pair (can be repeated) | - |
| `--project` | `-p` | No | Project name | Git repo name |
| `--environment` | `-e` | No | Environment name | `default` |
| `--org` | `-o` | No | Organization ID | Active org from config |

\* Either `-f` or `-k` must be specified

### Examples

#### Push from .env File

```bash
# Push .env.production to production environment
nvolt push -f .env.production -p my-app -e production
```

**Output:**
```
→ Project: my-app
→ Environment: production
→ Active Organization: Acme Corp (admin) [org-xyz789]
→ Machine Key ID: m-abc123
→ Mode: Full replacement (all existing variables will be replaced)
→ Variables to push: 5

Pushing secrets to server...

✓ Successfully pushed secrets!
→ 5 variables are now securely stored
```

**Behavior:** This performs a **full replacement** — all existing secrets in the specified project/environment are deleted and replaced with the contents of the file.

#### Push Individual Keys

```bash
# Add or update specific keys
nvolt push -k DATABASE_URL=postgres://prod:5432/db \
           -k API_KEY=sk_live_123456 \
           -k STRIPE_SECRET=sk_test_789 \
           -p my-app -e production
```

**Behavior:** This performs a **partial update** — only the specified keys are added or updated. Other existing keys remain untouched.

#### Push to Multiple Environments

```bash
# Development
nvolt push -f .env.dev -p my-app -e development

# Staging
nvolt push -f .env.staging -p my-app -e staging

# Production
nvolt push -f .env.prod -p my-app -e production
```

#### Auto-Detect Project from Git

```bash
# If in a git repository, project name is auto-detected
cd /path/to/my-app
nvolt push -f .env -e development
```

### .env File Format

nvolt supports standard `.env` file syntax:

```bash
# Comments are supported
DATABASE_URL=postgres://localhost:5432/mydb
API_KEY=sk_test_123456789

# Quotes are optional
REDIS_URL="redis://localhost:6379"

# Multi-line values (quoted)
PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----"

# Empty values
OPTIONAL_VAR=

# Special characters
PASSWORD=p@ssw0rd!#$%
```

### What Happens Internally

When you push secrets:

1. **Fetch Public Keys:** CLI retrieves public keys for all machines with access
2. **Generate Symmetric Key:** A unique AES-256 key is generated for encrypting the secrets
3. **Wrap Symmetric Key:** The symmetric key is encrypted with each machine's public key
4. **Encrypt Secrets:** All secret values are encrypted with the symmetric key
5. **Upload:** Encrypted secrets + wrapped keys are sent to the server
6. **Server Storage:** Server stores only encrypted data (no plaintext)

### Permissions Required

To push secrets, you need:

- **Write permission** for the specified environment
- **Admin** or **dev** role in the organization

If you lack permissions, you'll see:

```
⚠ Permission Denied

You don't have write permission for this environment.
  Project: my-app
  Environment: production

Please contact your organization admin to grant you access.
```

---

## nvolt pull

Pull and decrypt environment variables from the server.

### Usage

```bash
# Pull to console
nvolt pull -p <project> -e <environment>

# Pull to file
nvolt pull -f <file> -p <project> -e <environment>

# Pull specific key
nvolt pull -k <key-name> -p <project> -e <environment>
```

### Flags

| Flag | Short | Required | Description | Default |
|------|-------|----------|-------------|---------|
| `--file` | `-f` | No | Output file path | - (prints to console) |
| `--key` | `-k` | No | Pull specific key only | - (pulls all keys) |
| `--project` | `-p` | No | Project name | Git repo name |
| `--environment` | `-e` | No | Environment name | `default` |
| `--org` | `-o` | No | Organization ID | Active org from config |

### Examples

#### Pull to Console

```bash
nvolt pull -p my-app -e development
```

**Output:**
```
→ Project: my-app
→ Environment: development
→ Active Organization: Acme Corp (admin) [org-xyz789]
→ Machine Key ID: m-abc123

✓ Successfully pulled 5 variable(s)!

Decrypted Variables:

API_KEY=sk_test_123456789
DATABASE_URL=postgres://localhost:5432/mydb
REDIS_URL=redis://localhost:6379
SECRET_TOKEN=abcdef123456
STRIPE_KEY=pk_test_xyz789
```

#### Pull to File

```bash
nvolt pull -f .env.local -p my-app -e development
```

**Output:**
```
✓ Successfully pulled 5 variable(s)!
→ Saved to: .env.local
```

**File contents:**
```bash
API_KEY=sk_test_123456789
DATABASE_URL=postgres://localhost:5432/mydb
REDIS_URL=redis://localhost:6379
SECRET_TOKEN=abcdef123456
STRIPE_KEY=pk_test_xyz789
```

#### Pull Specific Key

```bash
nvolt pull -k DATABASE_URL -p my-app -e production
```

**Output:**
```
✓ Successfully pulled 1 variable(s)!

DATABASE_URL=postgres://prod.db:5432/myapp
```

**Use case:** Quickly retrieve a single secret without exposing all others.

### What Happens Internally

When you pull secrets:

1. **Request Secrets:** CLI requests encrypted secrets for the specified scope
2. **Fetch Wrapped Key:** Server sends wrapped symmetric key for your machine
3. **Unwrap Key:** CLI decrypts the wrapped key using your private key
4. **Decrypt Secrets:** All secret values are decrypted using the symmetric key
5. **Output:** Secrets are displayed or written to file

### Permissions Required

To pull secrets, you need:

- **Read permission** for the specified environment
- **Admin** or **dev** role in the organization

### Troubleshooting

**No secrets found:**

```
⚠ No variables found for this scope.
If you just logged in, please run the command `nvolt sync` from any authorized machine to sync your machine.
```

**Solution:** If you're a new machine in the organization, another authorized machine must run `nvolt sync` to re-wrap keys for you.

**No wrapped key for this machine:**

```
⚠ No secrets found for this machine

This machine hasn't been synced yet. To enable access:
  1. Run 'nvolt sync' from ANY authorized machine, OR
  2. Run 'nvolt push' from any machine to sync all machines
```

**Solution:** See [nvolt sync](#nvolt-sync) below.

---

## nvolt run

Run a command with environment variables injected from nvolt. **This is the recommended way to use nvolt in production.**

### Usage

```bash
nvolt run -p <project> -e <environment> -c "<command>"

# Alternative: command as arguments
nvolt run -p <project> -e <environment> <command>
```

### Flags

| Flag | Short | Required | Description | Default |
|------|-------|----------|-------------|---------|
| `--command` | `-c` | * | Command to execute | - |
| `--project` | `-p` | No | Project name | Git repo name |
| `--environment` | `-e` | No | Environment name | `default` |
| `--org` | `-o` | No | Organization ID | Active org from config |

\* Command can be provided via `-c` flag or as positional arguments

### Examples

#### Run Node.js Server

```bash
nvolt run -p my-app -e production -c "node server.js"
```

#### Run Python Application

```bash
nvolt run -p my-app -e staging -c "python manage.py runserver"
```

#### Run Shell Script

```bash
nvolt run -p my-app -e development -c "./scripts/start.sh"
```

#### Run Docker Compose

```bash
nvolt run -p my-app -e production -c "docker compose up"
```

#### Command as Arguments (no -c flag)

```bash
nvolt run -p my-app -e development npm start
```

### What Happens Internally

When you run a command:

1. **Pull Secrets:** All secrets for the specified project/environment are fetched and decrypted
2. **Prepare Environment:** Secrets are added to the command's environment variables
3. **Execute Command:** Your command runs with secrets loaded in memory
4. **Clean Memory:** After execution, secrets are cleared from memory

### Security Benefits

✅ **Secrets never touch disk** — No `.env` files created  
✅ **No shell history pollution** — Secrets aren't passed as command-line args  
✅ **Memory-only storage** — Secrets are cleared after execution  
✅ **Audit trail** — All secret access is logged server-side  
✅ **No accidental commits** — Can't commit what doesn't exist on disk  

### Output

```bash
$ nvolt run -p my-app -e production -c "npm start"

🔐 nvolt run
→ Project: my-app
→ Environment: production
→ Command: npm start

Pulling secrets from server...
✓ Successfully pulled 8 secret(s)!

Executing command...
─────────────────────────────────────────────────────
> my-app@1.0.0 start
> node server.js

Server listening on port 3000
```

### Exit Codes

The exit code of `nvolt run` matches the exit code of your command:

```bash
nvolt run -p my-app -e dev -c "exit 42"
echo $?  # Outputs: 42
```

This ensures compatibility with CI/CD systems that rely on exit codes.

---

## nvolt sync

Synchronize encryption keys across all machines in an organization. This command is essential after adding new machines or users.

### Usage

```bash
nvolt sync
```

### When to Use

Run `nvolt sync` after:

- Adding a new machine with `nvolt machine add`
- Adding a new user with `nvolt user add`
- A team member logs in for the first time
- Removing a machine or user (to revoke access)

### What It Does

With nvolt's simplified org-level encryption model:

1. **Org Master Key:** A single master encryption key exists per organization
2. **Re-wrap for All Machines:** The master key is re-wrapped with each machine's public key
3. **Universal Access:** All machines can now decrypt all secrets in the organization

This is more efficient than the older per-project/per-environment syncing model.

### Example

```bash
$ nvolt sync

🔄 Syncing org-level master key...

✓ Keys synchronized successfully!
→ All machines in your organization can now access all secrets
```

### Permissions Required

Any authenticated user in the organization can run `nvolt sync`.

### Background

Previously, nvolt required syncing each project/environment combination individually. The new model uses a single org-level master key that's wrapped for each machine, simplifying key management.

---

## nvolt set

Set default values for project, environment, organization, or server URL.

### Usage

```bash
nvolt set -e <environment>
nvolt set -o <org-id>
nvolt set -s <server-url>
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--environment` | `-e` | Set default environment |
| `--org` | `-o` | Set active organization |
| `--server-url` | `-s` | Set custom server URL (for self-hosting) |

At least one flag must be specified.

### Examples

#### Set Default Environment

```bash
nvolt set -e production
```

Now all commands will use `production` unless overridden:

```bash
# Uses production environment
nvolt pull -p my-app

# Override to use staging
nvolt pull -p my-app -e staging
```

#### Set Active Organization

```bash
nvolt set -o org-xyz789
```

#### Configure Self-Hosted Server

```bash
nvolt set -s https://nvolt.mycompany.com
```

#### Set Multiple Values

```bash
nvolt set -e production -o org-xyz789
```

**Output:**
```
✓ Default environment set to: production
✓ Active organization set to: org-xyz789
```

---

## Best Practices

### 1. Use `nvolt run` for Production

❌ **Don't:**
```bash
nvolt pull -f .env.production
node server.js
```

✅ **Do:**
```bash
nvolt run -p my-app -e production -c "node server.js"
```

**Why:** Secrets stay in memory only and are never written to disk.

### 2. Use Project and Environment Scoping

Organize secrets by:

- **Project:** Application name (e.g., `web-app`, `api`, `mobile`)
- **Environment:** Deployment stage (e.g., `dev`, `staging`, `production`)

```bash
nvolt push -f .env.dev -p web-app -e development
nvolt push -f .env.staging -p web-app -e staging
nvolt push -f .env.prod -p web-app -e production
```

### 3. Limit Production Access

Grant production write access only to trusted team members:

```bash
# Add user with read-only production access
nvolt user add junior@example.com -p web-app -e production \
  -a read=true,write=false,delete=false
```

### 4. Sync After Adding Machines

Always run sync after adding machines:

```bash
nvolt machine add ci-runner
nvolt sync
```

### 5. Use Specific Keys for Debugging

Avoid exposing all secrets when debugging:

```bash
# Only pull the specific secret you need
nvolt pull -k DATABASE_URL -p my-app -e production
```

### 6. Rotate Keys Periodically

For enhanced security:

```bash
# Quarterly rotation
nvolt machine add my-laptop-2024-Q1
nvolt sync

# Delete old machine key from server
nvolt machine rm my-laptop-2023-Q4
```

### 7. Audit Regularly

Check who has access:

```bash
nvolt user list
nvolt machine list
```

---

## Common Workflows

### Development Workflow

```bash
# 1. Clone repository
git clone https://github.com/mycompany/my-app.git
cd my-app

# 2. Login to nvolt
nvolt login

# 3. Pull development secrets
nvolt pull -f .env.local -e development

# 4. Start development server
nvolt run -e development -c "npm run dev"
```

### Production Deployment

```bash
# 1. SSH into production server
ssh prod-server

# 2. Authenticate with silent login
nvolt login --silent --machine prod-server --org org-xyz

# 3. Deploy with secrets
nvolt run -p my-app -e production -c "./deploy.sh"
```

### Team Onboarding

```bash
# On admin machine:
# 1. Add new team member's machine
nvolt machine add alice-laptop

# 2. Sync keys
nvolt sync

# 3. Invite user to org
nvolt user add alice@example.com

# On Alice's machine:
# 1. Login
nvolt login

# 2. Verify access
nvolt pull -p my-app -e development
```

---

## Related Commands

- **[nvolt machine add](machines.md#nvolt-machine-add)** — Add new machines
- **[nvolt user add](users.md#nvolt-user-add)** — Add users with permissions
- **[nvolt org set](organization.md#nvolt-org-set)** — Switch organizations

---

## Next Steps

- Learn about [Machine Management](machines.md)
- Set up [CI/CD Integration](../ci-cd-integration.md)
- Understand the [Security Model](../security-model.md)

---

[← Back to Documentation Home](../README.md)

