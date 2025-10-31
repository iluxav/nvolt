# User Management Commands

Manage users and permissions in your organization. **Admin role required.**

## Table of Contents

- [Overview](#overview)
- [nvolt user add](#nvolt-user-add)
- [nvolt user list](#nvolt-user-list)
- [nvolt user mod](#nvolt-user-mod)
- [nvolt user rm](#nvolt-user-rm)
- [Permission Model](#permission-model)
- [Common Workflows](#common-workflows)

---

## Overview

User management in nvolt allows organization admins to:

- **Invite users** to the organization
- **Grant permissions** at project and environment levels
- **Modify access** as team structure changes
- **Remove users** when they leave the organization

### Prerequisites

- You must have **admin** role in the organization
- Target user must have a registered nvolt account (created via `nvolt login`)

### Permission Hierarchy

```
Organization Level
  └── Role: admin or dev

    Project Level
      └── Permissions: {read, write, delete}

        Environment Level
          └── Permissions: {read, write, delete}
```

---

## nvolt user add

Add a user to an organization with optional project and environment permissions.

### Usage

```bash
# Interactive mode
nvolt user add <email>

# With project and environment
nvolt user add <email> -p <project> -e <environment>

# With explicit permissions
nvolt user add <email> -p <project> -e <environment> -a read=true,write=true,delete=false
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `email` | Yes | Email address of the user to add |

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--org` | `-o` | Organization ID | Active org from config |
| `--project` | `-p` | Project name | Interactive prompt |
| `--environment` | `-e` | Environment name | Interactive prompt if project specified |
| `--environment-permissions` | `-a` | Environment permissions string | Interactive selector |

### Examples

#### Add User with No Initial Permissions

```bash
nvolt user add alice@example.com
```

**Output:**
```
Adding User to Organization
→ Email: alice@example.com
→ Organization ID: org-xyz789
→ Sending request to server...
✓ User alice@example.com added successfully
  User ID: user-abc123
  Name: Alice Smith
```

User is added to the organization but has no access to any projects/environments yet.

#### Add User with Interactive Permission Setup

```bash
nvolt user add bob@example.com -p my-app -e production
```

**Interactive prompts:**
```
? Environment Permissions (use space to select, enter to confirm):
  ▸ ◉ read
    ◯ write
    ◯ delete

→ Permissions: read=true, write=false, delete=false
→ Sending request to server...
✓ User bob@example.com added successfully
```

#### Add User with Explicit Permissions

```bash
nvolt user add charlie@example.com \
  -p my-app \
  -e staging \
  -a read=true,write=true,delete=false
```

**Output:**
```
Adding User to Organization
→ Email: charlie@example.com
→ Organization ID: org-xyz789
→ Environment: staging
  Permissions: read=true, write=true, delete=false
→ Sending request to server...
✓ User charlie@example.com added successfully
```

#### Add User to Multiple Environments

```bash
# Add with development access
nvolt user add dana@example.com -p my-app -e development \
  -a read=true,write=true,delete=true

# Add production access (read-only)
nvolt user mod dana@example.com -p my-app -e production \
  -a read=true,write=false,delete=false
```

### Permission String Format

Format: `read=<bool>,write=<bool>,delete=<bool>`

Examples:
- `read=true,write=false,delete=false` — Read-only access
- `read=true,write=true,delete=false` — Read and write (typical developer)
- `read=true,write=true,delete=true` — Full access

### Automatic Key Re-Wrapping

When adding a user with environment permissions, nvolt automatically re-wraps encryption keys:

```
🔄 Re-wrapping encryption keys...
→ Found 2 project/environment combination(s) to sync

[1/2] Syncing my-app/development... ✓
[2/2] Syncing my-app/staging... ✓

Key Re-wrapping Summary:
✓ Successfully synced: 2
→ New user can now access synced secrets
```

This ensures the new user's machines can decrypt secrets immediately.

---

## nvolt user list

List all users in an organization with their roles.

### Usage

```bash
nvolt user list
```

**Alias:** `nvolt user ls`

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--org` | `-o` | Organization ID | Active org from config |

### Example

```bash
$ nvolt user list

Organization Users
→ Organization: Acme Corp (org-xyz789)
→ Fetching users...

Found 5 user(s):

┌──────────────────┬─────────────────────┬────────┬──────────────────┐
│ Name             │ Email               │ Role   │ User ID          │
├──────────────────┼─────────────────────┼────────┼──────────────────┤
│ Alice Smith      │ alice@example.com   │ admin  │ user-abc123      │
│ Bob Johnson      │ bob@example.com     │ dev    │ user-def456      │
│ Charlie Davis    │ charlie@example.com │ dev    │ user-ghi789      │
│ Dana Lee         │ dana@example.com    │ dev    │ user-jkl012      │
│ Eve Martinez     │ eve@example.com     │ admin  │ user-mno345      │
└──────────────────┴─────────────────────┴────────┴──────────────────┘
```

### Output Details

Each row shows:
- **Name:** User's full name
- **Email:** User's email address
- **Role:** `admin` or `dev`
- **User ID:** Unique identifier

### Use Cases

- **Security audit:** Review all users with access
- **Onboarding verification:** Confirm new team members were added
- **Access management:** Identify users to modify or remove

---

## nvolt user mod

Modify a user's role and permissions interactively.

### Usage

```bash
nvolt user mod <email>
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `email` | Yes | Email address of the user to modify |

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--org` | `-o` | Organization ID | Active org from config |

### Interactive Flow

When you run `nvolt user mod`, you'll be guided through:

1. **Fetching current permissions**
2. **Select new role** (admin or dev)
3. **Select project** (or "All")
4. **Select environment** (or "All")
5. **Set permissions** (read, write, delete)

### Example: Modify Single Environment

```bash
$ nvolt user mod bob@example.com

Modifying User Permissions
→ Email: bob@example.com
→ Organization ID: org-xyz789
→ Fetching current permissions...
→ Current Role: dev

? Select user role:
  ▸ admin
    dev

? Select project:
  ▸ my-app
    another-app
    All (WARNING! Grant permissions for ALL projects and all project Environments in this org)

? Select environment:
  ▸ development
    staging
    production
    All (Grant permissions for ALL environments in this project)

? Environment Permissions for 'production' (use space to select, enter to confirm):
  ▸ ◉ read
    ◉ write
    ◯ delete

→ Sending request to server...
✓ User permissions updated successfully
  User: Bob Johnson (bob@example.com)
```

### Example: Grant Access to All Environments

```bash
$ nvolt user mod charlie@example.com

? Select project:
  ▸ my-app

? Select environment:
  ▸ All (Grant permissions for ALL environments in this project)

⚠ WARNING! This will grant environment permissions for ALL environments in project 'my-app'

? Environment Permissions (applies to ALL environments in this project):
  ▸ ◉ read
    ◯ write
    ◯ delete

→ Updating permissions for my-app/development
→ Updating permissions for my-app/staging
→ Updating permissions for my-app/production

✓ Successfully updated permissions for user charlie@example.com in project my-app
```

### Example: Grant Access to Everything

```bash
? Select project:
  ▸ All (WARNING! Grant permissions for ALL projects and all project Environments in this org)

⚠ WARNING! This will grant environment permissions for ALL projects and ALL environments in this org

? Environment Permissions (applies to ALL environments):
  ▸ ◉ read
    ◉ write
    ◉ delete

→ Updating permissions for my-app/development
→ Updating permissions for my-app/staging
...

✓ Successfully updated permissions for user dana@example.com across all projects/environments
```

### Current Permissions Pre-Selected

The interactive prompts pre-select the user's current permissions, making it easy to see and modify existing access.

---

## nvolt user rm

Remove a user from an organization.

### Usage

```bash
nvolt user rm <email>
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `email` | Yes | Email address of the user to remove |

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--org` | `-o` | Organization ID | Active org from config |

### Example

```bash
$ nvolt user rm eve@example.com

Removing User from Organization
→ Email: eve@example.com
→ Organization ID: org-xyz789
⚠ This will remove the user from the organization and delete all their permissions.
Are you sure? (yes/no): yes

→ Sending request to server...
✓ User eve@example.com successfully removed from organization
```

### Confirmation Required

To prevent accidental deletion, you must type `yes` or `y` to confirm.

### What Happens

When a user is removed:

1. **User-organization link deleted**
2. **All permissions revoked** for that user
3. **Machines remain** but lose access (can't decrypt secrets)
4. **Audit trail preserved** (past actions still logged)

### Re-Adding a Removed User

To restore access:

```bash
nvolt user add eve@example.com -p my-app -e production
```

The user must go through permission setup again (permissions are not restored automatically).

---

## Permission Model

### Three-Tier RBAC System

nvolt implements a hierarchical permission model:

#### 1. Organization-Level Role

| Role | Capabilities |
|------|--------------|
| **admin** | • Add/remove users<br>• Add machines<br>• Grant any permissions<br>• Full secret access |
| **dev** | • Access secrets per granted permissions<br>• Cannot manage users/machines |

#### 2. Project-Level Permissions

**Note:** Currently, nvolt doesn't enforce project-level permissions separately. All permissions are at the environment level.

#### 3. Environment-Level Permissions

| Permission | Grants |
|------------|--------|
| **read** | • Pull secrets with `nvolt pull`<br>• Run commands with `nvolt run` |
| **write** | • Push secrets with `nvolt push`<br>• Update existing secrets |
| **delete** | • Delete secrets (via full replacement push) |

### Permission Precedence

1. **Organization admin** — Full access to everything
2. **Environment permissions** — Specific to project + environment
3. **No permission** — Access denied

### Auto-Provisioning

First user to push to a new project/environment automatically gets full permissions for that scope.

Example:
```bash
# Alice pushes to a new environment
nvolt push -f .env -p new-app -e staging

# Alice automatically gets read=true, write=true, delete=true for new-app/staging
```

---

## Common Workflows

### Workflow 1: Onboard a New Developer

**Step 1: Add user with dev access**

```bash
nvolt user add newdev@example.com
```

**Step 2: Grant development environment access**

```bash
nvolt user mod newdev@example.com
```

Select:
- Role: `dev`
- Project: `my-app`
- Environment: `development`
- Permissions: `read=true, write=true, delete=false`

**Step 3: Verify access**

Ask the developer to run:
```bash
nvolt pull -p my-app -e development
```

### Workflow 2: Promote Developer to Admin

```bash
nvolt user mod alice@example.com
```

Select:
- Role: `admin`

As an admin, Alice now has full access to all projects and environments.

### Workflow 3: Grant Production Read-Only Access

For junior developers who need to debug production:

```bash
nvolt user mod junior@example.com
```

Select:
- Role: `dev`
- Project: `my-app`
- Environment: `production`
- Permissions: `read=true, write=false, delete=false`

### Workflow 4: Temporary Contractor Access

**Add contractor:**
```bash
nvolt user add contractor@example.com -p client-project -e staging \
  -a read=true,write=true,delete=false
```

**After contract ends:**
```bash
nvolt user rm contractor@example.com
```

### Workflow 5: Multi-Environment Setup for Team Member

```bash
# Development: Full access
nvolt user mod dev@example.com
# Select: my-app → development → read,write,delete

# Staging: Read/write
nvolt user mod dev@example.com
# Select: my-app → staging → read,write

# Production: Read-only
nvolt user mod dev@example.com
# Select: my-app → production → read
```

### Workflow 6: Revoke Production Access

```bash
nvolt user mod dev@example.com
```

Select:
- Project: `my-app`
- Environment: `production`
- Permissions: *deselect all*

Or simply remove the user and re-add with new permissions:
```bash
nvolt user rm dev@example.com
nvolt user add dev@example.com -p my-app -e development
```

---

## Best Practices

### 1. Principle of Least Privilege

Grant only the minimum permissions necessary:

✅ **Good:**
- Junior devs: Read-only production
- Senior devs: Read/write staging, read-only production
- DevOps: Full access to all environments

❌ **Bad:**
- Everyone: Full access to production

### 2. Separate Dev and Admin Roles

Don't grant `admin` role unless necessary:

- **Admins:** Team leads, DevOps, CTO
- **Devs:** All other engineers

### 3. Audit Permissions Quarterly

```bash
# Review users
nvolt user list

# Review specific user's access
nvolt user mod user@example.com  # Then cancel to see current permissions
```

### 4. Use Interactive Mode for Safety

Avoid hardcoding permissions in scripts. Interactive mode shows current state and prevents mistakes.

### 5. Document Permission Policy

Maintain a policy document:

```markdown
## nvolt Permission Policy

### Roles
- Admin: VP Engineering, DevOps team
- Dev: All other engineers

### Environment Access
- **Development:** All engineers (read/write/delete)
- **Staging:** All engineers (read/write)
- **Production:** 
  - Senior engineers (read/write)
  - Junior engineers (read-only)
```

### 6. Revoke Access Immediately on Offboarding

```bash
# Employee's last day
nvolt user rm former-employee@example.com
```

Consider automating this with HRIS integrations.

---

## Troubleshooting

### "Permission denied: admin role required"

**Error:**
```
permission denied: admin role required
```

**Cause:** You're trying to manage users but have `dev` role.

**Solution:** Contact an organization admin to perform the operation or upgrade your role.

### "User not found"

**Error:**
```
user 'user@example.com' not found
```

**Cause:** User hasn't registered with nvolt yet.

**Solution:** Ask the user to run `nvolt login` first, then add them to the organization.

### "User already exists in organization"

**Error:**
```
user 'user@example.com' already exists in organization
```

**Solution:** Use `nvolt user mod` to update permissions instead.

### Key Re-Wrapping Failures

**Error:**
```
[1/3] Syncing my-app/production... ✗ Failed: permission denied
```

**Cause:** You don't have access to some project/environments that the new user is being granted access to.

**Solution:** Ask another admin with broader access to run the user add command.

---

## Related Commands

- **[nvolt machine add](machines.md#nvolt-machine-add)** — Add machines for users
- **[nvolt org set](organization.md#nvolt-org-set)** — Switch organizations
- **[nvolt sync](secrets.md#nvolt-sync)** — Sync keys after adding users

---

## Next Steps

- Learn about [Permission Model](../security-model.md#permissions-model)
- Set up [Team Workflows](../ci-cd-integration.md)
- Review [Security Best Practices](../security-model.md#best-practices)

---

[← Back to Documentation Home](../README.md)

