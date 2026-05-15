name: implement-new-feature
description: Guide for implementing new features in V-Computer (NestJS + Vue 3)
license: MIT
compatibility: opencode
metadata:
  audience: solo-developers
  workflow: planning-first
---

## Feature Implementation Workflow

### Step 1: Gather Requirements

Collect from user before coding:
- **Feature description**: What should it do?
- **Acceptance criteria**: How to verify success?
- **User stories**: Who uses it and why?
- **Dependencies**: Related features or APIs needed?

Ask clarifying questions if requirements are vague.

### Step 2: Plan Architecture

**Backend (NestJS)** - Create module structure:
```
backend/src/modules/[feature-name]/
├── [feature-name].controller.ts   # Request handlers only
├── [feature-name].module.ts       # Module registration
├── [feature-name].service.ts      # Business logic
├── dto/                           # Input validation DTOs
│   ├── create-[name].dto.ts
│   └── update-[name].dto.ts
└── entities/[name].entity.ts      # Prisma models (if new entity)
```

**Frontend (Vue 3)** - Create components:
```
frontend/src/
├── views/[FeatureName]View.vue    # Page container
├── services/[feature-name].api.ts # API client calls
└── stores/[feature-name].store.ts # Pinia state (if needed)
```

### Step 3: Database Schema

Update Prisma schema (`backend/prisma/schema.prisma`):
- Define new models if required
- Add relationships to existing models
- Run migrations: `npx prisma migrate dev --name [feature-name]`

### Step 4: Backend Implementation

**DTOs** (validate input, no business logic):
```typescript
// dto/create-product.dto.ts
import { IsString, IsNumber, Min } from 'class-validator'

export class CreateProductDto {
  @IsString()
  name: string

  @IsNumber()
  price: number

  @IsNumber({ minMessage: 'Price must be > 0' })
  stock: number
}
```

**Controller** (handle requests, return responses):
- Use `@Get`, `@Post`, etc. decorators
- Call service methods only
- Return DTO or error objects
- Never access DB directly

**Service** (business logic):
- Implement core functionality
- Handle transactions if needed
- Throw exceptions for errors
- Inject dependencies via constructor

### Step 5: Frontend Implementation

**API Service** (`frontend/src/services/[name].api.ts`):
```typescript
export const [featureName]Api = {
  get: (params?: any) => api.get(`/[${name}]/${id}`, { params }),
  getById: (id) => api.get(`/${[name]}s/${id}`),
  post: (data) => api.post('/[${name}s]', data)),
  put: (id, data) => api.put(`/${[name}s]`, id, data),
  delete: (id) => api.delete(`/[${name}s]/${id}`),
}
```

**View Component**:
- Use `<script setup lang="ts">`
- Composables for reusable logic (`useToast`, `useAuth`)
- Pinia stores for shared state
- Vuetify components from existing patterns
- Handle loading/error states with toasts

### Step 6: Testing & Deployment

**Test Checklist**:
- [ ] Backend endpoints return correct status codes
- [ ] Validation rejects invalid input
- [ ] Frontend displays data correctly
- [ ] Error handling shows user-friendly messages
- [ ] Authentication required where needed

**Run checks**:
```bash
# TypeScript check
npm run lint --prefix backend
npm run typecheck --prefix frontend

# Start services for manual testing
./scripts/start.sh
```

### Project Conventions to Follow

**Naming**: Descriptive, explicit names (e.g., `getProductById`, not `getData`)

**Type Safety**: No `any` types unless absolutely necessary

**Error Handling**: Consistent format with meaningful messages

**Code Style**: Clean, readable - single responsibility per function/file

**State Management**: Pinia for shared state, avoid global variables

### Quick Reference Paths

- **Backend modules**: `/work/vcomputer/backend/src/modules/`
- **Frontend views**: `/work/vcomputer/frontend/src/views/`
- **API services**: `/work/vcomputer/frontend/src/services/api.ts`
- **Pinia stores**: `/work/vcomputer/frontend/src/stores/`
- **Prisma schema**: `/work/vcomputer/backend/prisma/schema.prisma`

### When to Ask User

Before proceeding, confirm:
1. Feature requirements are clear
2. No conflicting existing features
3. Database changes needed? Yes/No
4. Authentication required? Yes/No
5. Priority level (High/Medium/Low)
