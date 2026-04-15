---
path: modules/config.md
page-type: module
summary: Configuration types for SSH-based automation, including NodeConfig with connection settings.
tags: [module, config, configuration]
created: 2025-04-14
updated: 2026-04-15
version: 2.0.0
---

## Changelog
- **v2.0.0** (2026-04-15): config package moved to types package - this page now redirects to types.md
- **v1.0.0** (2025-04-14): Initial creation

# config Package (Deprecated)

> **⚠️ Deprecated**: The `config` package has been moved to the `types` package as of v2.0.0. Please update your imports from `github.com/dracory/ork/config` to `github.com/dracory/ork/types`.

## Migration Guide

### Update Imports

**Before:**
```go
import "github.com/dracory/ork/config"

cfg := config.NodeConfig{...}
```

**After:**
```go
import "github.com/dracory/ork/types"

cfg := types.NodeConfig{...}
```

### What Changed

- `config.NodeConfig` → `types.NodeConfig`
- All NodeConfig methods remain the same
- Package location changed from `config/` to `types/`

## See Also

- [types](types.md) - Configuration types in the types package
- [ork](ork.md) - Uses NodeConfig for node configuration
- [ssh](ssh.md) - Uses NodeConfig for connections
- [Configuration](../configuration.md) - Configuration guide
