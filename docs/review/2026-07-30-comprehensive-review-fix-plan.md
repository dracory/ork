# Comprehensive Review Fix Plan: Concurrent Safe Skill Execution

This document analyzes the critical issue of shared, mutable skill instances during concurrent `Inventory` or `Group` execution in `ork` and proposes multiple potential solutions (including breaking changes). It also highlights existing patterns in the codebase that can be reused to solve this cleanly.

---

## The Issue

In `ork`, `Inventory.Run()`, `RunByID()`, and `Check()` handle concurrency by running each target node in its own goroutine (`inventory_implementation.go`). However, they call `skill.SetNodeConfig(...)`, `skill.SetArgs(...)`, and `skill.Run()` on **the exact same `RunnableInterface` instance** concurrently under the following typical conditions:

1. The caller passes a single skill instance to `inventory.Run(skill)`.
2. `RunByID(id)` is used, which retrieves the skill pointer from the global singleton registry (`registry.FindByID(id)`).

Because `types.BaseSkill` lacks synchronization and mutates state directly (e.g., updating the raw `args` map or `nodeCfg` struct), multiple goroutines concurrently executing will:
- **Race on fields**, leading to one node's commands/arguments leaking into another node's environment.
- **Trigger fatal, unrecoverable runtime errors** in Go (e.g., concurrent map read/write or concurrent map writes) when calling `SetArg`/`GetArg`, which instantly crashes the process.

---

## Existing Codebase Patterns & Capabilities to Reuse

When looking at how `ork` is currently designed, there is already a strong, established design pattern of **defensive copying** to prevent concurrent modifications and state leaks. We can reuse this pattern's design philosophy to solve the skill concurrency issue.

Existing examples in the codebase include:
1. **`Node.GetNodeConfig()`**: Returns a deep copy of the underlying `types.NodeConfig` (including a freshly copied `Args` map) so external modifications do not leak back into the node.
2. **`NewNodeFromConfig()`**: Deep-copies the provided `types.NodeConfig` immediately to ensure isolation.
3. **`Group.GetNodes()` / `Inventory.GetNodes()`**: Returns a copied slice of `NodeInterface` to prevent external modification of the internal slice.
4. **`Group.GetArgs()`**: Returns a copy of the arguments map.

By leveraging a similar **defensive copying (cloning) pattern** for skills, we naturally align with `ork`'s existing architectural conventions and design principles.

---

## Proposed Solutions

We have researched and formulated six distinct solutions to resolve this issue. They range from non-breaking/backward-compatible changes to highly robust breaking API redesigns.

---

### Solution 1: Add Explicit Clone Support to `RunnableInterface` (Breaking Change - Recommended)

Add a `Clone() RunnableInterface` method to the `RunnableInterface` in `types/runnable_interface.go`. Every skill (both built-in and user-defined) must implement this method.

#### Details

1. Define `Clone() RunnableInterface` inside `types.RunnableInterface`.
2. Implement deep cloning inside `types.BaseSkill`:
   ```go
   func (b *BaseSkill) Clone() *BaseSkill {
       if b == nil {
           return nil
       }
       clone := &BaseSkill{
           BaseBecome:  b.BaseBecome, // Struct copy (safe as it contains only a string)
           id:          b.id,
           description: b.description,
           nodeCfg:     b.nodeCfg,    // Struct copy (deep copy of Args is handled during node config copy)
           dryRun:      b.dryRun,
           timeout:     b.timeout,
       }
       if b.args != nil {
           clone.args = make(map[string]string, len(b.args))
           for k, v := range b.args {
               clone.args[k] = v
           }
       }
       return clone
   }
   ```
3. Update all 45+ built-in skills to implement `Clone()`. For example, in `skills/ping/ping.go`:
   ```go
   func (p *Ping) Clone() types.RunnableInterface {
       return &Ping{
           BaseSkill: p.BaseSkill.Clone(),
       }
   }
   ```
4. Within `Inventory` and `Node` runner implementations, clone the skill immediately prior to mutating or running it per-node. For example, in `inventory_implementation.go`:
   ```go
   clonedSkill := skill.Clone()
   nodeResults := n.Run(clonedSkill)
   ```

#### Pros & Cons

* **Pros:**
  - **Type-safe & Idiomatic:** The Go compiler guarantees that any skill registered or run implements `Clone()`.
  - **Completely Isolated:** No concurrent writes, read-after-write anomalies, or sharing of any state.
  - **Clear Contract:** Extensible for custom user-created skills.
  - **Reuses Existing Design Philosophy:** Leverages defensive copying, matching current patterns in `GetNodeConfig` and `GetNodes`.
* **Cons:**
  - **Breaking Change:** Any external/user-defined skill that implements `RunnableInterface` will fail to compile until they implement `Clone()`.
  - **Boilerplate:** Requires adding a simple boilerplate `Clone` method to every single built-in skill struct.

---

### Solution 2: Registry-Level Factory/Constructors & Clone on Run (Non-breaking Change)

Avoid breaking the `RunnableInterface` contract by adding an optional/fallback interface for cloning, while refactoring the global registry to store factory functions instead of pre-instantiated singleton pointers.

#### Details

1. Define an optional `Clonable` interface inside `types`:
   ```go
   type Clonable interface {
       Clone() RunnableInterface
   }
   ```
2. Implement `Clone() *BaseSkill` on `BaseSkill` as described in Solution 1.
3. In `inventory_implementation.go` and `node_implementation.go`, check if the skill implements `Clonable`. If so, clone it.
4. If the skill does not implement `Clonable` (e.g. legacy/third-party skills), use reflection to perform a shallow clone of the struct, copying its fields, and manually deep-copying `BaseSkill` if found.
5. In `registry.go`, keep storing instances but provide a mechanism to return a fresh copy (e.g. if the registered skill implements `Clonable`, return `skill.Clone()`).

#### Pros & Cons

* **Pros:**
  - **Non-breaking Change:** Existing third-party skills will continue to compile and run without modification.
  - Resolves concurrency issues seamlessly for both registry-sourced skills and directly passed skills.
* **Cons:**
  - **Complex Fallback Logic:** Relies on interface type assertions and reflection-based copies if a third-party skill doesn't implement `Clonable`.
  - Compile-time safety is lost for custom third-party skills that are not manually updated with `Clone()`.

---

### Solution 3: Dynamic Deep Copy via Reflection (Non-breaking Change)

Use Go's `reflect` package inside `inventory_implementation.go` and `node_implementation.go` to dynamically deep-copy any skill passed to `Run` or `Check` without introducing new interface methods or modifying any skill implementation.

#### Details

Write a robust cloning function using reflection:
```go
func cloneSkill(skill types.RunnableInterface) types.RunnableInterface {
    val := reflect.ValueOf(skill)
    if val.Kind() != reflect.Ptr || val.IsNil() {
        return skill
    }

    // Allocate new struct of the same type
    elem := val.Elem()
    newVal := reflect.New(elem.Type())
    newElem := newVal.Elem()

    // Copy exported fields.
    // Deep-copy any embedded BaseSkill and map fields manually
    // ...
    return newVal.Interface().(types.RunnableInterface)
}
```

#### Pros & Cons

* **Pros:**
  - **Zero Breaking Changes:** No modifications to `RunnableInterface` or any of the built-in/custom skills.
  - Solves the concurrency problem immediately and transparently for the user.
* **Cons:**
  - **Reflection Overhead:** Performance is slightly lower (though negligible for network-bound SSH operations).
  - **Fragile & Complex:** Deep-copying nested structures, map values, and embedded pointer fields correctly via reflection in Go is notoriously difficult and prone to edge-case bugs (especially with unexported fields or customized wrapper fields).

---

### Solution 4: Immutable/Stateless Skills API (Extremely Breaking Change)

Completely redesign the `RunnableInterface` and skill lifecycle. Instead of configuring the skill with `SetNodeConfig(...)`, `SetArgs(...)`, and `SetDryRun(...)` before execution, pass these values directly as immutable arguments to the `Run` and `Check` methods.

#### Details

1. Remove setter/getter state methods from `RunnableInterface`.
2. Change the signature of `Check` and `Run` to:
   ```go
   type RunnableInterface interface {
       GetID() string
       GetDescription() string
       Check(cfg NodeConfig, args map[string]string) (bool, error)
       Run(cfg NodeConfig, args map[string]string, dryRun bool) Result
   }
   ```
3. In `inventory_implementation.go`, the runner simply passes the node-specific config and consolidated arguments directly:
   ```go
   nodeResults := n.Run(skill, nodeArgs)
   ```

#### Pros & Cons

* **Pros:**
  - **Flawless Concurrency Design:** By making skills completely stateless, there is absolutely zero risk of concurrent access issues, race conditions, or unrecoverable map panics.
  - Simplifies the core of `types.BaseSkill` enormously.
* **Cons:**
  - **Extremely Disruptive breaking change:** Every single line of code, test, skill, playbook, and user script in the `ork` ecosystem would be completely broken and require rewriting.
  - Destroys the existing fluent/chaining API style that is heavily documented and advertised as a major feature of `ork`.

---

### Solution 5: Synchronization / Mutex-Protected BaseSkill (Incomplete Solution)

Protect state modification inside `BaseSkill` using a `sync.RWMutex` to avoid Go runtime map crashes.

#### Details

1. Add a `sync.RWMutex` field to `types.BaseSkill`.
2. Wrap every getter/setter (e.g., `GetArg`, `SetArg`, `SetNodeConfig`, `IsDryRun`) in `RWMutex` locks.

#### Pros & Cons

* **Pros:**
  - **Non-breaking Change:** Extremely simple to implement with minimal changes.
  - Prevents the unrecoverable Go runtime panic (fatal concurrent map write).
* **Cons:**
  - **Does Not Prevent Logic/State Races:** While it prevents the application from crashing, it does *not* prevent different execution goroutines from overwriting each other's configurations or arguments before they call `Run()`. For example, Node A's args could be updated to Node B's values right before Node A's runner executes, leading to silent, incorrect executions. Thus, this is not a viable stand-alone solution.

---

### Solution 6: Context-Bound Configuration Storage (Modern Go Approach - Breaking Change)

Pass execution context (such as the target node configuration and parameters) using Go's `context.Context` instead of storing them as fields on the skill.

#### Details

1. Keep getters and setters on `BaseSkill` for default properties (id, description), but remove `SetNodeConfig`, `SetArgs`, and `SetDryRun`.
2. Change execution signatures to accept a context:
   ```go
   type RunnableInterface interface {
       GetID() string
       GetDescription() string
       Check(ctx context.Context) (bool, error)
       Run(ctx context.Context) Result
   }
   ```
3. Store the active `NodeConfig` inside the `context.Context` under a custom type key prior to calling `Run(ctx)`.
4. In the skill's `Run` implementation, retrieve the config from the context:
   ```go
   cfg, ok := types.NodeConfigFromContext(ctx)
   ```

#### Pros & Cons

* **Pros:**
  - **Idiomatic Go:** Passing request/execution scoped parameters via context is standard pattern in modern Go libraries.
  - Fully concurrent-safe since the context is request-scoped and immutable across goroutine branches.
* **Cons:**
  - **Breaking Change:** Requires modifying the method signature of `Run()` and `Check()` on all skills.
  - Sightly less explicit than direct parameter passing (Solution 4).

---

## Recommendation & Action Plan

We highly recommend **Solution 1 (explicit `Clone` in `RunnableInterface`)**. Although it is technically a breaking change for custom skills, it provides **absolute compile-time correctness**, ensures Go idiomatic design, and completely avoids complex/fragile reflection hacks. Crucially, **it reuses the established design pattern of defensive copying already used extensively for `NodeConfig`, `Group` and `Inventory` structures in `ork`.** Since `ork` is still under active development, implementing this now prevents legacy design debt.

### Next Steps for Implementation:
1. Update `types/runnable_interface.go` to include `Clone() RunnableInterface`.
2. Update `types/base_skill.go` to implement `Clone() *BaseSkill` and `Clone() RunnableInterface`.
3. Generate/add the standard `Clone() types.RunnableInterface` boilerplate to all built-in skills in the `skills/` directory.
4. Modify `Inventory` and `Node` run/check executors to clone the skill prior to setting node config/args.
5. Run full test suites (`go test ./...`) to ensure compilation and logical correctness.
