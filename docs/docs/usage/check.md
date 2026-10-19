---
title: Check Command
description: Fail CI when generated files are out of date
---

# Check Command

`argus check` renders the same files as `argus scan` and compares them with the files on disk. Nothing is written.

```bash
argus check .
```

Each file is listed as `up to date`, `out of date`, `missing` or `skipped` (files argus leaves alone, such as an `AGENTS.md` without markers). Text outside the `ARGUS:AUTO` block never counts as drift.

| Exit status | Meaning |
|-------------|---------|
| `0` | Every file is up to date |
| `1` | `argus scan` would change a file |
| `2` | Error, for example a shallow clone or `ai` enabled in `.argus.yaml` |

## In CI

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - run: curl -fsSL https://raw.githubusercontent.com/Priyans-hu/argus/main/install.sh | bash
      - run: argus check
```

## Limits

- Commit conventions come from git history, so check needs a full clone (`fetch-depth: 0`).
- The title comes from the checkout directory name, so a clone under another name shows drift.
- `--ai` and `--full` output cannot be checked; there is no config key for `--full` yet.
