# Architecture Rules for argus

This project follows a **Standard Go Layout** architecture.

## Entry Point

Main entry: `cmd/argus/main.go`

## Layer Structure

Follow the dependency rules between layers:

### cmd

Entry points / CLI

**Packages**: `argus`

**Can depend on**: `analyzer`, `config`, `generator`, `merger`

### internal

Private packages

**Packages**: `analyzer`, `config`, `detector`, `generator`, `merger`

**Can depend on**: `config`, `detector`

### pkg

Public packages

**Packages**: `types`

## Key Directories

- `internal/config/` - Configuration
- `internal/detector/` - Detection logic
- `internal/generator/` - Code generation
- `internal/merger/` - Merge utilities
- `pkg/types/` - Type definitions
- `cmd/` - Command entrypoints
- `internal/analyzer/` - Analysis logic

## Diagram

```
```
                    ┌─────────────┐
                    │    argus    │
                    └──────┬──────┘
                           │
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  analyzer   │  │   config    │  │  detector   │
└─────────────┘  └─────────────┘  └─────────────┘
       │              │
┌─────────────┐  ┌─────────────┐
│  generator  │  │   merger    │
└─────────────┘  └─────────────┘
                           │
                           ▼
                   [External Services]
```

```

## Guidelines

- Respect layer boundaries - lower layers should not depend on higher layers
- Keep business logic in the appropriate layer
- Use dependency injection for cross-layer dependencies
- New features should follow the existing architectural patterns
