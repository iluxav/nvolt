# Organization Management Commands

Manage organizations and switch between multiple organizations.

## Table of Contents

- [Overview](#overview)
- [nvolt org](#nvolt-org)
- [nvolt org set](#nvolt-org-set)
- [Multi-Organization Workflows](#multi-organization-workflows)
- [Best Practices](#best-practices)

---

## Overview

Organizations in nvolt provide multi-tenancy, allowing you to:

- **Separate secrets** for different companies or teams
- **Manage access control** with admin and developer roles
- **Switch contexts** easily between organizations
- **Collaborate** with different groups of people

### Organization Structure

```
Organization (e.g., "Acme Corp")
  ├── Users (admins and developers)
  ├── Machines (laptops, servers, CI/CD runners)
  └── Projects
      └── Environments
          └── Secrets
```

### Roles

| Role | Permissions |
|------|-------------|
| **admin** | Full access: manage users, machines, and all secrets |
| **dev** | Limited access: manage secrets based on granted permissions |

---

## nvolt org

View the currently active organization or interactively switch between organizations.

### Usage

```bash
nvolt org
```

### Behavior

**If you have only one organization:**

```bash
$ nvolt org

→ Active Organization: My Personal Org (admin)
```

**If you have multiple organizations:**

Shows current organization and interactive selector:

```bash
$ nvolt org

Organization Management

→ Current active organization: Acme Corp

→ You belong to 3 organization(s)

# Interactive selector appears:
? Select organization:
  ▸ Acme Corp (admin)
    Personal Projects (dev)
    Client XYZ (dev)
```

Use arrow keys to select, press Enter to confirm.

### Output After Selection

```bash
✓ Active organization changed to: Personal Projects
```

**Note:** The selected organization is saved to `~/.nvolt/config.json` and will be used for all future commands until you switch again.

### When to Use

- After joining a new organization
- When working on different projects for different companies
- To verify which organization you're currently using
- Before running sensitive commands (e.g., production pushes)

---

## nvolt org set

Explicitly set the active organization using an interactive selector.

### Usage

```bash
nvolt org set
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--org` | `-o` | Organization ID to set (skips interactive prompt) |

### Examples

#### Interactive Selection

```bash
$ nvolt org set

? Select organization:
  ▸ Acme Corp (admin)
    Personal Projects (dev)
    Client XYZ (dev)

✓ Active organization set to: Acme Corp
```

#### Direct Selection (Non-Interactive)

```bash
nvolt org set -o org-abc123def456
```

Output:
```
✓ Active organization set to: Acme Corp
```

### Configuration

The active organization is stored in `~/.nvolt/config.json`:

```json
{
  "active_org": "org-abc123def456",
  ...
}
```

### When to Use

- Switching context between client projects
- Before performing organization-specific operations
- In scripts where you need deterministic org selection
- After mistakenly running a command in the wrong org

---

## Multi-Organization Workflows

### Scenario 1: Freelancer with Multiple Clients

As a freelancer working with multiple clients:

```bash
# Morning: Work on Client A's project
nvolt org set -o org-client-a
nvolt pull -p web-app -e production
nvolt run -p web-app -e production -c "npm run deploy"

# Afternoon: Work on Client B's project
nvolt org set -o org-client-b
nvolt pull -p mobile-api -e staging
nvolt run -p mobile-api -e staging -c "pytest"
```

### Scenario 2: Developer with Personal and Work Orgs

```bash
# Work on personal side project
nvolt org set -o org-personal
nvolt push -f .env.dev -p my-side-project -e development

# Switch to work organization
nvolt org set -o org-mycompany
nvolt pull -p work-app -e production
```

### Scenario 3: Agency Managing Multiple Client Accounts

**Setup once:**

```bash
# Configure per-client organizations
nvolt org set -o org-client-nike
nvolt push -f nike.env -p website -e production

nvolt org set -o org-client-adidas
nvolt push -f adidas.env -p website -e production
```

**Daily usage:**

```bash
# Deploy for Nike
nvolt org set -o org-client-nike
nvolt run -p website -e production -c "./deploy.sh"

# Deploy for Adidas
nvolt org set -o org-client-adidas
nvolt run -p website -e production -c "./deploy.sh"
```

### Scenario 4: Auto-Detect Organization

For projects consistently tied to one organization, use environment defaults:

```bash
# Set default org for this project
cd ~/projects/acme-app
nvolt set -o org-acme

# From now on, commands automatically use org-acme
nvolt pull -p acme-app -e production
nvolt push -f .env -p acme-app -e staging
```

---

## Organization Resolution Logic

nvolt uses smart logic to determine the active organization:

### 1. Single Organization

If you belong to only **one** organization:
- It's automatically selected
- No prompt shown
- Not saved to config (ephemeral)

```bash
$ nvolt pull -p my-app -e dev

→ Using organization: Acme Corp (auto-detected)
✓ Successfully pulled 5 variable(s)!
```

### 2. Multiple Organizations + Config Set

If you have **multiple** organizations and `active_org` is set in config:
- Uses the configured organization
- No prompt shown

```bash
$ cat ~/.nvolt/config.json | grep active_org
"active_org": "org-abc123"

$ nvolt pull -p my-app -e dev

→ Using organization: Acme Corp (org-abc123)
✓ Successfully pulled 5 variable(s)!
```

### 3. Multiple Organizations + No Config

If you have **multiple** organizations and no `active_org` in config:
- Shows interactive selector
- Asks if you want to save as default

```bash
$ nvolt pull -p my-app -e dev

? Select organization:
  ▸ Acme Corp (admin)
    Personal Projects (dev)

? Set as default organization? (y/N): y

✓ Set 'Acme Corp' as default organization
✓ Successfully pulled 5 variable(s)!
```

### Overriding with Flags

You can always override the active organization using the `-o` flag:

```bash
# Use specific org, ignoring config
nvolt pull -p my-app -e dev -o org-xyz789
```

---

## Listing Organizations

Currently, there's no dedicated `nvolt org list` command. To see your organizations, run:

```bash
nvolt org
```

This shows all organizations you belong to with their roles.

---

## Creating Organizations

Organizations are automatically created during first login. Each user gets a default personal organization.

**Future feature:** `nvolt org create <name>` will allow creating additional organizations.

---

## Joining Organizations

To join an existing organization:

1. An **admin** of the target organization must invite you:
   ```bash
   nvolt user add your-email@example.com
   ```

2. You'll receive an email notification (or immediate access if already logged in)

3. Run `nvolt org` to see the new organization in your list

4. Switch to it with `nvolt org set`

---

## Best Practices

### 1. Set Organization Context Before Sensitive Operations

Always verify your active organization before pushing to production:

```bash
# Check current org
nvolt org

# If wrong, switch
nvolt org set -o org-correct

# Then push
nvolt push -f .env.prod -p my-app -e production
```

### 2. Use Descriptive Organization Names

When creating organizations, use clear names:

✅ **Good:**
- "Acme Corporation"
- "Personal Projects"
- "Client: Nike"
- "Startup XYZ - Dev Team"

❌ **Bad:**
- "Org1", "Org2"
- "Test"
- "My Org"

### 3. Document Organization IDs

For team onboarding, maintain a document with organization IDs:

```markdown
## nvolt Organizations

- **Production:** org-abc123def456 (Acme Corp - Production)
- **Staging:** org-def456ghi789 (Acme Corp - Staging)
- **Development:** org-ghi789jkl012 (Acme Corp - Development)
```

Share this with new team members for easy setup.

### 4. Use Per-Project Defaults

Set organization defaults at the project level:

```bash
cd ~/projects/acme-app
nvolt set -o org-acme

cd ~/projects/personal-blog
nvolt set -o org-personal
```

### 5. Audit Organization Access

Regularly review who has access to each organization:

```bash
# Switch to org
nvolt org set -o org-production

# List users
nvolt user list

# List machines
nvolt machine list
```

### 6. Separate Personal and Work

Maintain distinct organizations for personal vs. professional projects:

```
Personal Organization
  └── Side projects, experiments

Work Organization
  └── Company projects, client work
```

Benefits:
- Clear separation of secrets
- Different backup/recovery policies
- Easier offboarding when leaving a company

---

## Configuration File

Organization settings are stored in `~/.nvolt/config.json`:

```json
{
  "jwt_token": "eyJhbGciOiJIUzI1NiIs...",
  "machine_id": "m-abc123def456",
  "active_org": "org-xyz789",
  "default_environment": "development",
  "server_url": "https://api.nvolt.io"
}
```

Your private key is stored separately in `~/.nvolt/private_key.pem`.

### Fields

- **jwt_token:** Authentication token
- **machine_id:** Unique machine identifier
- **active_org:** Currently selected organization ID
- **default_environment:** Default environment for commands (optional)
- **server_url:** nvolt server URL (for self-hosting)

---

## Troubleshooting

### "No active organization set"

**Error:**
```
no active organization set. Use -o flag or run 'nvolt org set' first
```

**Solution:**
```bash
nvolt org set
```

Select an organization from the interactive prompt.

### "Organization not found"

**Error:**
```
organization 'org-abc123' not found
```

**Causes:**
- Typo in organization ID
- You don't have access to this organization
- Organization was deleted

**Solution:**
```bash
# List your organizations
nvolt org

# Set to a valid organization
nvolt org set
```

### "Permission denied"

**Error:**
```
permission denied: admin role required
```

**Cause:** You're a `dev` role in this organization and attempted an admin operation (e.g., adding users).

**Solution:** Contact an organization admin to perform the operation or upgrade your role.

---

## Related Commands

- **[nvolt user add](users.md#nvolt-user-add)** — Add users to organization (admin only)
- **[nvolt machine list](machines.md#nvolt-machine-list)** — List machines in organization
- **[nvolt set](secrets.md#nvolt-set)** — Set default organization and environment

---

## Next Steps

- Learn about [User Management](users.md)
- Understand [Role-Based Access Control](../security-model.md#permissions-model)
- Set up [Team Workflows](../ci-cd-integration.md)

---

[← Back to Documentation Home](../README.md)

