# 🔒 Security Scanning Setup Guide

This document describes the security scanning setup for Renderowl 2.0.

## Overview

The security scanning workflow runs on every push to `main`, `master`, or `develop` branches, on every pull request to these branches, and on a daily schedule.

## Scans Included

| Scan Type | Tool | Languages | Purpose |
|-----------|------|-----------|---------|
| SAST | Semgrep | Go, TypeScript | Static Application Security Testing |
| SAST | CodeQL | Go, TypeScript | GitHub's static analysis engine |
| Dependencies | govulncheck, Nancy | Go | Go module vulnerability scanning |
| Dependencies | npm audit, Snyk | Node.js | JavaScript dependency scanning |
| Secrets | GitLeaks, TruffleHog | All | Hardcoded secret detection |
| Containers | Trivy | Docker | Container image vulnerability scanning |

## Required Secrets

Configure these secrets in your GitHub repository settings:

### Optional Secrets

| Secret | Purpose | Required? |
|--------|---------|-----------|
| `SNYK_TOKEN` | Snyk dependency scanning | No (falls back to npm audit) |
| `SEMGREP_APP_TOKEN` | Semgrep Cloud integration | No (local rules still work) |

### How to Add Secrets

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Add each secret name and value

## Configuration Files

| File | Purpose |
|------|---------|
| `.github/workflows/security-scan.yml` | Main workflow definition |
| `.github/security/semgrep-rules.yml` | Semgrep custom rules |
| `.github/security/codeql-config.yml` | CodeQL configuration |
| `.github/security/gitleaks.toml` | GitLeaks configuration |
| `.github/security/trivy.yaml` | Trivy scanner configuration |
| `.github/security/.trivyignore` | Vulnerabilities to ignore |

## Viewing Results

### GitHub Security Tab

All scan results are uploaded to the GitHub Security tab:

1. Go to **Security** → **Code scanning alerts**
2. Filter by tool (semgrep-sast, codeql, snyk, trivy, etc.)

### Workflow Artifacts

Detailed reports are available as workflow artifacts:

1. Go to **Actions** → **Security Scan** → Select a run
2. Scroll down to **Artifacts** section
3. Download individual scan results

### SARIF Files

All scanners output SARIF format compatible with GitHub's security dashboard.

## Failing Builds

The workflow will fail on:
- Critical or High severity SAST findings (Semgrep)
- Secrets detection (GitLeaks, TruffleHog)
- Critical container vulnerabilities (Trivy)

To adjust severity thresholds, edit the workflow file.

## Ignoring False Positives

### Semgrep

Add a comment to ignore specific findings:

```go
// nosemgrep: hardcoded-credentials-go
const testPassword = "test123" // This is test-only
```

### Trivy

Add CVE IDs to `.github/security/.trivyignore`:

```
CVE-2021-12345 # False positive, not applicable
```

### CodeQL

Create a `.github/codeql/codeql-config.yml` with exclusion rules.

## Custom Rules

### Adding Semgrep Rules

Edit `.github/security/semgrep-rules.yml` and add new rules following the Semgrep syntax.

### Adding GitLeaks Rules

Edit `.github/security/gitleaks.toml` and add new `[[rules]]` sections.

## Troubleshooting

### Workflow fails but no alerts shown

Check the workflow logs for detailed error messages. Some scanners may fail silently.

### SARIF upload fails

Ensure the `security-events: write` permission is set in the workflow.

### Secrets scan false positives

Update `.github/security/gitleaks.toml` allowlist section with patterns to ignore.

## Support

- [Semgrep Documentation](https://semgrep.dev/docs/)
- [CodeQL Documentation](https://codeql.github.com/docs/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [GitLeaks Documentation](https://github.com/gitleaks/gitleaks)
- [TruffleHog Documentation](https://github.com/trufflesecurity/trufflehog)
