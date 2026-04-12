# Ork Enhancement Proposals

This directory contains proposals for enhancing the Ork infrastructure automation tool. Each proposal outlines a specific feature or improvement with implementation details, examples, and considerations.

## Proposals Overview

### Implemented

- **Simplified API** - ✅ **IMPLEMENTED** - Fluent builder pattern via `NodeInterface` (`NewNode`, `SetPort`, `SetUser`, `Connect`, `RunCommand`, etc.)

- **Connection Reuse** - ✅ **IMPLEMENTED** - Persistent SSH connections via `Node.Connect()` / `Node.Close()`

- **Testing Framework** - ✅ **PARTIALLY IMPLEMENTED** - Unit tests exist; mock helpers (`sshtest`, `playbooktest`) remain

### Not Implemented (Priority Order)

#### High Priority

1. **[Structured Logging](2026-04-12-structured-logging.md)** - Replace `log.Printf` with `slog`
   - Status: Not Implemented
   - Blocked by: None

2. **[Idempotency Framework](2026-04-12-idempotency-framework.md)** - Standard `Result` type with `Changed` status
   - Status: Not Implemented
   - Blocked by: None

3. **[CLI Tool](2026-04-12-cli-tool.md)** - Command-line interface
   - Status: Not Implemented
   - Blocked by: Configuration Management

4. **[Configuration Management](2026-04-12-configuration-management.md)** - File/env config loading
   - Status: Not Implemented
   - Blocked by: None
   - Needed for: CLI Tool

5. **[Dry-Run Mode](2026-04-12-dry-run-mode.md)** - Preview changes before applying
   - Status: Not Implemented
   - Blocked by: Idempotency Framework

#### Medium Priority

6. **[Parallel Execution](2026-04-12-parallel-execution.md)** - Multi-host execution
   - Status: Not Implemented
   - Blocked by: Connection Pool

7. **[Connection Pooling](2026-04-12-connection-pooling.md)** - Multi-host connection limiting
   - Status: Partially Implemented (reuse done, pool remains)
   - Blocked by: None

8. **[Playbook Dependencies](2026-04-12-playbook-dependencies.md)** - Auto-resolve dependencies
   - Status: Not Implemented
   - Blocked by: Idempotency Framework (for caching)

9. **[Rollback Support](2026-04-12-rollback-support.md)** - Undo changes on failure
   - Status: Not Implemented
   - Blocked by: Complex; low priority

## Implementation Roadmap

### Phase 1: Foundation (Complete)
- ✅ Simplified API - Fluent builder pattern implemented
- ✅ Connection Reuse - Persistent connections via Node
- ✅ Testing Framework - Basic tests in place

### Phase 2: Core Features (Next)
1. **Structured Logging** - slog integration
2. **Idempotency Framework** - Result type and Check interface
3. **Configuration Management** - File/env config for CLI

### Phase 3: CLI & Safety
4. **CLI Tool** - cobra-based binary
5. **Dry-Run Mode** - Preview changes

### Phase 4: Scale (Later)
6. **Connection Pool** - Multi-host resource management
7. **Parallel Execution** - Fleet operations
8. **Playbook Dependencies** - Auto-resolution

### Phase 5: Advanced
9. **Rollback Support** - Transaction management

## Summary

| Proposal | Status | Notes |
|----------|--------|-------|
| Connection Pooling | Partially Implemented | Reuse via Node done; true pool remains |
| Testing Framework | Partially Implemented | Tests exist; mock helpers needed |
| Idempotency Framework | Not Implemented | Needs Result type, CheckablePlaybook |
| Structured Logging | Not Implemented | Replace log.Printf with slog |
| Configuration Management | Not Implemented | Required for CLI |
| CLI Tool | Not Implemented | Blocked by config management |
| Dry-Run Mode | Not Implemented | Blocked by idempotency |
| Parallel Execution | Not Implemented | Blocked by connection pool |
| Playbook Dependencies | Not Implemented | Blocked by idempotency |
| Rollback Support | Not Implemented | Low priority |

## Contributing

To propose a new enhancement:

1. Copy the template below
2. Create a new file: `YYYY-MM-DD-feature-name.md`
3. Fill in all sections
4. Submit for review

### Proposal Template

```markdown
# Proposal: Feature Name

**Date:** YYYY-MM-DD
**Status:** Draft | In Progress | Implemented | Rejected
**Author:** Your Name

## Problem Statement
What problem does this solve?

## Proposed Solution
High-level approach

## Implementation
Detailed technical design with code examples

## Benefits
Why should we do this?

## Challenges & Solutions
What could go wrong and how to handle it

## Implementation Plan
Step-by-step phases

## Success Metrics
How to measure success

## Open Questions
Unresolved issues
```

## Status Definitions

- **Draft**: Initial proposal, open for discussion
- **In Progress**: Actively being implemented
- **Implemented**: Completed and merged
- **Rejected**: Decided not to pursue
- **Deferred**: Good idea but not now

## Priority Levels

- **High**: Critical for production use or major pain point
- **Medium**: Valuable improvement, not blocking
- **Low**: Nice to have, can wait

## Discussion

For questions or feedback on any proposal:
- Open an issue referencing the proposal
- Comment on the pull request
- Discuss in team meetings

## Related Resources

- [Project README](../../README.md)
- [Contributing Guidelines](../../CONTRIBUTING.md) (if exists)
- [Architecture Documentation](../architecture/) (if exists)
