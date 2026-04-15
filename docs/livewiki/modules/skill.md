---
path: modules/skill.md
page-type: module
summary: BasePlaybook implementation and utility functions for automation tasks.
tags: [module, skill, baseplaybook]
created: 2025-04-14
updated: 2026-04-15
version: 2.0.0
---

## Changelog
- **v2.0.0** (2026-04-15): playbook package moved to types package - this page now redirects to types.md
- **v1.1.0** (2026-04-14): Updated BasePlaybook documentation
- **v1.0.0** (2025-04-14): Initial creation

# playbook Package (Deprecated)

> **⚠️ Deprecated**: The `playbook` package has been moved to the `types` package as of v2.0.0. Please update your imports from `github.com/dracory/ork/playbook` to `github.com/dracory/ork/types`.

## Migration Guide

### Update Imports

**Before:**
```go
import "github.com/dracory/ork/playbook"

pb := playbook.NewBasePlaybook()
```

**After:**
```go
import "github.com/dracory/ork/types"

pb := types.NewBasePlaybook()
```

### What Changed

- `playbook.BasePlaybook` → `types.BasePlaybook`
- `playbook.NewBasePlaybook()` → `types.NewBasePlaybook()`
- `playbook.BaseSkill` → `types.BaseSkill`
- `playbook.NewBaseSkill()` → `types.NewBaseSkill()`
- All functionality remains the same
- Package location changed from `playbook/` to `types/`

## See Also

- [types](types.md) - BasePlaybook, BaseSkill, and RunnableInterface in the types package
- [skills](skills.md) - Built-in skill implementations
- [ork](ork.md) - Uses types package
