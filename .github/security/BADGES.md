# 🔒 Security Scanning - Badge Configuration

Add these badges to your main README.md file:

## Main Security Badge

```markdown
![Security Scan](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main)
```

## Individual Scan Badges (Optional)

If you want badges for specific scan types, you can use workflow badges with job filters:

```markdown
### SAST
![SAST Semgrep](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main&job=semgrep-sast)
![SAST CodeQL](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main&job=codeql-sast)

### Dependency Scanning
![Deps Go](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main&job=dependency-scan-go)
![Deps Node](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main&job=dependency-scan-node)

### Secrets Detection
![Secrets](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main&job=secrets-scan)

### Container Scanning
![Container](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main&job=container-scan)
```

## GitHub Security Tab Badge

```markdown
[![Security Status](https://img.shields.io/badge/security-monitored-brightgreen.svg)](https://github.com/OWNER/REPO/security)
```

## Snyk Badge (if using Snyk)

Replace `PROJECT_ID` with your actual Snyk project ID:

```markdown
[![Known Vulnerabilities](https://snyk.io/test/github/OWNER/REPO/badge.svg)](https://snyk.io/test/github/OWNER/REPO)
```

## Complete Badge Section for README.md

Add this section to the top of your README.md:

```markdown
# Renderowl 2.0

![Security Scan](https://github.com/OWNER/REPO/actions/workflows/security-scan.yml/badge.svg?branch=main)
[![Security Status](https://img.shields.io/badge/security-monitored-brightgreen.svg)](https://github.com/OWNER/REPO/security)

A powerful video rendering platform built with Go and Next.js.

---
```

> ⚠️ **Note:** Replace `OWNER/REPO` with your actual GitHub username and repository name.
