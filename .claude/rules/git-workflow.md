# Git Workflow Rules for argus

Follow these git conventions for this project.

## Commit Messages

- **Style**: conventional
- **Format**: `<type>(<scope>): <description>`
- **Types**: feat, chore, fix, docs, test
- **Example**: `feat(detector): add new feature`

### Commit Types

| Type | Description |
|------|-------------|
| feat | New feature |
| fix | Bug fix |
| docs | Documentation only |
| style | Formatting, no code change |
| refactor | Code restructuring |
| test | Adding/fixing tests |
| chore | Maintenance tasks |

## Branch Naming

- **Format**: `<prefix>/<description>`
- **Prefixes**: feat, chore, fix
- **Examples**:
  - `feat/user-auth`
  - `chore/login-bug`
  - `fix/update-deps`

## Workflow

1. Create a feature branch from main
2. Make small, focused commits
3. Push and open a pull request
4. Address review comments
5. Squash and merge when approved
