# Security Model

An in-depth look at nvolt's cryptographic architecture, threat model, and security best practices.

## Table of Contents

- [Overview](#overview)
- [Cryptographic Architecture](#cryptographic-architecture)
- [Wrapped Key Encryption](#wrapped-key-encryption)
- [Challenge-Response Authentication](#challenge-response-authentication)
- [Threat Model](#threat-model)
- [Permissions Model](#permissions-model)
- [Best Practices](#best-practices)
- [Security Audit](#security-audit)

---

## Overview

nvolt is built on **Zero-Trust principles** with **end-to-end encryption**. The fundamental security guarantee is:

> **The server never has access to plaintext secrets or private keys.**

All cryptographic operations (encryption, decryption, signing) happen **client-side** on your machine. The server acts purely as encrypted storage.

---

## Cryptographic Architecture

### Key Types

nvolt uses a hybrid cryptographic system combining **asymmetric** and **symmetric** encryption:

| Key Type | Algorithm | Use Case | Location |
|----------|-----------|----------|----------|
| **Machine Private Key** | RSA-2048 | Unwrap symmetric keys | Client only |
| **Machine Public Key** | RSA-2048 | Wrap symmetric keys | Server + Client |
| **Symmetric Key** | AES-256-GCM | Encrypt secret values | Encrypted on server |

### Why Hybrid Cryptography?

- **RSA (Asymmetric):** Secure key distribution, per-machine access control
- **AES (Symmetric):** Fast encryption/decryption of large data

RSA is too slow for encrypting large amounts of data, so we use it only to encrypt small symmetric keys. The symmetric keys then encrypt the actual secrets.

---

## Wrapped Key Encryption

The **wrapped key approach** is the foundation of nvolt's security model.

### How It Works

#### Step 1: Generate Symmetric Key

When you push a secret:

```
Secret: DATABASE_URL=postgres://...
         ↓
Generate random symmetric key: K = <32-byte-random>
```

#### Step 2: Encrypt Secret with Symmetric Key

```
Secret (plaintext)
         ↓ AES-256-GCM encryption with key K
         ↓
Encrypted Secret (ciphertext)
```

#### Step 3: Wrap Symmetric Key for Each Machine

Fetch public keys for all authorized machines:

```
Machine A Public Key: PubKey_A
Machine B Public Key: PubKey_B
Machine C Public Key: PubKey_C

Symmetric Key K
         ↓ RSA encryption with PubKey_A
Wrapped Key for A: WrappedKey_A

Symmetric Key K
         ↓ RSA encryption with PubKey_B
Wrapped Key for B: WrappedKey_B

Symmetric Key K
         ↓ RSA encryption with PubKey_C
Wrapped Key for C: WrappedKey_C
```

#### Step 4: Send to Server

```json
{
  "key": "DATABASE_URL",
  "encrypted_value": "<AES-encrypted-secret>",
  "wrapped_keys": {
    "machine-a": "<RSA-encrypted-K-for-A>",
    "machine-b": "<RSA-encrypted-K-for-B>",
    "machine-c": "<RSA-encrypted-K-for-C>"
  }
}
```

The server stores this blob. It cannot decrypt anything because:
- It doesn't have the private keys to unwrap K
- It doesn't have K to decrypt the secret value

### Pulling Secrets (Decryption)

When Machine A pulls secrets:

#### Step 1: Request from Server

```
Machine A → Server: "Give me secrets for project X, environment Y"
```

#### Step 2: Server Responds

```json
{
  "key": "DATABASE_URL",
  "encrypted_value": "<AES-encrypted-secret>",
  "wrapped_key": "<RSA-encrypted-K-for-A>"
}
```

#### Step 3: Unwrap Symmetric Key

```
Wrapped Key
         ↓ RSA decryption with PrivKey_A (on Machine A)
         ↓
Symmetric Key K (recovered)
```

#### Step 4: Decrypt Secret

```
Encrypted Secret
         ↓ AES-256-GCM decryption with key K
         ↓
Plaintext Secret: DATABASE_URL=postgres://...
```

### Security Properties

✅ **Server compromise:** Server only has encrypted data  
✅ **Network interception:** TLS + encrypted payloads  
✅ **Per-machine access:** Only machines with wrapped keys can decrypt  
✅ **Key rotation:** Re-wrap keys without re-encrypting secrets  
✅ **Access revocation:** Remove wrapped key for a machine  

---

## Challenge-Response Authentication

For CI/CD silent authentication, nvolt uses **challenge-response** to prove possession of a private key without transmitting it.

### The Protocol

```
┌────────────┐                          ┌────────────┐
│   Client   │                          │   Server   │
│  (CI/CD)   │                          │            │
└──────┬─────┘                          └──────┬─────┘
       │                                       │
       │  1. Request Challenge                 │
       │  machine_id=github-actions            │
       ├──────────────────────────────────────>│
       │                                       │
       │  2. Generate Random Challenge         │
       │  challenge = <random-256-bit>         │
       │  challenge_id = <uuid>                │
       │  <────────────────────────────────────┤
       │  Encrypt challenge with PubKey        │
       │  encrypted_challenge = RSA(challenge) │
       │                                       │
       │  3. Decrypt Challenge with PrivKey    │
       │  challenge = RSA_decrypt(encrypted)   │
       │  Sign challenge: sig = Sign(challenge)│
       │                                       │
       │  4. Send Signature                    │
       │  challenge_id, signature              │
       ├──────────────────────────────────────>│
       │                                       │
       │  5. Verify Signature with PubKey      │
       │  valid = Verify(signature, challenge) │
       │  If valid: Issue JWT token            │
       │  <────────────────────────────────────┤
       │  JWT token                            │
       │                                       │
```

### Why This Is Secure

1. **Private key never transmitted:** Client proves possession without sending the key
2. **Replay protection:** Each challenge is used once (unique `challenge_id`)
3. **No shared secrets:** Server only has public key
4. **Man-in-the-middle resistant:** TLS + cryptographic proof

### Comparison with Other Approaches

| Approach | Security | nvolt Uses |
|----------|----------|------------|
| API keys stored in CI | ⚠️ Medium (key compromise = full access) | ❌ No |
| OAuth device flow | ✅ High (requires browser) | ❌ No (not CI-friendly) |
| Challenge-response with key pair | ✅ High (cryptographic proof) | ✅ **Yes** |

---

## Threat Model

### What nvolt Protects Against

#### ✅ Server Compromise

**Threat:** Attacker gains access to nvolt server database.

**Protection:** 
- All secrets are encrypted with AES-256-GCM
- Private keys are never sent to the server
- Attacker gets only encrypted blobs

**Outcome:** Secrets remain secure.

#### ✅ Network Interception (MITM)

**Threat:** Attacker intercepts network traffic between client and server.

**Protection:**
- TLS encryption for transport
- Secrets are already encrypted before transmission
- Even if TLS is broken, attacker gets encrypted data

**Outcome:** Secrets remain secure.

#### ✅ Unauthorized Machine Access

**Threat:** Attacker tries to access secrets from an unauthorized machine.

**Protection:**
- Server only sends wrapped keys for authorized machines
- Without the private key, attacker cannot unwrap the symmetric key
- Without the symmetric key, attacker cannot decrypt secrets

**Outcome:** Access denied.

#### ✅ Insider Threats

**Threat:** Malicious employee with server access tries to steal secrets.

**Protection:**
- Server stores only encrypted data
- RBAC limits access per user
- Environment-level permissions restrict production access
- Audit logs track all secret access

**Outcome:** Limited blast radius.

---

### What nvolt Does NOT Protect Against

#### ❌ Compromised Developer Machine

**Threat:** Attacker gains access to a developer's laptop with nvolt credentials.

**Risk:** 
- Private key is stored locally in `~/.nvolt/`
- Attacker can decrypt all secrets the compromised machine has access to

**Mitigation:**
- Use full disk encryption (FileVault, BitLocker, LUKS)
- Set strong login password
- Enable screen lock on idle
- Revoke machine access immediately if compromised:
  ```bash
  # From another machine
  nvolt sync  # Re-wrap keys excluding compromised machine
  ```

#### ❌ Malicious CLI Binary

**Threat:** User installs a modified nvolt CLI that exfiltrates secrets.

**Risk:**
- Malicious binary has access to decrypted secrets
- Could send plaintext secrets to attacker

**Mitigation:**
- Download from official sources only: https://install.nvolt.io
- Verify GPG signatures (future feature)
- Use package managers (Homebrew, apt) once available
- Monitor official releases and security advisories

#### ❌ Supply Chain Attacks

**Threat:** Compromised dependency in nvolt's build process.

**Risk:** Backdoored binary could steal secrets.

**Mitigation:**
- Pin dependency versions
- Use reproducible builds (future)
- Monitor security advisories
- Review code changes before updating

#### ❌ Social Engineering

**Threat:** Attacker tricks user into revealing private key.

**Risk:** Attacker gains access to all secrets.

**Mitigation:**
- Never share private keys via email, Slack, etc.
- Use encrypted channels for key transfer (CI/CD)
- Educate team on security practices

---

## Permissions Model

nvolt implements a **three-tier RBAC (Role-Based Access Control)** system:

### Tier 1: Organization Role

| Role | Capabilities |
|------|--------------|
| **admin** | • Add/remove users<br>• Add machines<br>• Grant permissions<br>• Full access to all secrets |
| **dev** | • Read/write secrets (per granted permissions)<br>• Cannot manage users or machines |

### Tier 2: Project-Level Permissions

*Note: Currently, nvolt enforces permissions at the environment level only. Project-level permissions are reserved for future use.*

### Tier 3: Environment-Level Permissions

Each user can have different permissions per environment:

| Permission | Grants Access To |
|------------|------------------|
| **read** | • `nvolt pull`<br>• `nvolt run` |
| **write** | • `nvolt push` (add/update secrets) |
| **delete** | • `nvolt push -f` (full replacement) |

### Examples

**Junior Developer:**
- Role: `dev`
- `development` environment: `read=true, write=true, delete=true`
- `staging` environment: `read=true, write=true, delete=false`
- `production` environment: `read=true, write=false, delete=false`

**Senior Developer:**
- Role: `dev`
- `development` environment: `read=true, write=true, delete=true`
- `staging` environment: `read=true, write=true, delete=true`
- `production` environment: `read=true, write=true, delete=false`

**DevOps Engineer:**
- Role: `admin`
- All environments: Full access (admin override)

---

## Best Practices

### 1. Use Full Disk Encryption

Protect private keys at rest:

**macOS:**
```bash
# Enable FileVault
System Preferences → Security & Privacy → FileVault → Turn On
```

**Linux:**
```bash
# LUKS encryption (during OS install)
# Or encrypt home directory:
sudo apt install ecryptfs-utils
ecryptfs-migrate-home -u $USER
```

**Windows:**
```
Settings → Update & Security → Device encryption → Turn On
```

### 2. Separate Machines for Production

Use dedicated machines for production access:

```bash
# Development machine
nvolt machine add alice-dev-laptop

# Production access (separate machine or ephemeral container)
nvolt machine add alice-prod-machine
```

Grant production access only to `alice-prod-machine`.

### 3. Rotate Keys Regularly

Quarterly key rotation for sensitive environments:

```bash
# Generate new key
nvolt machine add my-laptop-2024-Q4

# Sync to re-wrap secrets
nvolt sync

# Update local config
cp my-laptop-2024-Q4_key.pem ~/.nvolt/private_key.pem

# Remove old key from server (future feature)
# nvolt machine rm my-laptop-2024-Q3
```

### 4. Principle of Least Privilege

Grant minimum necessary permissions:

```bash
# Read-only production for junior devs
nvolt user mod junior@example.com -p my-app -e production \
  -a read=true,write=false,delete=false
```

### 5. Audit Regularly

Monthly access reviews:

```bash
# List users
nvolt user list

# List machines
nvolt machine list

# Review who has access to production
nvolt user mod user@example.com  # Check current permissions
```

### 6. Use Environment-Specific CI Machines

Separate machines for different environments:

```bash
nvolt machine add ci-staging
nvolt machine add ci-production
```

Store keys in environment-specific secrets.

### 7. Enable 2FA Everywhere

- nvolt account (OAuth provider: Google, GitHub)
- CI/CD platform (GitHub, GitLab, etc.)
- Cloud provider (AWS, GCP, Azure)

### 8. Monitor Audit Logs

Review server audit logs for:
- Failed authentication attempts
- Unusual access patterns
- Secret access from unexpected machines

### 9. Backup Configuration Securely

Backup `~/.nvolt/config.json` to encrypted storage:

```bash
# Encrypt with GPG
gpg -c ~/.nvolt/config.json

# Store encrypted file in password manager or secure cloud storage
```

### 10. Incident Response Plan

Have a plan for key compromise:

1. **Identify compromised machine**
2. **Remove machine access** (future: `nvolt machine rm`)
3. **Re-wrap keys** with `nvolt sync`
4. **Rotate all secrets** in affected environments
5. **Investigate** how compromise occurred
6. **Document** and share lessons learned

---

## Security Audit

### Self-Audit Checklist

Perform this audit quarterly:

#### Access Control

- [ ] Review all users: `nvolt user list`
- [ ] Remove former employees: `nvolt user rm`
- [ ] Verify admin role assignments
- [ ] Check production access is restricted

#### Machines

- [ ] List all machines: `nvolt machine list`
- [ ] Identify orphaned machines (old laptops, retired CI runners)
- [ ] Remove unused machines
- [ ] Verify CI machine keys are rotated

#### Secrets

- [ ] No plaintext secrets in git history
- [ ] `.nvolt/` in `.gitignore`
- [ ] No hardcoded secrets in code
- [ ] Production secrets separated from dev/staging

#### Infrastructure

- [ ] Full disk encryption enabled on all machines
- [ ] 2FA enabled on all accounts
- [ ] Password manager in use
- [ ] Regular OS and dependency updates

#### Policies

- [ ] Security policy documented
- [ ] Incident response plan exists
- [ ] Team trained on security practices
- [ ] Regular security reviews scheduled

---

## Compliance

nvolt's security model aligns with common compliance frameworks:

### SOC 2 Type II

- **Encryption:** AES-256-GCM, RSA-2048
- **Access Control:** RBAC with admin/dev roles
- **Audit Logging:** Server-side audit trail
- **Separation of Duties:** Admin vs. dev roles

### GDPR

- **Data Encryption:** At rest and in transit
- **Access Control:** Per-user, per-environment
- **Data Portability:** `nvolt pull` exports secrets
- **Right to Erasure:** `nvolt user rm` deletes access

### HIPAA

- **Encryption:** Meets encryption requirements
- **Access Control:** Granular permissions
- **Audit Trail:** Who accessed what, when
- **Unique User IDs:** Per-user, per-machine identity

### PCI DSS

- **Encryption:** Strong cryptography (AES-256)
- **Access Control:** Least privilege enforcement
- **Logging:** Audit trail of secret access
- **Key Management:** Wrapped key approach

*Note: nvolt provides security controls but does not guarantee compliance. Consult with a compliance expert for your specific use case.*

---

## Related Documentation

- **[Authentication](commands/authentication.md)** — Login and key management
- **[Machines](commands/machines.md)** — Per-machine key pairs
- **[Users](commands/users.md)** — Access control and permissions

---

[← Back to Documentation Home](README.md)

