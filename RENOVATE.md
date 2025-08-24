# Renovate Bot Configuration

This repository uses [Renovate Bot](https://renovatebot.com/) to automatically manage dependency updates. Renovate provides more sophisticated dependency management compared to the basic dependency update workflow.

## Configuration

The Renovate configuration is defined in `renovate.json` and includes:

### Key Features

- **Scheduled Updates**: Runs every Monday at 6 AM UTC to minimize disruption
- **Grouped Updates**: Groups related dependencies (e.g., HashiCorp Terraform packages, testing dependencies)
- **Security Alerts**: Automatically creates high-priority PRs for security vulnerabilities
- **Semantic Commits**: Uses conventional commit format for consistent changelog generation
- **Automerge**: Automatically merges low-risk updates like lock file maintenance

### Package Grouping

- **Terraform Framework**: All `github.com/hashicorp/terraform-*` packages
- **Testing Dependencies**: `testify`, `testcontainers-go`
- **Go Dependencies**: General Go module dependencies
- **Nix Flake Inputs**: Development environment dependencies
- **GitHub Actions**: Workflow dependency updates

### Update Strategy

- **Major Updates**: Require manual review (automerge disabled)
- **Security Updates**: High priority, require manual review
- **Minor/Patch Updates**: Can be automerged if tests pass
- **Lock Files**: Automatically maintained and merged

### Configuration Options

- **Timezone**: America/New_York
- **PR Limits**: Max 3 concurrent PRs, 2 per hour
- **Minimum Release Age**: 3 days for Go dependencies
- **Rebase Strategy**: Only when conflicted

## Workflows

### Renovate Workflow (`.github/workflows/renovate.yml`)

- Runs on schedule (Mondays at 6 AM UTC)
- Supports manual dispatch with configurable log levels
- Uses the latest Renovate GitHub Action (automatically updated)
- No hardcoded Renovate version - uses the latest stable release

### Existing Update Dependencies Workflow

The existing `update-deps.yml` workflow can work alongside Renovate or be replaced. Consider:

- **Keep Both**: Renovate for sophisticated management, manual workflow for emergency updates
- **Replace**: Disable the existing workflow in favor of Renovate's more advanced features

## Setup Requirements

### GitHub Token

Renovate requires a GitHub token with appropriate permissions:

1. **Using Default Token** (basic functionality):

   ```yaml
   token: ${{ secrets.GITHUB_TOKEN }}
   ```

2. **Using Custom Token** (recommended for advanced features):
   - Create a Personal Access Token with `repo`, `workflow`, and `user:email` scopes
   - Add it as a repository secret named `RENOVATE_TOKEN`

### Enabling Renovate

1. Install the [Renovate GitHub App](https://github.com/apps/renovate) on your repository
2. The configuration in `renovate.json` will be automatically detected
3. Renovate will start creating PRs according to the schedule

## Customization

### Adding Custom Package Rules

Edit `renovate.json` to add specific rules for new dependencies:

```json
{
  "packageRules": [
    {
      "description": "Custom rule for specific packages",
      "matchPackagePatterns": ["^your-package-pattern"],
      "groupName": "your group name",
      "automerge": false
    }
  ]
}
```

### Ignoring Dependencies

To ignore specific dependencies:

```json
{
  "ignoreDeps": [
    "package-name-to-ignore"
  ]
}
```

### Adjusting Schedule

Modify the schedule in both `renovate.json` and `.github/workflows/renovate.yml`:

```json
{
  "schedule": ["before 9am on monday"]
}
```

## Monitoring

- **Dependency Dashboard**: Renovate creates an issue with the current status
- **PR Labels**: All PRs are labeled with `dependencies` and specific update types
- **Security Alerts**: High-priority PRs for vulnerability fixes

## Best Practices

1. **Review Major Updates**: Always review major version updates manually
2. **Test Thoroughly**: Ensure your test suite is comprehensive
3. **Monitor Security**: Pay special attention to security-labeled PRs
4. **Update Configuration**: Regularly review and update Renovate configuration
5. **Coordinate with Team**: Ensure team members understand the automated update process
