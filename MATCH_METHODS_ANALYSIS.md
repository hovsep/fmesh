# `*Match` Methods Analysis

## Summary

Analysis of all `*Match` methods in the F-Mesh codebase, identifying unused methods that can be dropped.

## 📊 Current State

### Method Distribution Across Types

| Type | AllMatch | AnyMatch | NoneMatch | CountMatch | FirstMatch | MatchesAll/Any |
|------|----------|----------|-----------|------------|------------|----------------|
| `component.Collection` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `component.ActivationResultCollection` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `port.Collection` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `port.Group` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `signal.Group` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `cycle.Group` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `labels.Collection` | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ (both) |

**Total Methods**: 44 methods across 7 types

---

## 🔍 Usage Analysis

### 1. **AllMatch** (7 methods)
**Usage**: ✅ **INTERNAL** - Used by `AllHaveSignals()`
- **File**: `port/collection.go`
- **Code**:
  ```go
  func (c *Collection) AllHaveSignals() bool {
      return c.AllMatch(func(p *Port) bool {
          return p.HasSignals()
      })
  }
  ```
- **External Usage**: 0
- **Internal Usage**: 1
- **Recommendation**: ✅ **KEEP** - Internal dependency

---

### 2. **AnyMatch** (7 methods)
**Usage**: ✅ **INTERNAL** - Used by `NoneMatch()` and `AnyHasSignals()`

#### Internal Dependencies:
1. **NoneMatch implementations** (5 files):
   - `component/collection.go`: `return !c.AnyMatch(predicate)`
   - `component/activation_result_collection.go`: `return !c.AnyMatch(predicate)`
   - `port/collection.go`: `return !c.AnyMatch(predicate)`
   - `port/group.go`: `return !g.AnyMatch(predicate)`
   - `signal/group.go`: `return !g.AnyMatch(predicate)`
   - `cycle/group.go`: `return !g.AnyMatch(predicate)`
   - `labels/labels.go`: `return !c.AnyMatch(pred)`

2. **AnyHasSignals** (1 file):
   - `port/collection.go`:
     ```go
     func (c *Collection) AnyHasSignals() bool {
         return c.AnyMatch(func(p *Port) bool {
             return p.HasSignals()
         })
     }
     ```

- **External Usage**: 0
- **Internal Usage**: 8 places
- **Recommendation**: ✅ **KEEP** - Critical internal dependency

---

### 3. **NoneMatch** (7 methods)
**Usage**: ❌ **UNUSED EXTERNALLY**
- **External Usage**: 0 (only in unit tests)
- **Internal Usage**: Implemented as `!AnyMatch()` in all cases
- **Files**:
  - `component/collection.go`
  - `component/activation_result_collection.go`
  - `port/collection.go`
  - `port/group.go`
  - `signal/group.go`
  - `cycle/group.go`
  - `labels/labels.go`
- **Recommendation**: ❌ **DROP** - Never used externally, trivial to inline

---

### 4. **CountMatch** (7 methods)
**Usage**: ⚠️ **INTEGRATION TEST ONLY**
- **External Usage**: 1 (integration test)
- **Integration Test Usage**:
  ```go
  // integration_tests/ports/port_creation_test.go
  portsWithSignals := inputs.CountMatch(func(p *port.Port) bool {
      return p.HasSignals()
  })
  ```
- **Documented Usage**: Comment in `port/collection.go`
- **Recommendation**: ✅ **KEEP** - Used in integration test, documented pattern

---

### 5. **FirstMatch** (3 methods)
**Usage**: ❌ **UNUSED**
- **External Usage**: 0 (only in `cycle/group_test.go`)
- **Types**:
  - `signal.Group.FirstMatch()`
  - `port.Group.FirstMatch()`
  - `cycle.Group.FirstMatch()`
- **Recommendation**: ❌ **DROP** - Never used in production or integration tests

---

### 6. **MatchesAll** (1 method)
**Usage**: ❌ **UNUSED**
- **External Usage**: 0 (only in `labels/labels_test.go`)
- **Type**: `labels.Collection.MatchesAll()`
- **Recommendation**: ❌ **DROP** - Never used outside unit tests

---

### 7. **MatchesAny** (1 method)
**Usage**: ❌ **UNUSED**
- **External Usage**: 0 (only in `labels/labels_test.go`)
- **Type**: `labels.Collection.MatchesAny()`
- **Recommendation**: ❌ **DROP** - Never used outside unit tests

---

## 📋 Recommendations Summary

### ✅ **KEEP (2 method types = 14 methods)**

| Method | Count | Reason |
|--------|-------|--------|
| `AllMatch` | 7 | Used internally by `AllHaveSignals()` |
| `AnyMatch` | 7 | Used internally by `NoneMatch()` and `AnyHasSignals()` |
| **TOTAL** | **14** | **Core predicate operations** |

---

### ⚠️ **KEEP FOR NOW (1 method type = 7 methods)**

| Method | Count | Reason |
|--------|-------|--------|
| `CountMatch` | 7 | Used in 1 integration test, documented pattern |
| **TOTAL** | **7** | **Low usage but valid** |

---

### ❌ **DROP (4 method types = 23 methods)**

| Method | Count | Reason |
|--------|-------|--------|
| `NoneMatch` | 7 | Never used externally, trivial (`!AnyMatch()`) |
| `FirstMatch` | 3 | Never used at all (test-only) |
| `MatchesAll` | 1 | Never used (labels only) |
| `MatchesAny` | 1 | Never used (labels only) |
| **TOTAL** | **12** | **Dead code** |

---

## 🎯 Proposed Action: Remove 12 Methods

### **Conservative Approach** (Recommended)
Remove **12 unused methods**:
- **7** `NoneMatch` methods (all types)
- **3** `FirstMatch` methods (signal.Group, port.Group, cycle.Group)
- **1** `MatchesAll` method (labels.Collection)
- **1** `MatchesAny` method (labels.Collection)

**Impact**:
- **API Reduction**: 27% (44 → 32 methods)
- **Risk**: Very low - all unused
- **Breaking Changes**: Only unit tests
- **Lines Saved**: ~150-180 lines

---

### Alternative: Keep CountMatch Under Review
If you're concerned about `CountMatch` having only 1 usage:
- Keep for now (documented, in integration test)
- Monitor usage over next releases
- Consider deprecation if usage doesn't grow

---

## 💡 Key Insights

### Pattern Analysis

**Useful Match methods**:
- ✅ `AllMatch` - Boolean aggregation (all satisfy predicate)
- ✅ `AnyMatch` - Boolean existence (any satisfy predicate)
- ✅ `CountMatch` - Counting (how many satisfy predicate)

**Not useful**:
- ❌ `NoneMatch` - Trivial negation of `AnyMatch`
- ❌ `FirstMatch` - Duplicates `Filter().First()` pattern
- ❌ `MatchesAll/Any` - Special-case label matching (unused)

### The Real Pattern

Users need:
1. **Boolean checks**: `AllMatch`, `AnyMatch` ✅
2. **Counting**: `CountMatch` ✅
3. **Filtering**: `Filter()` (already exists) ✅

Users DON'T need:
1. Negations - they can use `!AnyMatch()` directly
2. `FirstMatch` - they can use `Filter().First()`
3. Special-case matchers - `AllMatch` with predicate is enough

---

## 🚀 Implementation Plan

### Phase 1: Remove NoneMatch (7 methods)

**Files to modify**:
1. `component/collection.go` - Remove `NoneMatch()`
2. `component/activation_result_collection.go` - Remove `NoneMatch()`
3. `port/collection.go` - Remove `NoneMatch()`
4. `port/group.go` - Remove `NoneMatch()`
5. `signal/group.go` - Remove `NoneMatch()`
6. `cycle/group.go` - Remove `NoneMatch()`
7. `labels/labels.go` - Remove `NoneMatch()`

**Tests to update**:
- `component/activation_result_collection_test.go` - Remove `TestActivationResultCollection_NoneMatch`
- `cycle/group_test.go` - Remove `TestGroup_NoneMatch`

---

### Phase 2: Remove FirstMatch (3 methods)

**Files to modify**:
1. `signal/group.go` - Remove `FirstMatch()`
2. `port/group.go` - Remove `FirstMatch()`
3. `cycle/group.go` - Remove `FirstMatch()`

**Tests to update**:
- `cycle/group_test.go` - Remove `TestGroup_FirstMatch`

---

### Phase 3: Remove MatchesAll/MatchesAny (2 methods)

**Files to modify**:
1. `labels/labels.go` - Remove `MatchesAll()` and `MatchesAny()`

**Tests to update**:
- `labels/labels_test.go` - Remove `TestLabelsCollection_MatchesAll` and `TestLabelsCollection_MatchesAny`

---

## 📊 Expected Results

### Before
- **Total Match methods**: 44
- **AllMatch**: 7
- **AnyMatch**: 7
- **NoneMatch**: 7
- **CountMatch**: 7
- **FirstMatch**: 3
- **MatchesAll/Any**: 2

### After
- **Total Match methods**: 32 (-27%)
- **AllMatch**: 7 (keep)
- **AnyMatch**: 7 (keep)
- **CountMatch**: 7 (keep)
- ~~**NoneMatch**: 0~~ (dropped)
- ~~**FirstMatch**: 0~~ (dropped)
- ~~**MatchesAll/Any**: 0~~ (dropped)

### Code Impact
- **Methods removed**: 12
- **Lines saved**: ~150-180
- **Tests removed/updated**: ~100 lines
- **Risk**: Very low

---

## ✅ Final Recommendation

**Remove 12 unused Match methods immediately**:
- 7 `NoneMatch` - trivial negation
- 3 `FirstMatch` - redundant
- 2 `MatchesAll/Any` - unused special cases

This will:
- Simplify the API by 27%
- Remove dead code
- Keep all actually useful methods
- No production code impact

