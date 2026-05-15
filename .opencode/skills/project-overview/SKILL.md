---
name: project-overview
description: Understand the vcomputer project architecture and structure
license: MIT
compatibility: opencode
metadata:
  audience: developers
  workflow: development
---

## Project Overview

vcomputer is a full-stack inventory management application with Vue.js frontend and NestJS backend.

## Tech Stack

- **Frontend**: Vue 3 + Vuetify 3 + TypeScript
- **Backend**: NestJS + TypeScript + Prisma ORM
- **Database**: PostgreSQL
- **Deployment**: Docker + Nginx

## Project Structure

```
/work/vcomputer/
├── frontend/           # Vue 3 + Vuetify app
│   └── src/
│       ├── views/      # 26 view components
│       ├── components/ # Reusable components
│       ├── stores/     # Pinia auth store
│       ├── services/  # API client
│       ├── composables/ # Vue composables
│       └── router/     # Vue Router config
├── frontend-react/     # React alternative (not in use)
├── backend/            # NestJS API
│   └── src/
│       ├── modules/   # Feature modules
│       │   ├── auth/
│       │   ├── items/
│       │   ├── categories/
│       │   ├── customers/
│       │   ├── transactions/
│       │   ├── daily-expenses/
│       │   ├── daily-incomes/
│       │   ├── expense-categories/
│       │   ├── income-categories/
│       │   ├── expense-templates/
│       │   ├── warranties/
│       │   ├── repairs/
│       │   ├── debts/
│       │   ├── moneysources/
│       │   ├── public-products/
│       │   ├── statistics/
│       │   ├── system-log/
│       │   ├── config/
│       │   └── ai/
│       ├── common/    # Shared utilities
│       └── prisma/    # Database schema
├── docker-compose.yml  # Production services
└── nginx.conf         # Reverse proxy config
```

## Key Features

1. **Items Management** (`ItemsView.vue`)
   - CRUD for inventory items
   - Categories with parent-child hierarchy
   - Photo upload
   - Purchase import dialog

2. **POS/Sales** (`PosView.vue`)
   - Product selection by category
   - Cart management
   - Warranty creation

3. **Transactions** (`TransactionsView.vue`)
   - Money in/out tracking
   - Cancel/refund functionality

4. **Daily Expenses** (`DailyExpensesView.vue`)
   - Expense templates
   - Category management

5. **Categories** (`CategoriesView.vue`)
   - Parent-child hierarchy
   - Category display: `parent > child`

## Common Patterns

1. **ItemTypeSelect Component** (`ItemTypeSelect.vue`)
   - Shows category with parent: `parent > child`
   - Returns only ID (not object)

2. **MoneyInput Component** (`MoneyInput.vue`)
   - Formats VND currency
   - Live preview

3. **DateTimeSelect Component** (`DateTimeSelect.vue`)
   - datetime-local input
   - Syncs with server time

4. **Authentication Flow**
   - Token persisted in localStorage
   - Validated on app mount
   - Redirects to `/login` if invalid

## API Endpoints

- `POST /api/auth/login`, `POST /api/auth/register`
- `GET/POST/PUT/DELETE /api/items`
- `GET/POST/PUT/DELETE /api/categories`
- `GET/POST/PUT/DELETE /api/transactions`
- `POST /api/transactions/:id/cancel` - Refund transaction
- `GET/POST /api/expense-templates`
- `GET/POST /api/daily-expenses`, `GET/POST /api/daily-incomes`

## Database

- PostgreSQL with Prisma ORM
- 16 tables: items, categories, customers, transactions, warranties, repairs, debts, moneysources, etc.

## Deployment

```bash
docker-compose build frontend backend
docker-compose up -d frontend backend nginx
```

## When to Use This Skill

Use this when you need to:
- Understand where a feature is implemented
- Find related components/views
- Understand data flow between frontend and backend
- Know how to add new features following existing patterns