# Machine Management Commands

Manage cryptographic machine identities for team members and CI/CD runners.

## Table of Contents

- [Overview](#overview)
- [nvolt machine add](#nvolt-machine-add)
- [nvolt machine list](#nvolt-machine-list)
- [Common Workflows](#common-workflows)
- [Security Considerations](#security-considerations)

---

## Overview

In nvolt, a **machine** represents any device or compute environment with its own cryptographic identity:

- Developer laptops/workstations
- CI/CD runners (GitHub Actions, GitLab CI, CircleCI)
- Production servers
- Staging environments
- Docker containers

Each machine has:
- **Unique identifier** (e.g., `m-abc123def456` or custom name like `github-actions`)
- **RSA key pair** (2048-bit)
- **Public key** registered on the server
- **Private key** stored locally on the machine

### Why Per-Machine Keys?

✅ **Granular Access Control:** Revoke access for a single compromised machine  
✅ **Audit Trail:** Track which machine accessed which secrets  
✅ **Zero-Trust:** Each machine must prove its identity independently  
✅ **Key Rotation:** Rotate keys per-machine without affecting others  

---

## nvolt machine add

Generate a new RSA key pair for an additional machine (typically for CI/CD or new team members).

### Usage

```bash
nvolt machine add <machine-name>
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `machine-name` | Yes | Unique identifier for the machine (alphanumeric, hyphens allowed) |

### What It Does

When you run `nvolt machine add`:

1. **Generate Key Pair:** Creates a new RSA-2048 key pair
2. **Register Public Key:** Sends public key to server
3. **Save Private Key:** Writes private key to `<machine-name>_key.pem` in current directory
4. **Auto-Sync:** Automatically re-wraps all existing secrets for the new machine

### Examples

#### Add a CI/CD Machine

```bash
nvolt machine add github-actions
```

**Output:**
```
🔑 Generating keys for machine: github-actions
→ Saving public key to server...
✓ Public key saved to server
✓ Private key saved to: /home/user/github-actions_key.pem
→ File permissions set to 600 (owner read/write only)

📋 Next Steps
1. Securely transfer the key file to the destination machine:
   scp github-actions_key.pem user@destination:~/.nvolt/private_key.pem

   OR use pbcopy to copy and paste manually:
   cat github-actions_key.pem | pbcopy
   # Then on destination machine:
   pbpaste > ~/.nvolt/private_key.pem && chmod 600 ~/.nvolt/private_key.pem

2. On the destination machine, authenticate:
   nvolt login --silent --machine github-actions --org org-xyz

⚠️  Remember to delete the local key file after transfer: rm github-actions_key.pem

🔄 Syncing keys for all machines...
✓ Synced 3/3 project/environment combination(s)
✓ Machine 'github-actions' is ready! All machines can now access secrets.
```

#### Add a Team Member's Machine

```bash
nvolt machine add alice-laptop
```

Then securely send `alice-laptop_key.pem` to Alice via:
- Encrypted email
- Password manager (1Password, Bitwarden)
- Secure file sharing (Signal, encrypted cloud storage)

#### Add a Production Server

```bash
nvolt machine add prod-server-01
```

Transfer the key securely:
```bash
# Using SCP over SSH
scp prod-server-01_key.pem admin@prod.example.com:~/.nvolt/private_key.pem

# Using rsync
rsync -avz -e ssh prod-server-01_key.pem admin@prod.example.com:~/.nvolt/private_key.pem

# Then on the server, set permissions
ssh admin@prod.example.com "chmod 600 ~/.nvolt/private_key.pem"
```

### File Output

The private key is saved to:
```
<machine-name>_key.pem
```

Example `github-actions_key.pem`:
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAyF2PQ3K5x6wN...
... (many lines) ...
-----END RSA PRIVATE KEY-----
```

**Security:**
- File permissions are automatically set to `600` (owner read/write only)
- **Delete this file after transferring it to the destination machine**
- Never commit `.pem` files to version control

### Automatic Key Syncing

After adding a machine, nvolt automatically:

1. Fetches all project/environment combinations in your organization
2. Re-wraps the master encryption key for all machines (including the new one)
3. Ensures the new machine can access all existing secrets

**Output:**
```
🔄 Syncing keys for all machines...
✓ Synced 3/3 project/environment combination(s)
✓ Machine 'github-actions' is ready! All machines can now access secrets.
```

### Error Handling

**Machine name already exists:**
```
failed to save public key: machine 'github-actions' already exists
```

**Solution:** Choose a different name or delete the old machine first.

**Permission denied:**
```
failed to save public key: permission denied
```

**Solution:** Only organization admins can add machines. Contact your admin.

---

## nvolt machine list

List all machines in an organization with their details.

### Usage

```bash
nvolt machine list
```

**Aliases:** `nvolt machine ls`

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--org` | `-o` | Organization ID | Active org from config |

### Example

```bash
$ nvolt machine list

Organization Machines
→ Organization: Acme Corp (org-xyz789)
→ Fetching machines...

Found 5 machine(s):

┌─────────────────────┬──────────────┬─────────────────────┬─────────────────────┬─────────────────────┐
│ Machine Name        │ User Name    │ User Email          │ Created At          │ Machine ID          │
├─────────────────────┼──────────────┼─────────────────────┼─────────────────────┼─────────────────────┤
│ alice-laptop        │ Alice Smith  │ alice@example.com   │ 2024-10-01 14:23:11 │ m-abc123def456      │
│ bob-workstation     │ Bob Johnson  │ bob@example.com     │ 2024-10-05 09:15:42 │ m-def456ghi789      │
│ github-actions      │ Admin User   │ admin@example.com   │ 2024-10-10 11:30:55 │ github-actions      │
│ gitlab-ci           │ Admin User   │ admin@example.com   │ 2024-10-12 16:45:22 │ gitlab-ci           │
│ prod-server-01      │ DevOps Team  │ devops@example.com  │ 2024-10-15 08:12:37 │ prod-server-01      │
└─────────────────────┴──────────────┴─────────────────────┴─────────────────────┴─────────────────────┘
```

### Output Details

Each row shows:

- **Machine Name:** Custom name or auto-generated ID
- **User Name:** Name of the user who created the machine
- **User Email:** Email of the user who created the machine
- **Created At:** Timestamp of machine creation
- **Machine ID:** Unique identifier for the machine

### Filtering by Organization

```bash
# List machines in specific organization
nvolt machine list -o org-abc123
```

### Use Cases

- **Security Audit:** Review all machines with access to secrets
- **Access Revocation:** Identify machines to remove (coming soon)
- **Troubleshooting:** Verify a machine exists before debugging access issues
- **Compliance:** Document all devices with production access

---

## Common Workflows

### Workflow 1: Onboard a New Team Member

**On admin's machine:**

```bash
# 1. Add machine for new team member
nvolt machine add alice-laptop

# 2. Securely send alice-laptop_key.pem to Alice
# Use encrypted email, password manager, or secure file transfer

# 3. Add Alice as a user (if not already in org)
nvolt user add alice@example.com
```

**On Alice's machine:**

```bash
# 1. Install nvolt
curl -fsSL https://install.nvolt.io/latest/install.sh | bash

# 2. Save private key
mkdir -p ~/.nvolt
# (Copy alice-laptop_key.pem to ~/.nvolt/private_key.pem)
chmod 600 ~/.nvolt/private_key.pem

# 3. Authenticate with silent login
nvolt login --silent --machine alice-laptop --org org-xyz789

# 4. Verify access
nvolt pull -p my-app -e development
```

### Workflow 2: Set Up GitHub Actions

**Step 1: Generate key on your local machine**

```bash
nvolt machine add github-actions
```

**Step 2: Add private key to GitHub Secrets**

```bash
# Copy private key to clipboard
cat github-actions_key.pem | pbcopy  # macOS
cat github-actions_key.pem | xclip -selection clipboard  # Linux
```

Then in GitHub:
1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Name: `NVOLT_PRIVATE_KEY`
4. Value: Paste the private key
5. Click **Add secret**

**Step 3: Add org ID to GitHub Secrets**

1. Get your org ID: `cat ~/.nvolt/config.json | grep active_org`
2. Add secret: Name `NVOLT_ORG_ID`, Value `org-xyz789`

**Step 4: Delete local key file**

```bash
rm github-actions_key.pem
```

**Step 5: Configure workflow**

`.github/workflows/deploy.yml`:
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

### Workflow 3: Rotate Machine Keys

For enhanced security, rotate keys quarterly:

```bash
# 1. Add new machine with versioned name
nvolt machine add my-laptop-2024-Q4

# 2. Transfer key to your machine
cp my-laptop-2024-Q4_key.pem ~/.nvolt/private_key.pem
chmod 600 ~/.nvolt/private_key.pem

# 3. Authenticate
nvolt login --silent --machine my-laptop-2024-Q4 --org org-xyz

# 4. Verify access
nvolt pull -p my-app -e production

# 5. Remove old machine (future feature)
# nvolt machine rm my-laptop-2024-Q3

# 6. Clean up
rm my-laptop-2024-Q4_key.pem
```

### Workflow 4: Emergency Access Revocation

If a machine is compromised:

**Immediate actions:**

```bash
# 1. From another machine, generate new machine key
nvolt machine add emergency-access

# 2. Re-sync all secrets (revokes old machine's access)
nvolt sync

# 3. Optionally: Change all secrets
nvolt push -f .env.new -p my-app -e production
```

**Future feature:** `nvolt machine rm <machine-name>` will directly revoke access.

---

## Security Considerations

### Best Practices

✅ **One key per machine** — Never copy the same private key to multiple machines  
✅ **Delete after transfer** — Remove `.pem` files after securely transferring them  
✅ **Secure transfer methods** — Use encrypted channels (SSH, encrypted email, password managers)  
✅ **Audit regularly** — Run `nvolt machine list` monthly to review access  
✅ **Rotate keys** — Generate new keys every 90 days for sensitive environments  
✅ **Limit CI/CD scope** — Use separate machines for dev vs prod CI/CD  

### Private Key Storage

**Where private keys are stored:**

- **Interactive login:** `~/.nvolt/config.json` (generated automatically)
- **Silent login:** `~/.nvolt/private_key.pem` (transferred from admin)

**Protecting private keys:**

```bash
# Ensure proper permissions
chmod 600 ~/.nvolt/config.json
chmod 600 ~/.nvolt/private_key.pem

# Encrypt your home directory (Linux)
sudo apt install ecryptfs-utils
ecryptfs-migrate-home -u $USER

# FileVault encryption (macOS)
# System Preferences → Security & Privacy → FileVault → Turn On

# BitLocker encryption (Windows)
# Settings → Update & Security → Device encryption
```

### Machine Naming Conventions

Use descriptive, unique names:

✅ **Good:**
- `alice-laptop`
- `github-actions-prod`
- `gitlab-ci-staging`
- `prod-server-us-east-1`
- `dev-container-bob`

❌ **Bad:**
- `machine1`, `machine2` (not descriptive)
- `my-machine` (ambiguous)
- `test` (unclear purpose)

### Threat Model

| Scenario | Risk | Mitigation |
|----------|------|------------|
| Laptop stolen | ⚠️ High | Full disk encryption, revoke machine access |
| CI/CD secret leaked | ⚠️ High | Rotate CI machine key, use separate keys per environment |
| Admin machine compromised | 🔴 Critical | Revoke all machines, regenerate all keys, rotate all secrets |
| Former employee | ⚠️ Medium | Remove user via `nvolt user rm`, their machines auto-revoked |

---

## Related Commands

- **[nvolt login --silent](authentication.md#silent-login-for-cicd)** — Authenticate a newly provisioned machine
- **[nvolt sync](secrets.md#nvolt-sync)** — Re-wrap keys after adding machines
- **[nvolt user add](users.md#nvolt-user-add)** — Add users to organization

---

## Next Steps

- Learn about [User Management](users.md)
- Set up [CI/CD Integration](../ci-cd-integration.md)
- Understand the [Security Model](../security-model.md)

---

[← Back to Documentation Home](../README.md)

