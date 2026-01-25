---
sidebar_position: 4
title: Pattern Detection
description: How argus identifies code patterns
---

# Pattern Detection

argus identifies common code patterns across your codebase.

## Data Fetching

| Pattern | Detection |
|---------|-----------|
| React Query | `useQuery`, `useMutation` calls |
| SWR | `useSWR` calls |
| Axios | `axios` imports |
| Fetch | `fetch()` calls |

## State Management

| Pattern | Detection |
|---------|-----------|
| useState | React hook usage |
| Redux | `useSelector`, `useDispatch` |
| Zustand | `create` from zustand |
| MobX | `observer`, `observable` |

## API Patterns

| Pattern | Detection |
|---------|-----------|
| REST | HTTP method handlers |
| GraphQL | `gql` templates, resolvers |
| tRPC | `trpc` router definitions |
| gRPC | `.proto` files |

## Authentication

| Pattern | Detection |
|---------|-----------|
| JWT | `jwt` imports, token handling |
| OAuth | OAuth provider configs |
| Session | Session middleware |

## Database & ORM

| Pattern | Detection |
|---------|-----------|
| Prisma | `@prisma/client` |
| Drizzle | `drizzle-orm` |
| GORM | `gorm.io/gorm` |
| SQLAlchemy | `sqlalchemy` imports |

## Testing Patterns

| Pattern | Detection |
|---------|-----------|
| Unit tests | `test`, `it`, `describe` |
| Mocking | `mock`, `jest.mock` |
| Fixtures | `pytest.fixture` |
| Snapshots | `toMatchSnapshot` |

## React Patterns

| Pattern | Detection |
|---------|-----------|
| Hooks | `use*` function calls |
| Context | `createContext`, `useContext` |
| HOC | Component wrapping patterns |
| Render props | Function as children |

## Go Patterns

| Pattern | Detection |
|---------|-----------|
| Cobra CLI | `cobra.Command` |
| Goroutines | `go func` |
| Channels | `make(chan` |
| Context | `context.Context` |
| Error wrapping | `fmt.Errorf` with `%w` |

## Python Patterns

| Pattern | Detection |
|---------|-----------|
| Decorators | `@decorator` syntax |
| Type hints | Function annotations |
| Async/await | `async def`, `await` |
| Dataclasses | `@dataclass` |
| Pydantic | `BaseModel` inheritance |

## ML Patterns

argus detects machine learning patterns:

| Pattern | Detection |
|---------|-----------|
| PyTorch | `torch` imports |
| TensorFlow | `tensorflow` imports |
| scikit-learn | `sklearn` imports |
| Transformers | `transformers` imports |
