# Ork Enhancement Proposals

This directory contains proposals for enhancing the Ork infrastructure automation tool. Each proposal outlines a specific feature or improvement with implementation details, examples, and considerations.

## Remaining Work (Priority Order)

#### High Priority

1. **[Structured Logging](2026-04-12-structured-logging.md)** - Replace `log.Printf` with `slog`
   - Status: Not Implemented
   - Blocked by: None

2. **[Configuration Management](2026-04-12-configuration-management.md)** - File/env config loading
   - Status: Not Implemented
   - Blocked by: None
   - Needed for: CLI Tool

3. **[CLI Tool](2026-04-12-cli-tool.md)** - Command-line interface
   - Status: Not Implemented
   - Blocked by: Configuration Management

4. **[Dry-Run Mode](2026-04-12-dry-run-mode.md)** - Preview changes before applying
   - Status: Not Implemented
   - Unblocked: Idempotency Framework now implemented

#### Medium Priority

5. **[Connection Pooling](2026-04-12-connection-pooling.md)** - Multi-host connection limiting
   - Status: Partially Implemented (reuse done, pool remains)
   - Blocked by: None

6. **[Parallel Execution](2026-04-12-parallel-execution.md)** - Multi-host execution
   - Status: Not Implemented
   - Blocked by: Connection Pool

7. **[Playbook Dependencies](2026-04-12-playbook-dependencies.md)** - Auto-resolve dependencies
   - Status: Not Implemented
   - Unblocked: Idempotency Framework now implemented

8. **[Rollback Support](2026-04-12-rollback-support.md)** - Undo changes on failure
   - Status: Not Implemented
   - Blocked by: Complex; low priority

## Implementation Roadmap

### Phase 1: Core Features
1. **Structured Logging** - slog integration
2. **Configuration Management** - File/env config for CLI

### Phase 2: CLI & Safety
3. **CLI Tool** - cobra-based binary (blocked by Configuration Management)
4. **Dry-Run Mode** - Preview changes

### Phase 3: Scale
5. **Connection Pool** - Multi-host resource management
6. **Parallel Execution** - Fleet operations (blocked by Connection Pool)
7. **Playbook Dependencies** - Auto-resolution

### Phase 4: Advanced
8. **Rollback Support** - Transaction management

## Summary

| Proposal | Status | Blocked By |
|----------|--------|------------|
| Structured Logging | Not Implemented | - |
| Configuration Management | Not Implemented | - |
| CLI Tool | Not Implemented | Configuration Management |
| Dry-Run Mode | Not Implemented | - |
| Connection Pooling | Partially Implemented | - |
| Parallel Execution | Not Implemented | Connection Pooling |
| Playbook Dependencies | Not Implemented | - |
| Rollback Support | Not Implemented | - |

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
