# CI/CD Integration

Learn how to integrate nvolt into your CI/CD pipelines for secure secret management in automated environments.

## Table of Contents

- [Overview](#overview)
- [Setup Steps](#setup-steps)
- [GitHub Actions](#github-actions)
- [GitLab CI](#gitlab-ci)
- [CircleCI](#circleci)
- [Jenkins](#jenkins)
- [Bitbucket Pipelines](#bitbucket-pipelines)
- [Docker](#docker)
- [Security Best Practices](#security-best-practices)

---

## Overview

nvolt's silent authentication mode enables secure secret management in CI/CD without browser interaction.

### How It Works

```
┌─────────────────────┐
│  1. Generate Key    │
│  (Developer Machine)│
│  nvolt machine add  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  2. Store Key       │
│  (CI/CD Secrets)    │
│  NVOLT_PRIVATE_KEY  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  3. Silent Login    │
│  (CI/CD Runner)     │
│  nvolt login --silent│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  4. Run with Secrets│
│  nvolt run -c "..."  │
└─────────────────────┘
```

### Prerequisites

- **nvolt account** with organization access
- **Admin or dev role** in the organization
- **CI/CD platform** account with secret storage capability

---

## Setup Steps

Follow these steps for any CI/CD platform:

### Step 1: Generate Machine Key

On your local machine (where you're logged into nvolt):

```bash
nvolt machine add <ci-machine-name>
```

Examples:
- `github-actions`
- `gitlab-ci-production`
- `circleci-staging`

**Output:**
```
🔑 Generating keys for machine: github-actions
✓ Public key saved to server
✓ Private key saved to: /path/to/github-actions_key.pem

📋 Next Steps
1. Securely transfer the key file to the destination machine
2. Authenticate with: nvolt login --silent --machine github-actions --org org-xyz
```

### Step 2: Get Organization ID

```bash
cat ~/.nvolt/config.json | grep active_org
```

Output:
```
"active_org": "org-abc123def456"
```

### Step 3: Store Secrets in CI/CD Platform

Store these as **secret environment variables**:

- `NVOLT_PRIVATE_KEY`: Contents of `<ci-machine-name>_key.pem`
- `NVOLT_ORG_ID`: Your organization ID from step 2

### Step 4: Delete Local Key File

```bash
rm <ci-machine-name>_key.pem
```

### Step 5: Configure CI/CD Pipeline

Use the examples below for your specific platform.

---

## GitHub Actions

### Basic Workflow

`.github/workflows/deploy.yml`:

```yaml
name: Deploy
on:
  push:
    branches: [main]

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
      
      - name: Deploy Application
        run: |
          nvolt run -p my-app -e production -c "./deploy.sh"
        env:
          NVOLT_ORG_ID: ${{ secrets.NVOLT_ORG_ID }}
```

### Multi-Environment Workflow

```yaml
name: Deploy
on:
  push:
    branches: [main, develop]

jobs:
  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup nvolt
        run: |
          curl -sL https://install.nvolt.io/cli | bash
          mkdir -p ~/.nvolt
          echo "${{ secrets.NVOLT_PRIVATE_KEY }}" > ~/.nvolt/private_key.pem
          chmod 600 ~/.nvolt/private_key.pem
      
      - name: Deploy to Staging
        run: nvolt run -p my-app -e staging -c "./deploy.sh staging"

  deploy-production:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production  # GitHub environment protection
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup nvolt
        run: |
          curl -sL https://install.nvolt.io/cli | bash
          mkdir -p ~/.nvolt
          echo "${{ secrets.NVOLT_PRIVATE_KEY_PROD }}" > ~/.nvolt/private_key.pem
          chmod 600 ~/.nvolt/private_key.pem
      
      - name: Deploy to Production
        run: nvolt run -p my-app -e production -c "./deploy.sh production"
```

### Storing Secrets in GitHub

1. Go to repository **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Add `NVOLT_PRIVATE_KEY`:
   ```bash
   cat github-actions_key.pem | pbcopy  # Copy to clipboard
   ```
   Paste into GitHub secret value field
4. Add `NVOLT_ORG_ID` with your org ID
5. Click **Add secret**

### Using GitHub Environments

For production workflows with approval gates:

```yaml
jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment: 
      name: production
      url: https://myapp.com
    steps:
      # Same as above, but uses environment-specific secrets
      - name: Setup nvolt
        run: |
          curl -sL https://install.nvolt.io/cli | bash
          mkdir -p ~/.nvolt
          echo "${{ secrets.NVOLT_PRIVATE_KEY }}" > ~/.nvolt/private_key.pem
          chmod 600 ~/.nvolt/private_key.pem
```

---

## GitLab CI

### Basic Pipeline

`.gitlab-ci.yml`:

```yaml
stages:
  - deploy

variables:
  NVOLT_CONF: ".nvolt"

before_script:
  - curl -sL https://install.nvolt.io/cli | bash
  - mkdir -p ~/.nvolt
  - echo "$NVOLT_PRIVATE_KEY" > ~/.nvolt/private_key.pem
  - chmod 600 ~/.nvolt/private_key.pem

deploy-staging:
  stage: deploy
  script:
    - nvolt run -p my-app -e staging -c "./deploy.sh"
  only:
    - develop

deploy-production:
  stage: deploy
  script:
    - nvolt run -p my-app -e production -c "./deploy.sh"
  only:
    - main
  when: manual  # Require manual approval
```

### Using GitLab Environments

```yaml
deploy-production:
  stage: deploy
  environment:
    name: production
    url: https://myapp.com
  script:
    - nvolt run -p my-app -e production -c "./deploy.sh"
  only:
    - main
  when: manual
```

### Storing Secrets in GitLab

1. Go to **Settings** → **CI/CD** → **Variables**
2. Add `NVOLT_PRIVATE_KEY`:
   - **Key:** `NVOLT_PRIVATE_KEY`
   - **Value:** Contents of `gitlab-ci_key.pem`
   - **Type:** File or Variable
   - **Protected:** Yes (for production)
   - **Masked:** Yes
3. Add `NVOLT_ORG_ID` similarly

---

## CircleCI

### Basic Configuration

`.circleci/config.yml`:

```yaml
version: 2.1

jobs:
  deploy:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      
      - run:
          name: Install nvolt
          command: |
            curl -sL https://install.nvolt.io/cli | bash
            mkdir -p ~/.nvolt
            echo "$NVOLT_PRIVATE_KEY" > ~/.nvolt/private_key.pem
            chmod 600 ~/.nvolt/private_key.pem
      
      - run:
          name: Deploy with secrets
          command: |
            nvolt run -p my-app -e production -c "./deploy.sh"

workflows:
  version: 2
  deploy-workflow:
    jobs:
      - deploy:
          filters:
            branches:
              only: main
```

### With Approval Gate

```yaml
workflows:
  version: 2
  deploy-workflow:
    jobs:
      - hold:
          type: approval
          filters:
            branches:
              only: main
      
      - deploy:
          requires:
            - hold
```

### Storing Secrets in CircleCI

1. Go to **Project Settings** → **Environment Variables**
2. Add `NVOLT_PRIVATE_KEY` with the private key content
3. Add `NVOLT_ORG_ID` with your org ID

---

## Jenkins

### Pipeline Script

`Jenkinsfile`:

```groovy
pipeline {
    agent any
    
    environment {
        NVOLT_PRIVATE_KEY = credentials('nvolt-private-key')
        NVOLT_ORG_ID = credentials('nvolt-org-id')
    }
    
    stages {
        stage('Setup') {
            steps {
                sh '''
                    curl -sL https://install.nvolt.io/cli | bash
                    mkdir -p ~/.nvolt
                    cp $NVOLT_PRIVATE_KEY ~/.nvolt/private_key.pem
                    chmod 600 ~/.nvolt/private_key.pem
                '''
            }
        }
        
        stage('Deploy') {
            steps {
                sh '''
                    nvolt run -p my-app -e production -c "./deploy.sh"
                '''
            }
        }
    }
    
    post {
        always {
            sh 'rm -rf ~/.nvolt'
        }
    }
}
```

### Declarative Pipeline with Input

```groovy
pipeline {
    agent any
    
    stages {
        stage('Approval') {
            when {
                branch 'main'
            }
            steps {
                input message: 'Deploy to production?', ok: 'Deploy'
            }
        }
        
        stage('Deploy') {
            steps {
                sh '''
                    curl -sL https://install.nvolt.io/cli | bash
                    mkdir -p ~/.nvolt
                    echo "$NVOLT_PRIVATE_KEY" > ~/.nvolt/private_key.pem
                    chmod 600 ~/.nvolt/private_key.pem
                    nvolt run -p my-app -e production -c "./deploy.sh"
                '''
            }
        }
    }
}
```

### Storing Secrets in Jenkins

1. Go to **Manage Jenkins** → **Manage Credentials**
2. Add **Secret file** or **Secret text**:
   - ID: `nvolt-private-key`
   - Content: Private key from `jenkins_key.pem`
3. Add org ID similarly

---

## Bitbucket Pipelines

### Basic Configuration

`bitbucket-pipelines.yml`:

```yaml
image: atlassian/default-image:latest

pipelines:
  branches:
    main:
      - step:
          name: Deploy to Production
          deployment: production
          script:
            - curl -sL https://install.nvolt.io/cli | bash
            - mkdir -p ~/.nvolt
            - echo "$NVOLT_PRIVATE_KEY" > ~/.nvolt/private_key.pem
            - chmod 600 ~/.nvolt/private_key.pem
            - nvolt run -p my-app -e production -c "./deploy.sh"
    
    develop:
      - step:
          name: Deploy to Staging
          deployment: staging
          script:
            - curl -sL https://install.nvolt.io/cli | bash
            - mkdir -p ~/.nvolt
            - echo "$NVOLT_PRIVATE_KEY" > ~/.nvolt/private_key.pem
            - chmod 600 ~/.nvolt/private_key.pem
            - nvolt run -p my-app -e staging -c "./deploy.sh"
```

### Storing Secrets in Bitbucket

1. Go to **Repository Settings** → **Pipelines** → **Repository variables**
2. Add `NVOLT_PRIVATE_KEY` (mark as **Secured**)
3. Add `NVOLT_ORG_ID`

---

## Docker

### Dockerfile with nvolt

```dockerfile
FROM node:18-alpine

# Install nvolt
RUN curl -sL https://install.nvolt.io/cli | sh

# Set working directory
WORKDIR /app

# Copy application
COPY . .

# Install dependencies
RUN npm install

# Run with nvolt (expects private key at runtime)
CMD ["sh", "-c", "nvolt run -p my-app -e production -c 'npm start'"]
```

### Docker Compose with nvolt

`docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    volumes:
      - ${HOME}/.nvolt:/root/.nvolt:ro  # Mount nvolt config as read-only
    environment:
      - NVOLT_ORG_ID=${NVOLT_ORG_ID}
```

Run:
```bash
docker-compose up
```

### Docker Secret Management

Using Docker secrets (Swarm mode):

```bash
# Create secret
echo "$(cat ci-runner_key.pem)" | docker secret create nvolt_key -

# Use in service
docker service create \
  --name my-app \
  --secret nvolt_key \
  --env NVOLT_ORG_ID=org-xyz \
  my-app-image
```

Inside container:
```bash
mkdir -p ~/.nvolt
cp /run/secrets/nvolt_key ~/.nvolt/private_key.pem
chmod 600 ~/.nvolt/private_key.pem
nvolt run -p my-app -e production -c "npm start"
```

---

## Security Best Practices

### 1. Use Separate Keys for Different Environments

Generate different machine keys for staging vs. production:

```bash
# Staging CI
nvolt machine add ci-staging

# Production CI
nvolt machine add ci-production
```

Store in environment-specific secrets.

### 2. Restrict Secret Access

On platforms that support it, limit secret access:

- **GitHub:** Use environment protection rules
- **GitLab:** Mark variables as "Protected" (only protected branches)
- **CircleCI:** Use Contexts with team restrictions

### 3. Rotate Keys Regularly

```bash
# Quarterly rotation
nvolt machine add github-actions-2024-q4
nvolt sync

# Update GitHub secret with new key

# Remove old machine key (future feature)
# nvolt machine rm github-actions-2024-q3
```

### 4. Audit CI/CD Access

```bash
# List all machines
nvolt machine list

# Look for CI machines
# Review creation dates
```

### 5. Use Ephemeral Runners

For self-hosted runners, use ephemeral instances that are destroyed after each job. This ensures private keys don't persist.

### 6. Enable 2FA on CI/CD Platform

Protect access to your CI/CD platform with two-factor authentication.

### 7. Monitor CI/CD Logs

Review pipeline logs for:
- Unexpected nvolt commands
- Failed authentication attempts
- Access to sensitive environments

### 8. Principle of Least Privilege

Grant CI machines only the access they need:

```bash
# Read-only for test runners
nvolt user mod ci-test@example.com -p my-app -e staging \
  -a read=true,write=false,delete=false
```

### 9. Separate Machine per Pipeline

For critical systems, use separate machines per pipeline:

- `github-actions-build`
- `github-actions-test`
- `github-actions-deploy-staging`
- `github-actions-deploy-production`

### 10. Never Log Private Keys

Ensure your CI/CD platform masks secrets in logs. Test:

```yaml
# This should be masked in logs
- run: echo "$NVOLT_PRIVATE_KEY"
```

---

## Troubleshooting

### "failed to read private key"

**Error:**
```
failed to read private key from ~/.nvolt/private_key.pem: no such file or directory
```

**Solution:**
- Verify secret is correctly set in CI/CD platform
- Check file creation command in pipeline script
- Ensure `mkdir -p ~/.nvolt` runs before writing key

### "authentication failed: machine not found"

**Error:**
```
authentication failed: machine 'github-actions' not found in organization
```

**Solution:**
- Run `nvolt machine add github-actions` from your local machine
- Verify you're using the correct machine name in `--machine` flag
- Ensure machine was created in the correct organization

### "permission denied"

**Error:**
```
You don't have write permission for this environment.
```

**Solution:**
- Check CI machine's permissions: Ask admin to run `nvolt user mod ci-user@example.com`
- Verify you're targeting the correct environment
- Ensure the user associated with the machine has appropriate access

### Private Key Masked in Logs

If you see `***` instead of the key content:

**Cause:** CI platform is (correctly) masking the secret.

**Solution:** This is expected behavior. The key is still written correctly to the file.

---

## More Examples

For more CI/CD examples and templates, check the [documentation](https://docs.nvolt.io) or contact support@nvolt.io.

---

## Related Documentation

- **[nvolt login --silent](commands/authentication.md#silent-login-for-cicd)** — Silent authentication details
- **[nvolt machine add](commands/machines.md#nvolt-machine-add)** — Generate machine keys
- **[nvolt run](commands/secrets.md#nvolt-run)** — Run commands with secrets

---

[← Back to Documentation Home](README.md)

