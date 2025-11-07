# CLAUDE.md

## 🧭 Purpose

This document defines the **development instructions and constraints** for building **nvolt**, a **GitHub-native, Zero-Trust CLI** for managing encrypted environment variables without a centralized backend, login, or organization model.

Claude (or any LLM agent) must **strictly follow these specifications** when generating code, commands, or project structure.

---

## 🧩 Core Concept

nvolt is a **Zero-Trust, cryptographically enforced secret manager** built entirely around **Git** and **local files**.

- **No server** — all data lives in GitHub (or any Git-compatible repo).
- **No JWT, OAuth, or API** — authentication is handled by Git.
- **All encryption/decryption happens locally** using per-machine keypairs.
- **Access control** is defined by the presence of wrapped key files.
- **`.nvolt/` directories** act as encrypted, committed `.env` replacements — safe to include in any repo.

---

## 🧠 Developer Experience Overview

nvolt operates in **two layers of responsibility**:

### 1️⃣ Local Mode (always-on)

- Works entirely with **local files** (acts like a secure `.env` manager).
- Performs encryption, decryption, and key management.
- **Never performs Git operations.**

### 2️⃣ Global Mode (optional)

- Can manage a **dedicated global GitHub repository** for shared secrets.
- Performs **safe Git helper operations** (clone, pull, push) in `~/.nvolt/projects/`.
- **Never** modifies Git state in the user’s application repositories.

---

## 🧱 Core Concepts

### **Project**

Represents the **application name**.

- By default, nvolt automatically detects the project name from:

  - `package.json` → `name`
  - `go.mod` → `module`
  - `Cargo.toml` → `package.name`
  - `pyproject.toml` → `project.name`
  - Otherwise, falls back to **current directory name**.

- The project name can be overridden anytime using:
  ```bash
  nvolt push -p my-project
  ```

### **Environment**

Represents a configuration tier, e.g.:

```
default, staging, production, ci
```

- Defines which secret set to use within `.nvolt/secrets/`.
- Can be overridden using:
  ```bash
  nvolt pull -e staging
  ```

---

## ⚙️ CLI Command: `nvolt init`

Initializes machine keys and sets up the vault structure.

### Flow

1. **Check for machine keypair**

   - If `~/.nvolt/private_key.pem` doesn’t exist:
     - Generate RSA or Ed25519 keypair.
     - Store private key at `~/.nvolt/private_key.pem`.
     - Store public key metadata at `~/.nvolt/machines/machine-info.json`.

2. **If `--repo <org/repo>` provided:**

   - Clone the Git repository into `~/.nvolt/projects/<org>/<repo>` if not already present.
     ```bash
     git clone git@github.com:<org>/<repo>.git ~/.nvolt/projects/<org>/<repo>
     ```
   - Subsequent nvolt operations in this vault may automatically perform safe Git operations (`pull`, `commit`, `push`).

3. **If no repo provided:**
   - Initialize `.nvolt/` in the **current directory**.
   - In this mode, **nvolt never performs Git commands** — all Git management is user-controlled.

### Summary

| Mode                        | Storage                            | Git Behavior          |
| --------------------------- | ---------------------------------- | --------------------- |
| `nvolt init --repo org/app` | `~/.nvolt/projects/org/app/.nvolt` | nvolt manages Git ops |
| `nvolt init`                | `./.nvolt`                         | user manages Git ops  |

---

## ⚙️ Git Integration Rules

1. **nvolt never handles Git authentication.**

   - If Git commands fail, surface the error and exit gracefully.

2. **Git operations always use `-C` flag:**

   ```bash
   git -C ~/.nvolt/projects/org/app pull
   git -C ~/.nvolt/projects/org/app push
   ```

3. **No directory switching (`cd`)** — always use subprocess with `-C`.

4. **Only perform Git operations** for repos in `~/.nvolt/projects/`.

5. **Never perform Git operations** in local app directories.

---

## 🔐 Cryptographic Model

- **Per-machine RSA or Ed25519 keypairs**

  - Private key: `~/.nvolt/private_key.pem`
  - Public key: `.nvolt/machines/<machine>.json`

- **Project master key**
  - Generated locally on first `push`.
  - Stored only as _wrapped versions_ for authorized machines.

### **Encryption Flow**

1. Encrypt secret values using the project master key (AES-GCM).
2. Wrap master key for each machine’s public key.
3. Store wrapped key in `.nvolt/wrapped_keys/`.

### **Decryption Flow**

1. Load machine’s wrapped key from `.nvolt/wrapped_keys/`.
2. Decrypt wrapped key using the local private key.
3. Decrypt secrets using the project master key.

---

## 🧩 CLI Commands

| Command                                               | Description                                                       |
| ----------------------------------------------------- | ----------------------------------------------------------------- |
| `nvolt init [--repo <url>]`                           | Initializes `.nvolt/` or global vault. Creates repo if specified. |
| `nvolt machine add <name>`                            | Generates new keypair for CI or another device.                   |
| `nvolt machine rm <name>`                             | Revokes access and re-wraps master key.                           |
| `nvolt push [-f <envfile>] [-e <env>] [-p <project>]` | Encrypts and writes secrets.                                      |
| `nvolt pull [-e <env>] [-p <project>]`                | Decrypts and prints or writes secrets.                            |
| `nvolt run [-e <env>] [-c <cmd>]`                     | Loads decrypted secrets and runs command.                         |
| `nvolt sync [--rotate]`                               | Re-wraps or rotates master keys.                                  |
| `nvolt vault show`                                    | Displays machine access and key info.                             |
| `nvolt vault verify`                                  | Verifies integrity of encrypted files and keys.                   |

---

## 🧰 Implementation Notes

- Written in **Go** or **Rust**.
- Use **AES-GCM** for symmetric encryption.
- Use **RSA/Ed25519** for key wrapping.
- JSON for all structured files.

### Example encrypted secret

```json
{
  "version": 2,
  "data": "base64(aes_ciphertext)",
  "nonce": "base64(iv)",
  "tag": "base64(tag)"
}
```

### Example wrapped key

```json
{
  "machine_id": "m-abc123",
  "public_key_fingerprint": "SHA256:abcd1234",
  "wrapped_key": "base64(aes_ciphertext)",
  "granted_by": "alice@example.com",
  "granted_at": "2025-11-07T12:00:00Z"
}
```

---

## 📁 Directory Structures

### Global Mode

```
~/.nvolt/
  ├── private_key.pem
  ├── machines/
  │    └── m-localhost.json
  └── projects/
       └── myorg/app/
            ├── .git/
            └── .nvolt/
                 ├── secrets/
                 ├── wrapped_keys/
                 ├── machines/
                 └── keyinfo.json
```

### Local Mode

```
my-app/
  ├── .git/
  ├── .nvolt/
  │    ├── secrets/
  │    ├── wrapped_keys/
  │    ├── machines/
  │    └── keyinfo.json
  └── src/
```

---

## 🧭 Security Rules

1. **Never** send any data over the network except Git operations.
2. **Never** prompt for GitHub or GitLab credentials.
3. **Never** store plaintext secrets or keys on disk.
4. **Never** modify user Git configuration.
5. **Always** fail fast on Git or crypto errors.
6. **Always** verify presence of `~/.nvolt/private_key.pem`.
7. **Always** re-wrap or rotate keys upon machine removal.

---

## 🧠 Philosophy Recap

- **Git controls visibility** — who can fetch ciphertext.
- **nvolt controls meaning** — who can decrypt it.
- **Two responsibilities only:**
  1. Local file database operations.
  2. Optional Git helpers for global vaults.
- **No servers, no logins, no trust** — only cryptographic proof.
- **`.nvolt/` replaces `.env`** — auditable, encrypted config-as-code.

---

## ✅ Deliverables

Claude must:

1. Implement the full CLI as described above.
2. Follow Go/Rust idioms with minimal dependencies.
3. Support both local and global modes.
4. Never introduce online dependencies or login systems.
5. Update `./TASKS.md` when implementing features.

---

## ⚠️ Prohibited Actions

- Do **not** create additional markdown guides.
- Do **not** use OAuth or web APIs.
- Do **not** attempt to manage Git authentication.
- Do **not** log or print secrets.
- Do **not** change cryptographic primitives.

---

## 🧩 Quick Reference

| Concept              | Description                     |
| -------------------- | ------------------------------- |
| `.nvolt/`            | Local encrypted vault           |
| `~/.nvolt/projects/` | Global shared vaults            |
| `private_key.pem`    | Local machine private key       |
| `wrapped_keys/`      | Defines machine access          |
| `secrets/*.enc.json` | Encrypted environment variables |

---

**End of CLAUDE.md**
