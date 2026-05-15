# Bug Fixing Skill - Professional Debugging Guide

This skill provides systematic approaches to identify and fix two main types of bugs:
- **Syntax Bugs**: Code errors preventing compilation/execution
- **Logic Bugs**: Errors in program behavior despite valid syntax

---

## 🎯 Core Principles

### 1. Always Start with Reproduction
- Understand the bug first before fixing it
- Create minimal reproduction steps
- Never fix without verifying the issue exists

### 2. Type Safety First
- TypeScript is mandatory - no `any` unless absolutely necessary
- Define proper interfaces/types for all data structures
- Use strict mode settings when possible

### 3. Readability Over Cleverness
- Clear, self-explanatory code > clever one-liners
- Single responsibility per function (keep <10 lines ideal)
- Avoid deep nesting (>3 levels is too much)

---

## 🔍 Syntax Bug Detection & Fixing

### Step-by-Step Process:

#### 1. **Identify Error Location**
```bash
# Run linter/type checker first
npm run lint          # ESLint errors
npx tsc --noEmit      # TypeScript compilation errors
npm test              # Test failures with stack traces
```

#### 2. **Common Syntax Issues to Check:**

| Issue | How to Spot | Fix Pattern |
|-------|-------------|-------------|
| Missing/extra brackets `{}` | Linter error, unmatched braces | Add missing or remove extra |
| Semicolon errors | Run-time `Unexpected token` | Ensure proper statement termination |
| Import/export mismatch | Module not found errors | Check import vs export consistency |
| Bracket mismatches in arrays | Index out of bounds | Verify array structure matches logic |

#### 3. **Fix Checklist:**
- [ ] All brackets `{}` balanced and properly nested
- [ ] String quotes matching (`'` or `"`)
- [ ] Function calls with correct parentheses `()`
- [ ] Proper variable declarations (let/const vs var)
- [ ] Correct import/export statements
- [ ] No trailing commas in JavaScript objects

#### 4. **Immediate Actions:**
```bash
# After fix, verify syntax is valid
npx tsc --noEmit && echo "✓ Syntax OK" || exit 1
npm run lint | grep -v "warning" > /dev/null && echo "✓ Lint OK"
```

---

## 🧠 Logic Bug Detection & Fixing

### Step-by-Step Process:

#### 1. **Understand Expected vs Actual Behavior**
```typescript
// BAD: Don't guess, document the bug first
function calculateTotal(items) { /* ... */ } // What breaks?

// GOOD: Document the issue clearly
/**
 * BUG: Returns incorrect total when items array has nested arrays
 * INPUT: [{price: 10}, [5]]
 * EXPECTED: 15
 * ACTUAL: NaN (due to non-numeric price)
 */
```

#### 2. **Add Strategic Logging**
Use meaningful logs ONLY for debugging, remove after fix:

```typescript
// DEBUG MODE - Add these temporarily
const debug = process.env.NODE_ENV === 'development' ? console.log : () => {}

debug('Input received:', input);          // Check what comes in
debug('Processing step 1/3...');          // Track execution flow  
debug('Intermediate result:', intermediate, expected: value);    // Validate each step
```

#### 3. **Systematic Debugging Approach:**

**A. Isolate the Problem Area**
- Use binary search approach on code sections
- Comment out parts to narrow down location
- Add boundary tests (empty input, null, max values)

**B. Check Common Logic Pitfalls:**

| Scenario | Bug Pattern | Fix Strategy |
|----------|-------------|--------------|
| Off-by-one errors | `for` loop bounds wrong | Verify: should it be `<` or `<=`? |
| Async race conditions | Unordered results | Use Promise.all() or proper sequencing |
| Infinite loops | Missing exit condition | Add debug counter/logic check |
| Wrong comparison operators | `=` instead of `==`/`===` | Always use strict equality `===` |
| Type coercion issues | `"5" + 3 = "53"` vs `"5"-3=2` | Ensure consistent types before operations |

**C. Add Validation Guards:**
```typescript
// GOOD: Validate inputs early, fail fast with clear messages
function processItems(items) {
  if (!Array.isArray(items)) {
    throw new TypeError('Expected array, got ' + typeof items);
  }
  
  // Process...
}

// BAD: Silent failures that cause logic bugs later
function processItems(items) {
  const result = [];  // What if items is undefined?
  for (item of items) { /* ... */ } // Crashes silently!
```

#### 4. **Unit Testing Strategy**

Create tests BEFORE fixing to verify behavior:

```typescript
// Test file pattern - covers edge cases
describe('functionName', () => {
  test('handles normal case correctly', () => {});
  
  test('returns empty array for null input', () => {});
  
  test('throws error when required field missing', () => {});
  
  test('processes large datasets without timeout', () => {});
  
  test('works with edge values (0, negative, max)', () => {});
});

// Run targeted tests
npm test -- testName=functionName.test.ts
```

#### 5. **Fix Verification Checklist:**
- [ ] Bug is actually fixed (test passes)
- [ ] No new bugs introduced (run full test suite)
- [ ] Edge cases handled properly
- [ ] Error messages are clear and helpful
- [ ] Code follows project patterns
- [ ] Types are correct everywhere

---

## 🛠️ Practical Debugging Tools & Commands

### Essential CLI Commands:
```bash
# 1. Find all TypeScript errors
npx tsc --noEmit

# 2. Run linter with specific file focus
npm run lint src/**/*.ts | grep -v "warning" || true

# 3. Get detailed error context
npm test -- --verbose testName.test.ts

# 4. Watch mode for development (if applicable)
watchmedo auto-restart --pattern="*.js,*.ts" --signal SIGUSR1 my-app.js
```

### VS Code Shortcuts:
- `F5` - Debug current file
- `Ctrl+Shift+F5` - Stop debugging
- `Ctrl+Alt+F5` - Step over (skip function call)
- `Ctrl+Alt+B` - Step into (enter function)
- `Ctrl+F10` - Continue execution

### Chrome DevTools:
- **Sources panel**: Set breakpoints, inspect variables
- **Console**: Type `$0` for last error message
- **Performance tab**: Profile slow functions

---

## 📋 Pre-Fix Checklist (Run Before Making Changes)

```markdown
## Bug Investigation Complete? ✓
- [ ] I can reproduce the bug consistently
- [ ] I know what input causes it
- [ ] I know expected vs actual behavior
- [ ] I've identified where in codebase the issue is
- [ ] I understand WHY it's happening (root cause)

## Fix Plan Ready? ✓
- [ ] I have a clear fix approach documented mentally or on paper
- [ ] I can test my fix before committing
- [ ] The fix doesn't change unrelated functionality
- [ ] No simpler solution exists first

## Testing Strategy Defined? ✓
- [ ] Existing tests cover this code path (or will break)
- [ ] New test added to prevent regression
- [ ] Test runs and passes locally BEFORE git add
```

---

## 🚫 Common Mistakes to Avoid

### ❌ DON'T:
1. **Fix without understanding** → Always read the full stack trace first
2. **Use `any` type** → Define proper types for all variables/functions
3. **Remove console.log blindly** → Add debug helper function instead, keep it in dev only
4. **Change multiple things at once** → Fix one logical issue per commit
5. **Ignore failing tests** → Every test failure is a clue

### ✅ DO:
1. **Read error messages completely** - First 3 lines often contain the real issue
2. **Search for similar bugs in git history** (`git log --all -S "error message"`)
3. **Use TypeScript strict mode** to catch type issues early
4. **Add tests before production code changes when possible**
5. **Keep fixes minimal and focused**

---

## 🔄 Post-Fix Verification (Run After Each Fix)

```bash
# 1. Verify syntax is valid
npx tsc --noEmit && echo "✓ No TypeScript errors" || exit 1

# 2. Run linter to catch style issues  
npm run lint | grep -v "warning" > /dev/null && echo "✓ Lint clean"

# 3. Run full test suite
npm test && echo "✓ All tests pass" || exit 1

# 4. Check for new issues in git diff
git diff --check && echo "✓ No whitespace/formatting changes"

# If all green, safe to commit! ✅
```

---

## 📝 Example Workflow: Complete Bug Fix Cycle

### Scenario: Function returns wrong value when input is empty array

#### Step 1: Reproduce & Understand
```typescript
// Before fix - document the bug
function calculateDiscount(items) {
  // BUG: Returns undefined for empty arrays instead of 0
  if (items.length === 0) return; // ❌ WRONG! Should be return 0
  
  const total = items.reduce((sum, item) => sum + item.price, 0);
  return Math.max(0, total * 0.1);
}

// Test to confirm bug:
console.log(calculateDiscount([])); // undefined instead of 0 ❌
```

#### Step 2: Create Fix Plan
- Root cause: Missing return value for empty array case
- Fix: Change `return` to `return 0;`
- Add test: Verify empty array returns 0

#### Step 3: Implement Fix with Test First
```typescript
// src/services/discount.test.ts (create before fixing)
describe('calculateDiscount', () => {
  test('returns 0 for empty array', () => {
    expect(calculateDiscount([])).toBe(0); // New test!
  });
  
  test('calculates correct discount for items', () => {
    const result = calculateDiscount([{ price: 100 }]);
    expect(result).toBe(10);
  });
});

// src/services/discount.ts (the fix)
function calculateDiscount(items: Product[]): number {
  if (!items || items.length === 0) return 0; // ✅ FIXED
  
  const total = items.reduce((sum, item) => sum + item.price, 0);
  return Math.max(0, total * 0.1);
}
```

#### Step 4: Verify Fix
```bash
# Run new test first to ensure it would fail before fix
npm test -- discount.test.ts           # Shows failure (expected)

# Apply fix, run again - should pass now  
git add . && npm test                 # All tests green ✅
npx tsc --noEmit                      # No type errors ✅
```

---

## 🎓 Pro Tips from Real-World Experience

1. **When stuck after 30 minutes**: Take a break, come back fresh or ask for help
2. **Rubber duck debugging**: Explain the code line-by-line to someone (even an imaginary colleague)
3. **Binary search on bugs**: Comment out half the file at a time if unsure where issue is
4. **Check recent changes first**: `git log -1 --stat` shows what changed before this bug appeared
5. **Look for TODO/FIXME comments**: They often point to known issues

---

## 📚 Quick Reference: Bug Type Identification

### Syntax Bugs Show As:
- Red squiggly lines in editor (VS Code)
- Compilation errors with file/line numbers
- Linter warnings that are actually critical
- Stack traces pointing to specific line/column

**Fix approach**: Fix the exact error reported, then verify syntax is clean before moving on.

### Logic Bugs Show As:
- Tests failing silently or intermittently  
- Wrong output but code "runs" without errors
- Race conditions (works sometimes, not others)
- Edge cases that break unexpectedly

**Fix approach**: Add debug logs → isolate problem area → understand expected vs actual → fix root cause → add test to prevent regression.
