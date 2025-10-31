# Authentication Commands

nvolt supports two authentication modes: **interactive OAuth** for developers and **silent authentication** for CI/CD machines.

## Table of Contents

- [nvolt login](#nvolt-login)
- [Silent Login for CI/CD](#silent-login-for-cicd)
- [Understanding Machine Identity](#understanding-machine-identity)
- [Security Considerations](#security-considerations)

---

## nvolt login

Interactive authentication using OAuth via browser.

### Usage

```bash
nvolt login
```

### What Happens

When you run `nvolt login`:

1. **Key Generation:** A unique RSA key pair (2048-bit) is generated for your machine
2. **Browser Opens:** Your default browser opens to the OAuth login page
3. **Authentication:** You authenticate using Google, GitHub, or another provider
4. **Token Exchange:** The server issues a JWT token for your session
5. **Key Registration:** Your public key is sent to the server and registered
6. **Organization Setup:** You're automatically added to an organization

### Output

```bash
$ nvolt login

Logging in...
Waiting for authentication...
..................

✓ Successfully authenticated!
```

### Configuration File

After successful login, your configuration is stored in two files:

**`~/.nvolt/config.json`:**

```json
{
  "jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "machine_id": "m-abc123def456",
  "active_org": "org-xyz789",
  "server_url": "https://api.nvolt.io"
}
```

**`~/.nvolt/private_key.pem`:**

```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----
```

> ⚠️ **Security Warning:** These files contain sensitive cryptographic keys. Never commit `~/.nvolt/` to version control or share these files with others.

### Troubleshooting

**Browser doesn't open automatically:**

If the browser fails to open, you'll see a URL to manually open:

```bash
⚠  Failed to open browser automatically
→ Please manually open: https://api.nvolt.io/login?machine_id=m-abc123
```

Copy the URL and paste it into your browser.

**Authentication timeout:**

If authentication takes longer than 2 minutes, the command will timeout. Try running `nvolt login` again.

---

## Silent Login for CI/CD

Silent authentication allows machines (like CI/CD runners) to authenticate without a browser using a pre-provisioned private key.

### Usage

```bash
nvolt login --silent --machine <machine-name> --org <org-id>
```

### Flags

| Flag        | Short | Required | Description                         |
| ----------- | ----- | -------- | ----------------------------------- |
| `--silent`  | `-s`  | Yes      | Enable silent authentication mode   |
| `--machine` | `-m`  | Yes      | Machine name (identifier)           |
| `--org`     | `-o`  | Yes      | Organization ID to authenticate for |

### Prerequisites

Before using silent login:

1. **Generate a machine key** from an authenticated machine:

   ```bash
   nvolt machine add ci-runner-prod
   ```

2. **Securely transfer the private key** to the CI/CD machine:
   - The private key will be saved as `ci-runner-prod_key.pem`
   - Transfer it to the destination machine at `~/.nvolt/private_key.pem`
   - Set proper permissions: `chmod 600 ~/.nvolt/private_key.pem`

### How It Works

Silent authentication uses a **challenge-response protocol**:

```
1. CLI requests challenge from server
   → Server: "Prove you have the private key"

2. Server generates random challenge and sends it encrypted with machine's public key
   → Challenge: <encrypted-random-data>

3. CLI decrypts challenge using private key
   → Decrypted: <random-data>

4. CLI signs the challenge with private key
   → Signature: <cryptographic-signature>

5. Server verifies signature using public key
   → If valid: Issues JWT token
```

**Security benefit:** The private key never leaves the machine, and the server never sees it.

### Example: GitHub Actions

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

      - name: Authenticate
        run: nvolt login --silent --machine github-actions --org ${{ secrets.NVOLT_ORG_ID }}

      - name: Deploy
        run: nvolt run -p my-app -e production -c "./deploy.sh"
```

### Example: GitLab CI

```yaml
deploy:
  stage: deploy
  script:
    - curl -sL https://install.nvolt.io/cli | bash
    - mkdir -p ~/.nvolt
    - echo "$NVOLT_PRIVATE_KEY" > ~/.nvolt/private_key.pem
    - chmod 600 ~/.nvolt/private_key.pem
    - nvolt login --silent --machine gitlab-ci --org $NVOLT_ORG_ID
    - nvolt run -p my-app -e production -c "./deploy.sh"
  only:
    - main
```

### Output

```bash
$ nvolt login --silent --machine ci-runner-prod --org org-xyz789

🔐 Silent login for machine: ci-runner-prod (org: org-xyz789)

✓ Successfully authenticated!
```

### Error Handling

**Private key not found:**

```bash
failed to read private key from ~/.nvolt/private_key.pem: no such file or directory
Please ensure the private key exists
```

**Solution:** Ensure the private key file exists at `~/.nvolt/private_key.pem`

**Invalid machine name:**

```bash
authentication failed: machine 'ci-runner-prod' not found in organization
```

**Solution:** The machine must first be created using `nvolt machine add <machine-name>` from an authenticated machine.

**Wrong organization:**

```bash
authentication failed: machine does not belong to organization 'org-abc123'
```

**Solution:** Verify the organization ID matches the one where the machine was created.

---

## Understanding Machine Identity

### What is a Machine?

In nvolt, a **machine** represents a unique device or compute environment with its own cryptographic identity:

- Developer laptop
- CI/CD runner (GitHub Actions, GitLab CI, etc.)
- Production server
- Team member's workstation

### Machine ID

Each machine has a unique identifier (e.g., `m-abc123def456`) that's automatically generated during first login. This ID is used to:

- Track which machine performed operations (audit trail)
- Encrypt secrets specifically for this machine
- Manage access control

### Per-Machine Keys

Every machine has its own RSA key pair:

- **Private Key:** Stored locally, never leaves the machine
- **Public Key:** Registered on the server

This ensures:

- Secrets can be encrypted for specific machines
- Compromising one machine doesn't compromise others
- Fine-grained access control per device

### Viewing Your Machine ID

```bash
cat ~/.nvolt/config.json | grep machine_id
```

Output:

```
"machine_id": "m-abc123def456"
```

---

## Security Considerations

### Best Practices

✅ **Never share your private key** with anyone  
✅ **Use unique keys per machine** — don't copy the same key to multiple devices  
✅ **Rotate keys periodically** — generate new machine keys every 90 days  
✅ **Use silent login for CI/CD** — avoid interactive login in automated environments  
✅ **Set restrictive permissions** — `chmod 600 ~/.nvolt/config.json`  
✅ **Use separate machines for production** — isolate production access

### What to Protect

Your `~/.nvolt/` directory contains:

- **Private RSA key** (`private_key.pem`) — Can decrypt all secrets you have access to
- **JWT token** (`config.json`) — Grants API access to the nvolt server
- **Machine ID** (`config.json`) — Identifies your device

**Security tips:**

1. Add `~/.nvolt/` to `.gitignore` globally:

   ```bash
   echo ".nvolt/" >> ~/.gitignore_global
   git config --global core.excludesfile ~/.gitignore_global
   ```

2. Set restrictive permissions:

   ```bash
   chmod 600 ~/.nvolt/private_key.pem
   chmod 600 ~/.nvolt/config.json
   ```

3. Backup your config securely (use password manager or encrypted storage)

4. If compromised, immediately:
   - Delete the machine: `nvolt machine rm <machine-id>` (from another device)
   - Generate a new key: `nvolt machine add my-laptop-new`
   - Re-sync secrets: `nvolt sync`

### Threat Scenarios

| Threat               | nvolt Protection                                  | Your Responsibility                |
| -------------------- | ------------------------------------------------- | ---------------------------------- |
| Server breach        | ✅ Server only has encrypted data and public keys | Monitor security advisories        |
| Network interception | ✅ TLS encryption + encrypted payloads            | Use trusted networks               |
| Stolen laptop        | ⚠️ Private key on disk                            | Encrypt disk, set strong password  |
| Malicious teammate   | ✅ RBAC and per-environment permissions           | Review user permissions regularly  |
| CI/CD compromise     | ⚠️ Private key in CI environment                  | Limit CI secret scope, rotate keys |

---

## Related Commands

- **[nvolt machine add](machines.md#nvolt-machine-add)** — Create a new machine identity
- **[nvolt machine list](machines.md#nvolt-machine-list)** — List all machines in your organization
- **[nvolt org set](organization.md#nvolt-org-set)** — Switch active organization

---

## Next Steps

- Learn about [Secrets Management](secrets.md) — Push, pull, and run commands
- Set up [CI/CD Integration](../ci-cd-integration.md) — Automate secret management
- Understand the [Security Model](../security-model.md) — Deep dive into cryptography

---

[← Back to Documentation Home](../README.md)
