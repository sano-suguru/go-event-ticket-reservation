---
name: context7
description: Fetch official documentation for major libraries and frameworks. Use for library lookups, API references, how-to queries, and framework-specific questions
---

# Context7 Documentation Lookup

Retrieves official documentation for libraries and frameworks.

## Workflow

### 1. Resolve Library ID (Required)

```
mcp__context7__resolve-library-id
libraryName: "react" or "vue" or "express"
```

### 2. Fetch Documentation

```
mcp__context7__get-library-docs
context7CompatibleLibraryID: "/vuejs/core"
topic: "composition-api"     # Optional: specific topic
mode: "code"                 # "code" or "info"
page: 1                      # Paginate if needed
```

## Mode Selection

**code (default):**
API references, code examples, implementation patterns

**info:**
Conceptual guides, architecture explanations

## Tips

- Always resolve library ID first (don't guess ID format)
- Use `page: 2, 3...` if initial results insufficient
- Try alternative library names if no results (e.g., "vue router" vs "vue-router")
- Version-specific IDs: `/org/project/v3`

## When to Use

- Library API specification lookup
- Official best practices
- Framework-specific implementation patterns
- Code examples and samples

## Examples

**User**: How to use React Hooks?

**Agent**: Fetches React official docs and explains Hooks patterns

---

**User**: Next.js 14 App Router usage?

**Agent**: Retrieves Next.js v14 docs and explains App Router