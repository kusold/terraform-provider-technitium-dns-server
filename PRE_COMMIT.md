# Pre-commit Hooks Setup

This repository uses pre-commit hooks to ensure code quality and consistency before commits. We provide both a simple Git hook and a comprehensive pre-commit framework configuration.

## Quick Setup (Recommended)

Run the development setup task which includes pre-commit installation:

```bash
task dev-setup
```

Or manually install pre-commit hooks:

```bash
task pre-commit:install
```

## Available Approaches

### 1. Simple Git Hook (Already Active)

A basic pre-commit hook is installed at `.git/hooks/pre-commit` that runs `task fmt` before every commit.

**Features:**
- ✅ Automatically runs `task fmt` (which executes `go fmt ./...`)
- ✅ Prevents commits if formatting changes are needed
- ✅ Works immediately without additional setup
- ✅ No external dependencies

**Usage:**
- Runs automatically on every `git commit`
- If formatting is needed, the commit is aborted with instructions
- Add the formatted files and commit again

### 2. Pre-commit Framework (Advanced)

A comprehensive setup using the [pre-commit](https://pre-commit.com) framework with multiple hooks.

**Features:**
- ✅ Go formatting (`go fmt`)
- ✅ Import organization (`goimports`)
- ✅ Go module tidying (`go mod tidy`)
- ✅ Go vet checks
- ✅ Trailing whitespace removal
- ✅ YAML validation
- ✅ Large file detection
- ⚠️ Security scanning (disabled due to false positives in API docs)
- ⚠️ Markdown linting (disabled due to documentation formatting)
- ✅ Large file detection

**Prerequisites:**
- `pre-commit` and `goimports` are included in the Nix development environment
- Run `nix develop` or `direnv allow` to get the required tools

## Available Tasks

### Pre-commit Management
```bash
# Install pre-commit hooks
task pre-commit:install

# Run all pre-commit hooks on all files
task pre-commit:run

# Update pre-commit hooks to latest versions
task pre-commit:update
```

### Code Formatting
```bash
# Basic Go formatting (used by simple Git hook)
task fmt

# Advanced formatting with import organization
task fmt-imports
```

## Configuration Files

- **`.git/hooks/pre-commit`** - Simple Git hook script
- **`.pre-commit-config.yaml`** - Pre-commit framework configuration
- **`.secrets.baseline`** - Baseline for secret detection (detect-secrets)

## Hook Details

### Automatic Hooks (Run on every commit)
1. **Go formatting** - Ensures consistent code style
2. **Trailing whitespace removal** - Cleans up line endings
3. **End-of-file fixing** - Ensures files end with newlines
4. **YAML validation** - Checks YAML syntax
5. **Secret detection** - Prevents accidental secret commits
6. **Import organization** - Sorts and groups imports properly

### Manual Hooks (Run explicitly)
1. **golangci-lint** - Comprehensive Go linting (run with `pre-commit run golangci-lint`)
2. **Task lint** - Custom lint task (run with `pre-commit run task-lint`)

## Skipping Hooks

### Skip for a single commit:
```bash
git commit --no-verify -m "commit message"
```

### Skip specific hooks:
```bash
SKIP=go-vet-mod,go-unit-tests git commit -m "commit message"
```

### Skip all pre-commit framework hooks but keep simple Git hook:
```bash
pre-commit uninstall
# The simple Git hook will still run
```

## CI Integration

The pre-commit configuration includes CI settings for automatic updates and fixes when using pre-commit.ci service.

## Troubleshooting

### Pre-commit command not found
Enter the Nix development shell:
```bash
nix develop
# or if using direnv
direnv allow
```

### Hooks failing
Run hooks manually to see detailed output:
```bash
task pre-commit:run
```

### Reset hooks
Uninstall and reinstall:
```bash
pre-commit uninstall
task pre-commit:install
```

### Update hook versions
```bash
task pre-commit:update
```

## Customization

### Adding new hooks
Edit `.pre-commit-config.yaml` and add new hook configurations.

### Excluding files
Use the `exclude` pattern in individual hooks or add to `.pre-commit-config.yaml`.

### Adjusting secret detection
Update `.secrets.baseline` after reviewing detected secrets:
```bash
detect-secrets scan --update .secrets.baseline
```

## Best Practices

1. **Run locally first**: Test changes with `task pre-commit:run` before committing
2. **Keep baseline updated**: Regularly update the secrets baseline for legitimate secrets
3. **Review auto-fixes**: Always review automatically applied fixes before pushing
4. **Update regularly**: Run `task pre-commit:update` periodically to get latest hook versions
5. **Use appropriate skip**: Only skip hooks when absolutely necessary and understand the implications
