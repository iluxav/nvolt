# nvolt Development Tasks

This document tracks the implementation progress of nvolt, a GitHub-native, Zero-Trust CLI for managing encrypted environment variables.

**Status Legend:**
- `[ ]` Not started
- `[~]` In progress
- `[x]` Completed

---

## Phase 1: Core Infrastructure ✅

### 1.1 Project Setup ✅
- [x] Initialize Go/Rust project structure
- [x] Set up dependency management (go.mod or Cargo.toml)
- [x] Configure linting and formatting tools
- [x] Create basic CLI framework (cobra/clap)
- [x] Set up unit testing infrastructure

### 1.2 Cryptographic Primitives ✅
- [x] Implement RSA/Ed25519 keypair generation
- [x] Implement AES-GCM encryption/decryption
- [x] Implement key wrapping (RSA-OAEP or similar)
- [x] Add secure key derivation functions
- [x] Create cryptographic utility functions
- [x] Add comprehensive crypto tests

### 1.3 File System Management ✅
- [x] Implement `.nvolt/` directory structure creation
- [x] Create `~/.nvolt/` home directory management
- [x] Implement safe file read/write operations
- [x] Add atomic file operations (write-then-rename)
- [x] Create directory traversal utilities
- [x] Add file permission verification

---

## Phase 2: Core Commands - Machine & Key Management 🚧

### 2.1 `nvolt init` Command ✅
- [x] Parse `--repo` flag for global/local mode detection
- [x] Implement local mode initialization (`./.nvolt/`)
- [x] Implement global mode initialization (`~/.nvolt/projects/`)
- [x] Generate machine keypair on first init
- [x] Create machine info file (`~/.nvolt/machines/machine-info.json`)
- [x] Implement Git clone for global repos
- [x] Add validation for existing vaults
- [x] Handle initialization errors gracefully

### 2.2 `nvolt machine add` Command ✅
- [x] Generate new keypair for specified machine
- [x] Create machine metadata JSON
- [x] Store public key in `.nvolt/machines/<name>.json`
- [x] Export private key for manual transfer
- [x] Add machine fingerprint calculation
- [x] Implement machine naming validation
- [x] Add duplicate machine detection

### 2.3 `nvolt machine rm` Command ✅
- [x] Remove machine from `.nvolt/machines/`
- [x] Remove machine's wrapped key from `.nvolt/wrapped_keys/`
- [~] Trigger automatic key re-wrapping
- [~] Update access logs
- [x] Add confirmation prompt
- [~] Handle last-machine removal protection

---

## Phase 3: Project Detection & Context ✅

### 3.1 Automatic Project Name Detection ✅
- [x] Implement `package.json` parser (Node.js)
- [x] Implement `go.mod` parser (Go)
- [x] Implement `Cargo.toml` parser (Rust)
- [x] Implement `pyproject.toml` parser (Python)
- [x] Add fallback to directory name
- [x] Create priority chain for detection
- [x] Add manual override via `-p` flag

### 3.2 Environment Management ✅
- [x] Implement environment detection logic
- [x] Create default environment handling
- [x] Add `-e` flag support across commands
- [x] Validate environment names
- [x] Create environment-specific secret paths

---

## Phase 4: Secret Management Commands ✅

### 4.1 `nvolt push` Command ✅
- [x] Parse `-f` flag for .env file input
- [x] Parse `-e` flag for environment selection
- [x] Parse `-p` flag for project override
- [x] Parse `-k` flag for inline key=value pairs
- [x] Read and parse .env file format
- [x] Generate project master key (if first push)
- [x] Encrypt secrets using AES-GCM
- [x] Create encrypted secret files in `.nvolt/secrets/<env>/`
- [x] Wrap master key for all machines
- [x] Store wrapped keys in `.nvolt/wrapped_keys/`
- [~] Update `keyinfo.json` metadata
- [~] Perform Git commit/push (global mode only)
- [x] Add validation for empty secrets

### 4.2 `nvolt pull` Command ✅
- [x] Parse `-e` flag for environment selection
- [x] Parse `-p` flag for project override
- [~] Perform Git pull (global mode only)
- [x] Load machine's wrapped key
- [x] Decrypt wrapped key using private key
- [x] Decrypt secrets using project master key
- [x] Output secrets in .env format
- [x] Add `--write` flag to save to .env file
- [x] Handle missing wrapped key errors
- [x] Validate decryption integrity

### 4.3 `nvolt run` Command ✅
- [x] Parse `-e` flag for environment selection
- [x] Parse `-c` flag for command to execute
- [x] Decrypt secrets into memory
- [x] Set environment variables in subprocess
- [x] Execute specified command
- [x] Clean up secrets from memory after execution
- [x] Handle command execution errors
- [x] Support command arguments and quotes

---

## Phase 5: Advanced Operations ✅

### 5.1 `nvolt sync` Command ✅
- [x] Implement basic re-wrapping logic
- [x] Add `--rotate` flag for master key rotation
- [x] Generate new master key on rotation
- [x] Re-encrypt all secrets with new key
- [x] Re-wrap new key for all machines
- [x] Remove old wrapped keys (handled by re-wrapping)
- [~] Update keyinfo.json with rotation metadata
- [~] Perform Git commit/push (global mode only)

### 5.2 `nvolt vault show` Command ✅
- [x] Display all registered machines
- [x] Show machine fingerprints
- [x] Display granted access timestamps
- [x] Show which machines have wrapped keys
- [x] List available environments
- [~] Show project master key metadata
- [x] Format output in readable table

### 5.3 `nvolt vault verify` Command ✅
- [x] Verify all encrypted files are readable
- [x] Verify wrapped keys for current machine
- [~] Validate keyinfo.json structure
- [x] Check for orphaned wrapped keys
- [~] Verify machine public keys match fingerprints
- [~] Validate Git repository state (global mode)
- [x] Report integrity status

---

## Phase 6: Git Integration ✅

### 6.1 Git Helper Functions ✅
- [x] Implement safe `git clone` with `-C` flag
- [x] Implement safe `git pull` with `-C` flag
- [x] Implement safe `git add` with `-C` flag
- [x] Implement safe `git commit` with `-C` flag
- [x] Implement safe `git push` with `-C` flag
- [x] Add Git error handling and surfacing
- [x] Never perform Git ops in local app directories
- [x] Add Git availability detection

### 6.2 Global Mode Git Automation ✅
- [x] Auto-pull before read operations (pull command)
- [x] Auto-commit after write operations (push, sync, machine add/rm)
- [x] Generate meaningful commit messages
- [x] Add conflict detection (SafePull)
- [x] Handle merge conflicts gracefully (SafePull)
- [x] Skip Git ops if not in global mode (mode detection)

---

## Phase 7: Data Structures & Formats

### 7.1 Encrypted Secret Format (v2)
- [ ] Implement JSON structure with version field
- [ ] Add base64-encoded ciphertext
- [ ] Add nonce/IV field
- [ ] Add authentication tag
- [ ] Support backward compatibility
- [ ] Add schema validation

### 7.2 Wrapped Key Format
- [ ] Implement machine_id field
- [ ] Add public_key_fingerprint
- [ ] Add wrapped_key (base64)
- [ ] Add granted_by metadata
- [ ] Add granted_at timestamp
- [ ] Support key rotation metadata

### 7.3 Machine Info Format
- [ ] Store machine identifier
- [ ] Store public key in PEM format
- [ ] Store fingerprint (SHA256)
- [ ] Add creation timestamp
- [ ] Add machine description/hostname

### 7.4 Keyinfo Metadata
- [ ] Track project master key version
- [ ] Store key rotation history
- [ ] Track authorized machines
- [ ] Add last modified timestamp

---

## Phase 8: Security Hardening

### 8.1 Key Management Security
- [ ] Ensure private keys never leave `~/.nvolt/`
- [ ] Set proper file permissions (0600 for keys)
- [ ] Implement secure key deletion
- [ ] Add key expiration support
- [ ] Validate key strength on generation

### 8.2 Encryption Security
- [ ] Use cryptographically secure random number generation
- [ ] Implement proper nonce generation
- [ ] Add authenticated encryption verification
- [ ] Prevent IV reuse
- [ ] Add timing-attack protections

### 8.3 Secret Handling
- [ ] Never log secrets to stdout/stderr
- [ ] Clear secrets from memory after use
- [ ] Prevent secrets in error messages
- [ ] Add secure string comparison
- [ ] Implement secret redaction in debug logs

### 8.4 Git Security
- [ ] Never store private keys in Git
- [ ] Validate .gitignore for sensitive files
- [ ] Prevent accidental plaintext commits
- [ ] Add pre-commit hooks (optional)

---

## Phase 9: Error Handling & UX

### 9.1 Error Handling
- [ ] Create custom error types
- [ ] Add context to all errors
- [ ] Implement graceful failure modes
- [ ] Add helpful error messages
- [ ] Create error recovery suggestions
- [ ] Add error codes for automation

### 9.2 User Experience
- [ ] Add progress indicators for long operations
- [ ] Implement colored output (optional)
- [ ] Add verbose/debug flags
- [ ] Create helpful command examples
- [ ] Add interactive prompts where needed
- [ ] Implement dry-run mode

### 9.3 Validation
- [ ] Validate all user inputs
- [ ] Check for required files before operations
- [ ] Validate Git repository state
- [ ] Verify encryption before writing
- [ ] Add sanity checks for key operations

---

## Phase 10: Testing & Documentation

### 10.1 Unit Tests
- [ ] Test cryptographic functions
- [ ] Test file operations
- [ ] Test Git integration
- [ ] Test project detection
- [ ] Test all CLI commands
- [ ] Achieve >80% code coverage

### 10.2 Integration Tests
- [ ] Test full init → push → pull flow
- [ ] Test machine addition/removal
- [ ] Test key rotation
- [ ] Test Git operations in global mode
- [ ] Test multiple environments
- [ ] Test error scenarios

### 10.3 End-to-End Tests
- [ ] Test real GitHub repository
- [ ] Test multi-machine scenarios
- [ ] Test CI/CD integration
- [ ] Test large secret files
- [ ] Test concurrent operations

### 10.4 Documentation
- [ ] Create comprehensive README.md
- [ ] Add command reference documentation
- [ ] Create architecture documentation
- [ ] Add security model documentation
- [ ] Create migration guides
- [ ] Add troubleshooting guide

---

## Phase 11: Advanced Features (Optional)

### 11.1 Secret Sharing
- [ ] Implement secret-level access control
- [ ] Add secret expiration
- [ ] Support secret versions/history
- [ ] Add secret audit logs

### 11.2 Multi-Repository Support
- [ ] Support multiple global repos
- [ ] Add repo switching commands
- [ ] Implement repo aliasing
- [ ] Add cross-repo secret referencing

### 11.3 CLI Enhancements
- [ ] Add shell completion (bash/zsh/fish)
- [ ] Create config file support
- [ ] Add command aliases
- [ ] Implement interactive mode
- [ ] Add JSON/YAML output formats

### 11.4 CI/CD Integration
- [ ] Create GitHub Actions example
- [ ] Create GitLab CI example
- [ ] Add CircleCI example
- [ ] Document Jenkins integration
- [ ] Create Docker container

---

## Phase 12: Release & Distribution

### 12.1 Build System
- [ ] Create cross-platform build scripts
- [ ] Set up GitHub Actions for releases
- [ ] Generate Linux binaries (amd64, arm64)
- [ ] Generate macOS binaries (Intel, Apple Silicon)
- [ ] Generate Windows binaries
- [ ] Create installation scripts

### 12.2 Distribution
- [ ] Publish to GitHub Releases
- [ ] Create Homebrew formula
- [ ] Create apt/deb packages
- [ ] Create rpm packages
- [ ] Submit to crates.io or Go modules
- [ ] Create Docker images

### 12.3 Versioning
- [ ] Implement semantic versioning
- [ ] Create CHANGELOG.md
- [ ] Add version command
- [ ] Implement upgrade checks
- [ ] Add migration tools for breaking changes

---

## Implementation Priority

### High Priority (MVP)
1. Phase 1: Core Infrastructure
2. Phase 2: Machine & Key Management
3. Phase 3: Project Detection
4. Phase 4.1-4.2: push and pull commands
5. Phase 6.1: Basic Git helpers
6. Phase 7: Data structures
7. Phase 8: Security hardening

### Medium Priority
8. Phase 4.3: run command
9. Phase 5: Advanced operations
10. Phase 9: Error handling & UX
11. Phase 10: Testing

### Low Priority (Post-MVP)
12. Phase 11: Advanced features
13. Phase 12: Distribution

---

## Notes

- Update this file as tasks are completed
- Mark tasks with `[x]` when done
- Add subtasks as needed during implementation
- Reference CLAUDE.md for all specification details
- Never deviate from the Zero-Trust, Git-native architecture

---

**Last Updated:** 2025-11-07
