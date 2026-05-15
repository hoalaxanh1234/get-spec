---
name: e2e-test
description: Create E2E tests using Playwright for backend and frontend
license: MIT
metadata:
  audience: developers
  workflow: testing
---

## E2E Testing Guide - Playwright

### Project Context

- **Monorepo**: Backend (NestJS) + Frontend (Vue 3 + Vuetify)
- **Backend**: NestJS + Prisma + PostgreSQL (host: postgres, port: 5432)
- **Frontend**: Vue 3 + Vuetify + Pinia + Vue Router
- **Authentication**: JWT Bearer tokens (stored in localStorage)
- **API Base**: `/api`
- **Way to implement**: we will add data-testid attributes to key elements for easier selection in tests

### Setup

```bash
# Install Playwright in backend (recommended for monorepo)
cd /work/vcomputer/backend
npm init playwright@latest -- --yes --quiet

# Or install globally
npm install -g @playwright/test
```

### Configuration

Create `playwright.config.ts` in backend root:

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: {
    command: 'npm run start:dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
})
```

### Environment Variables

Create `.env.test` in backend:

```
DATABASE_URL="postgresql://postgres:postgres@postgres:5432/vcomputer_test"
JWT_SECRET="test-jwt-secret-key"
JWT_EXPIRES_IN="15m"
```

### Fixtures

Create `e2e/fixtures/global-fixture.ts`:

```typescript
import { test as base } from '@playwright/test'
import { PrismaClient } from '@prisma/client'

const prisma = new PrismaClient()

export const test = base.extend({
  async db() {
    await prisma.$connect()
    yield prisma
    await prisma.$disconnect()
  },
  async cleanDb({ db }: { db: PrismaClient }) {
    await db.transaction.deleteMany()
    await db.item.deleteMany()
    await db.category.deleteMany()
    await db.user.deleteMany()
  },
})

export { expect } from '@playwright/test'
```

### Authentication Flow

```typescript
import { test, expect } from './fixtures/global-fixture'

test.describe('Authentication', () => {
  test('login with valid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name="username"]', 'testuser')
    await page.fill('[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    await expect(page).toHaveURL('/items')
    const token = await page.evaluate(() => localStorage.getItem('token'))
    expect(token).toBeTruthy()
  })

  test('login with Thông tin đăng nhập không hợp lệ shows error', async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name="username"]', 'wrong')
    await page.fill('[name="password"]', 'wrong')
    await page.click('button[type="submit"]')
    
    await expect(page.locator('.v-alert')).toContainText('Invalid')
  })
})
```

### API Integration Tests

```typescript
import { test, expect } from './fixtures/global-fixture'

test.describe('API Tests', () => {
  let token: string

  test.beforeAll(async ({ request }) => {
    const res = await request.post('/api/auth/login', {
      data: { username: 'testuser', password: 'password123' }
    })
    const body = await res.json()
    token = body.access_token
  })

  test('GET /api/items returns items', async ({ request }) => {
    const res = await request.get('/api/items', {
      headers: { Authorization: `Bearer ${token}` }
    })
    expect(res.ok()).toBeTruthy()
  })

  test('POST /api/items creates item', async ({ request }) => {
    const res = await request.post('/api/items', {
      headers: { Authorization: `Bearer ${token}` },
      data: { name: 'Test Item', price: 10000, categoryId: 'cat-1' }
    })
    expect(res.ok()).toBeTruthy()
    const item = await res.json()
    expect(item.name).toBe('Test Item')
  })
})
```

### User Flows

```typescript
import { test, expect } from './fixtures/global-fixture'

test.describe('Item Management', () => {
  test.beforeEach(async ({ page, cleanDb }) => {
    await page.login('admin', 'password')
  })

  test('create new item via form', async ({ page }) => {
    await page.goto('/items')
    await page.click('text=Add')
    await page.fill('[name="name"]', 'New Item')
    await page.fill('[name="price"]', '50000')
    await page.click('button:has-text("Save")')
    
    await expect(page.locator('text=New Item')).toBeVisible()
  })

  test('edit existing item', async ({ page }) => {
    await page.goto('/items')
    await page.click('[row="0"] >> text=Edit')
    await page.fill('[name="name"]', 'Updated Item')
    await page.click('button:has-text("Save")')
    
    await expect(page.locator('text=Updated Item')).toBeVisible()
  })

  test('delete item with confirmation', async ({ page }) => {
    await page.goto('/items')
    await page.click('[row="0"] >> text=Delete')
    await page.click('button:has-text("Confirm")')
    
    await expect(page.locator('text=Deleted')).toBeVisible()
  })
})
```

### POS View Tests

```typescript
import { test, expect } from './fixtures/global-fixture'

test.describe('POS View', () => {
  test.beforeEach(async ({ page, cleanDb, db }) => {
    // Setup: Create test items
    await db.item.createMany({
      data: [
        { id: 'item-1', name: 'Coffee', price: 25000, categoryId: 'cat-1' },
        { id: 'item-2', name: 'Tea', price: 20000, categoryId: 'cat-1' },
      ]
    })
    await page.login('cashier', 'password')
    await page.goto('/pos')
  })

  test('add item to cart', async ({ page }) => {
    await page.click('text=Coffee')
    await page.click('text=Add to Cart')
    
    await expect(page.locator('.cart-item')).toContainText('Coffee')
    await expect(page.locator('.total')).toContainText('25,000')
  })

  test('complete transaction', async ({ page }) => {
    await page.click('text=Coffee')
    await page.click('text=Checkout')
    
    await expect(page.locator('.receipt')).toBeVisible()
  })
})
```

### Visual Regression (Optional)

```bash
npm install @chromium-ui/screenshot
```

```typescript
import { test, expect } from '@playwright/test'
import { visualDiff } from '@chromium-ui/screenshot'

test('homepage visual', async ({ page }) => {
  await page.goto('/')
  
  await expect(page).toPassVisualTest()
})
```

### Running Tests

```bash
# All tests
npx playwright test

# Specific file
npx playwright test e2e/auth.spec.ts

# With UI
npx playwright test --ui

# Debug mode
npx playwright test --debug

# Generate report
npx playwright show-report
```

### CI Configuration (GitHub Actions)

```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - name: Start services
        run: docker-compose up -d postgres
      - name: Install and Run Tests
        run: |
          cd backend
          npm ci
          npm run migrate
          npx playwright install --with-deps
          npx playwright test
```

### Page Object Pattern

```typescript
// e2e/pages/LoginPage.ts
import { Page, Locator } from '@playwright/test'

export class LoginPage {
  readonly username: Locator
  readonly password: Locator
  readonly submit: Locator
  readonly error: Locator

  constructor(private page: Page) {
    this.username = page.locator('[name="username"]')
    this.password = page.locator('[name="password"]')
    this.submit = page.locator('button[type="submit"]')
    this.error = page.locator('.v-alert')
  }

  async login(username: string, password: string) {
    await this.username.fill(username)
    await this.password.fill(password)
    await this.submit.click()
  }
}

// Usage
test('login', async ({ page }) => {
  const loginPage = new LoginPage(page)
  await loginPage.goto()
  await loginPage.login('user', 'pass')
  await expect(page).toHaveURL('/items')
})
```

### Database Setup for Tests

Create `prisma/schema.test.prisma`:

```prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

model User {
  id       String @id @default(uuid())
  username String @unique
  password String
  // ... other fields
}
```

Run migration:
```bash
DATABASE_URL="postgresql://..." npx prisma migrate dev --name init --schema prisma/schema.test.prisma
```

### Troubleshooting

- **Port already in use**: Stop dev servers or change port in config
- **Auth failures**: Ensure JWT_SECRET matches between test and app
- **Database errors**: Clean DB between tests with `cleanDb` fixture
- **Timeout errors**: Increase `webServer.timeout` in config
- **Flaky tests**: Add `test.sleep()` or wait for specific elements

### Best Practices

1. **Isolate tests**: Each test should clean its own data
2. **Use fixtures**: Share setup code across tests
3. **Be specific**: Test one thing per test
4. **Name clearly**: `test.describe('feature', () => { test('sub-action', ...) })`
5. **Handle auth**: Store token in context, reuse across tests
6. **Take screenshots**: On failure for debugging
7. **Use locators**: Prefer `getByRole`, `getByLabel` over XPath