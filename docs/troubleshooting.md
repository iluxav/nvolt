# Troubleshooting

Common issues and solutions when using nvolt.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Authentication Issues](#authentication-issues)
- [Secret Management Issues](#secret-management-issues)
- [Permission Issues](#permission-issues)
- [Machine Issues](#machine-issues)
- [Organization Issues](#organization-issues)
- [Network Issues](#network-issues)
- [Getting Help](#getting-help)

---

## Installation Issues

### nvolt command not found

**Symptoms:**
```bash
$ nvolt --version
bash: nvolt: command not found
```

**Causes & Solutions:**

1. **nvolt not in PATH:**
   ```bash
   # Check if nvolt is installed
   ls -la /usr/local/bin/nvolt
   
   # If it exists, add to PATH
   export PATH="/usr/local/bin:$PATH"
   
   # Make permanent (bash)
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   
   # Make permanent (zsh)
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```

2. **Installation failed:**
   ```bash
   # Try manual installation
   curl -sL https://install.nvolt.io/cli -o install.sh
   bash install.sh
   
   # Or build from source
   git clone https://github.com/yourusername/nvolt-cli.git
   cd nvolt-cli
   go build -o nvolt cli/main.go
   sudo mv nvolt /usr/local/bin/
   ```

### Permission denied during installation

**Symptoms:**
```bash
mv: cannot move 'nvolt' to '/usr/local/bin/nvolt': Permission denied
```

**Solution:**
```bash
# Install to user directory instead
mkdir -p ~/.local/bin
mv nvolt ~/.local/bin/

# Add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Or use sudo:
```bash
sudo mv nvolt /usr/local/bin/
```

---

## Authentication Issues

### Browser doesn't open during login

**Symptoms:**
```bash
$ nvolt login
Logging in...
⚠  Failed to open browser automatically
→ Please manually open: https://api.nvolt.io/login?machine_id=m-abc123
```

**Solution:**
Copy the URL and paste it into your browser manually.

**Prevent in future:**
- Ensure a default browser is set
- Check firewall isn't blocking browser launch

### Authentication timeout

**Symptoms:**
```bash
$ nvolt login
Logging in...
Waiting for authentication...
..........................................
authentication timeout
```

**Causes & Solutions:**

1. **Didn't complete OAuth flow:**
   - Open the URL in browser
   - Complete authentication within 2 minutes

2. **Network issues:**
   - Check internet connection
   - Verify access to `api.nvolt.io`

3. **Browser blocked cookies:**
   - Enable cookies for `nvolt.io`
   - Try incognito/private mode

### Silent login fails with "private key not found"

**Symptoms:**
```bash
$ nvolt login --silent --machine ci-runner --org org-xyz
failed to read private key from ~/.nvolt/private_key.pem: no such file or directory
```

**Solution:**
```bash
# Ensure private key exists
ls -la ~/.nvolt/private_key.pem

# If missing, copy from secure storage
mkdir -p ~/.nvolt
cp /path/to/ci-runner_key.pem ~/.nvolt/private_key.pem
chmod 600 ~/.nvolt/private_key.pem
```

### "machine not found in organization"

**Symptoms:**
```bash
authentication failed: machine 'ci-runner' not found in organization
```

**Causes & Solutions:**

1. **Machine not created:**
   ```bash
   # On an authenticated machine:
   nvolt machine add ci-runner
   ```

2. **Wrong machine name:**
   ```bash
   # List machines to find correct name
   nvolt machine list
   
   # Use exact machine name in --machine flag
   nvolt login --silent --machine <correct-name> --org org-xyz
   ```

3. **Wrong organization:**
   ```bash
   # Verify org ID
   nvolt machine list -o org-xyz
   ```

---

## Secret Management Issues

### "No variables found for this scope"

**Symptoms:**
```bash
$ nvolt pull -p my-app -e production

⚠ No variables found for this scope.
If you just logged in, please run the command `nvolt sync` from any authorized machine to sync your machine.
```

**Causes & Solutions:**

1. **New machine not synced:**
   ```bash
   # On ANY authorized machine (not the new one):
   nvolt sync
   
   # Then try pull again on new machine
   nvolt pull -p my-app -e production
   ```

2. **No secrets pushed yet:**
   ```bash
   # Push some secrets first
   nvolt push -k TEST=value -p my-app -e production
   ```

3. **Wrong project/environment:**
   ```bash
   # List available secrets (from another machine)
   nvolt pull  # Shows what's available
   
   # Use correct project/environment
   nvolt pull -p correct-project -e correct-environment
   ```

### "No wrapped key for this machine"

**Symptoms:**
```bash
⚠ No secrets found for this machine

This machine hasn't been synced yet. To enable access:
  1. Run 'nvolt sync' from ANY authorized machine, OR
  2. Run 'nvolt push' from any machine to sync all machines
```

**Solution:**
```bash
# From any authorized machine (e.g., your laptop):
nvolt sync

# Or push new secrets (automatically syncs):
nvolt push -k DUMMY=value -p my-app -e dev
```

### "WrappedKey is empty"

**Symptoms:**
Error during pull mentioning "WrappedKey is empty"

**Cause:** Machine was added but secrets were never synced.

**Solution:**
```bash
# From an authorized machine:
nvolt sync

# Then try pull again
nvolt pull -p my-app -e production
```

### Push fails with "invalid key-value format"

**Symptoms:**
```bash
$ nvolt push -k MYVAR=value=with=equals

invalid key-value format 'MYVAR=value=with=equals': multiple = signs
```

**Cause:** Secret value contains `=` signs.

**Solution:**

Use a file instead:
```bash
# Create .env file
echo 'MYVAR=value=with=equals' > .env.temp

# Push from file
nvolt push -f .env.temp -p my-app -e dev

# Clean up
rm .env.temp
```

Or quote the entire value in shell:
```bash
nvolt push -k "MYVAR=value=with=equals" -p my-app -e dev
```

---

## Permission Issues

### "Permission Denied" when pushing/pulling

**Symptoms:**
```bash
$ nvolt push -f .env -p my-app -e production

⚠ Permission Denied

You don't have write permission for this environment.
  Project: my-app
  Environment: production

Please contact your organization admin to grant you access.
```

**Causes & Solutions:**

1. **Insufficient permissions:**
   ```bash
   # Contact admin to grant access
   # Admin runs:
   nvolt user mod your-email@example.com
   # Select project, environment, and grant permissions
   ```

2. **Wrong organization:**
   ```bash
   # Verify active org
   nvolt org
   
   # Switch if needed
   nvolt org set -o correct-org-id
   ```

### "admin role required"

**Symptoms:**
```bash
$ nvolt user add newuser@example.com

permission denied: admin role required
```

**Cause:** You have `dev` role but tried an admin operation.

**Solution:**
Contact an organization admin to:
- Perform the operation for you, OR
- Upgrade your role to `admin`

---

## Machine Issues

### Machine add fails with "machine already exists"

**Symptoms:**
```bash
$ nvolt machine add github-actions

failed to save public key: machine 'github-actions' already exists
```

**Solutions:**

1. **Use a different name:**
   ```bash
   nvolt machine add github-actions-v2
   ```

2. **Remove old machine first (future feature):**
   ```bash
   # Future: nvolt machine rm github-actions
   ```

### Can't list machines

**Symptoms:**
```bash
$ nvolt machine list

✗ Failed to fetch machines: permission denied
```

**Cause:** Non-admin users may have limited access.

**Solution:**
- Only admins can list all machines
- Contact admin for machine information

---

## Organization Issues

### "No active organization set"

**Symptoms:**
```bash
$ nvolt pull -p my-app -e dev

no active organization set. Use -o flag or run 'nvolt org set' first
```

**Solution:**
```bash
# Set active organization
nvolt org set

# Or specify org in command
nvolt pull -p my-app -e dev -o org-xyz789
```

### Can't switch organizations

**Symptoms:**
```bash
$ nvolt org set

You do not belong to any organization
```

**Cause:** You're not a member of any organization yet.

**Solution:**
- Ask an admin to add you: `nvolt user add your-email@example.com`
- Or create a new organization by logging in (auto-created on first login)

### Wrong organization selected

**Symptoms:**
Working in wrong org context, secrets not found.

**Solution:**
```bash
# Check current org
nvolt org

# Switch to correct org
nvolt org set -o correct-org-id

# Or use -o flag in commands
nvolt pull -p my-app -e dev -o correct-org-id
```

---

## Network Issues

### "connection refused"

**Symptoms:**
```bash
$ nvolt login

failed to poll for token: connection refused
```

**Causes & Solutions:**

1. **Server down:**
   - Check [status.nvolt.io](https://status.nvolt.io) (if exists)
   - Wait and retry

2. **Network issues:**
   ```bash
   # Test connectivity
   curl -I https://api.nvolt.io
   
   # Check DNS resolution
   nslookup api.nvolt.io
   
   # Try with different network (mobile hotspot)
   ```

3. **Firewall blocking:**
   - Check corporate firewall rules
   - Verify proxy settings
   - Whitelist `*.nvolt.io`

### "TLS handshake timeout"

**Symptoms:**
```bash
failed to connect: TLS handshake timeout
```

**Solutions:**

1. **Proxy issues:**
   ```bash
   # Check proxy settings
   echo $HTTP_PROXY
   echo $HTTPS_PROXY
   
   # Unset if incorrect
   unset HTTP_PROXY HTTPS_PROXY
   ```

2. **Clock skew:**
   ```bash
   # Sync system clock
   sudo ntpdate -s time.nist.gov  # Linux
   
   # Or enable automatic time sync
   # System Preferences → Date & Time → Set time automatically (macOS)
   ```

### Self-hosted server connection issues

**Symptoms:**
```bash
$ nvolt pull

connection to https://nvolt.mycompany.com refused
```

**Solutions:**

1. **Set correct server URL:**
   ```bash
   nvolt set -s https://nvolt.mycompany.com
   ```

2. **Verify server is running:**
   ```bash
   curl https://nvolt.mycompany.com/health
   ```

3. **Check TLS certificate:**
   ```bash
   # For self-signed certs (development only):
   export NVOLT_SKIP_TLS_VERIFY=true
   nvolt pull
   ```

---

## CI/CD Issues

### Private key not working in CI

**Symptoms:**
```bash
# In CI logs
authentication failed: invalid signature
```

**Causes & Solutions:**

1. **Newlines not preserved:**
   ```yaml
   # GitHub Actions - Use quotes and literal style
   - run: |
       echo "${{ secrets.NVOLT_PRIVATE_KEY }}" > ~/.nvolt/private_key.pem
   ```

2. **Secret not masked properly:**
   ```bash
   # Verify file was created (should see file size, not contents)
   ls -lh ~/.nvolt/private_key.pem
   ```

3. **Wrong permissions:**
   ```bash
   chmod 600 ~/.nvolt/private_key.pem
   ```

### CI runs fail intermittently

**Symptoms:**
Sometimes works, sometimes fails with "authentication timeout"

**Cause:** Network instability or server rate limiting.

**Solution:**
Add retry logic:
```yaml
- name: Authenticate (with retry)
  run: |
    for i in 1 2 3; do
      nvolt login --silent --machine ci-runner --org ${{ secrets.NVOLT_ORG_ID }} && break
      echo "Retry $i failed, waiting..."
      sleep 5
    done
```

---

## Configuration Issues

### Config file corrupted

**Symptoms:**
```bash
$ nvolt pull

failed to load config: invalid JSON
```

**Solution:**
```bash
# Backup corrupted files
mv ~/.nvolt/config.json ~/.nvolt/config.json.bak
mv ~/.nvolt/private_key.pem ~/.nvolt/private_key.pem.bak

# Re-login
nvolt login

# Restore active org if needed
nvolt org set -o your-org-id
```

### Lost private key

**Symptoms:**
```bash
# After reinstalling OS or losing ~/.nvolt/
$ nvolt pull

authentication failed: no private key found
```

**Solutions:**

1. **Restore from backup:**
   ```bash
   # If you backed up ~/.nvolt/
   cp ~/backup/.nvolt/config.json ~/.nvolt/
   cp ~/backup/.nvolt/private_key.pem ~/.nvolt/
   chmod 600 ~/.nvolt/private_key.pem
   ```

2. **Generate new machine key:**
   ```bash
   # From another authenticated machine:
   nvolt machine add my-laptop-new
   
   # On this machine, save the new key:
   mkdir -p ~/.nvolt
   mv my-laptop-new_key.pem ~/.nvolt/private_key.pem
   chmod 600 ~/.nvolt/private_key.pem
   
   # Silent login
   nvolt login --silent --machine my-laptop-new --org org-xyz
   
   # Sync to get access to secrets
   nvolt sync
   ```

---

## Performance Issues

### Slow push/pull

**Symptoms:**
Operations take a long time (>30 seconds) for small secrets.

**Causes & Solutions:**

1. **Many machines in org:**
   - Pushing wraps keys for all machines (slow with 100+ machines)
   - Contact support about org-level encryption optimization

2. **Large secrets:**
   - Avoid storing large files as secrets
   - Use references instead (S3 URLs, etc.)

3. **Network latency:**
   ```bash
   # Check latency
   ping api.nvolt.io
   
   # Use closer region (if self-hosting)
   nvolt set -s https://nvolt-us-west.mycompany.com
   ```

### CLI hangs

**Symptoms:**
CLI appears frozen, no output.

**Solution:**
```bash
# Kill the process
Ctrl+C

# Try with verbose mode (future feature)
# nvolt pull --verbose

# Check for server issues
curl https://api.nvolt.io/health
```

---

## Environment Variable Issues

### Secrets not injected in `nvolt run`

**Symptoms:**
```bash
$ nvolt run -p my-app -e dev -c "env | grep DATABASE_URL"
# No output
```

**Causes & Solutions:**

1. **Secret doesn't exist:**
   ```bash
   # Verify secret was pushed
   nvolt pull -p my-app -e dev | grep DATABASE_URL
   ```

2. **Wrong environment:**
   ```bash
   # Check you're using correct environment
   nvolt pull -p my-app -e dev
   ```

3. **Command doesn't inherit env vars:**
   ```bash
   # Some shells/commands don't inherit env
   # Use shell explicitly:
   nvolt run -p my-app -e dev -c "bash -c 'env | grep DATABASE_URL'"
   ```

### Environment variable precedence issues

**Symptoms:**
Wrong value for environment variable in `nvolt run`.

**Cause:** Local env vars override nvolt secrets.

**Solution:**
```bash
# Unset local variable
unset DATABASE_URL

# Then run
nvolt run -p my-app -e dev -c "npm start"

# Or use fresh shell
nvolt run -p my-app -e dev -c "env -i bash -c 'npm start'"
```

---

## Getting Help

If you can't resolve your issue:

### 1. Check Documentation

- [Getting Started](getting-started.md)
- [Commands Reference](commands/)
- [Security Model](security-model.md)

### 2. Search GitHub Issues

[github.com/yourusername/nvolt-cli/issues](https://github.com/yourusername/nvolt-cli/issues)

### 3. Join Community

- [GitHub Discussions](https://github.com/yourusername/nvolt-cli/discussions)
- [Discord Server](https://discord.gg/nvolt) (if exists)

### 4. Contact Support

- **Email:** support@nvolt.io
- **Response time:** 24-48 hours for free users, <4 hours for paid plans

### 5. File a Bug Report

File issues at: [github.com/yourusername/nvolt-cli/issues](https://github.com/yourusername/nvolt-cli/issues)

When reporting issues, include:

```bash
# System info
uname -a
nvolt --version

# Error message (full output)
nvolt <command> 2>&1

# Steps to reproduce
# 1. Run X
# 2. Do Y
# 3. See error Z
```

**Template:**

```markdown
## Description
Brief description of the issue

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: macOS 13.2
- nvolt version: v1.2.3
- Shell: zsh 5.8

## Error Messages
```
paste error output here
```

## Additional Context
Any other relevant information
```

---

## Quick Diagnostics

Run these commands to gather diagnostic information:

```bash
# Version info
nvolt --version

# Check config
cat ~/.nvolt/config.json | jq '.machine_id, .active_org, .server_url'

# Check private key exists
ls -la ~/.nvolt/private_key.pem

# Test network
curl -I https://api.nvolt.io

# Check permissions
ls -la ~/.nvolt/

# Test authentication
nvolt org

# List available resources
nvolt machine list
nvolt user list
```

---

[← Back to Documentation Home](README.md)

