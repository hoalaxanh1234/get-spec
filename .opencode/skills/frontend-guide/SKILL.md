---
name: frontend-guide
description: Guide for Vue.js frontend development patterns
license: MIT
compatibility: opencode
metadata:
  audience: frontend-developers
  workflow: development
---

## Vue 3 + Vuetify Development Guide

## Views Structure

All views are in `/work/vcomputer/frontend/src/views/`:
- ItemsView.vue - Item management with filters
- PosView.vue - Point of Sale
- TransactionsView.vue - Transaction management
- DailyExpensesView.vue - Daily expenses
- CategoriesView.vue - Category management
- ExpenseIncomeCategoriesView.vue - Combined category management

## Common Components

### ItemTypeSelect
Location: `src/components/ItemTypeSelect.vue`
- Shows parent > child hierarchy
- Returns only category ID (string), not full object

```vue
<ItemTypeSelect v-model="form.categoryId" label="Loại hàng" />
```

### MoneyInput
Location: `src/components/MoneyInput.vue`
- Currency formatting for VND
- Live preview

```vue
<MoneyInput v-model="form.amount" label="Số tiền" />
```

### DateTimeSelect
Location: `src/components/DateTimeSelect.vue`
- datetime-local input
- Syncs server time on mount

```vue
<DateTimeSelect v-model="form.buyDateTime" label="Ngày mua" />
```

### ClientSelect
Location: `src/components/ClientSelect.vue`
- Autocomplete with search
- Inline create form

## View Pattern Template

```vue
<template>
  <div>
    <div class="content-header mb-4">
      <div class="d-flex justify-space-between align-center">
        <h1 class="text-h5">Title</h1>
        <v-btn color="primary" @click="openDialog()">Add</v-btn>
      </div>
    </div>

    <v-card>
      <v-data-table :headers="headers" :items="items" :loading="loading">
        <template #item.actions="{ item }">
          <v-btn icon @click="openDialog(item)">Edit</v-btn>
          <v-btn icon color="error" @click="deleteItem(item)">Delete</v-btn>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="400">
      <v-card>
        <v-card-title>{{ editing ? 'Edit' : 'Add' }}</v-card-title>
        <v-card-text>
          <!-- Form fields -->
        </v-card-text>
        <v-card-actions>
          <v-btn @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="save">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../services/api'

const items = ref([])
const loading = ref(false)
const dialog = ref(false)
const editing = ref(null)

const form = ref({})

const headers = [
  { title: 'Name', key: 'name' },
  { title: 'Actions', key: 'actions', sortable: false },
]

async function fetchItems() {
  loading.value = true
  try {
    const res = await api.get('/endpoint')
    items.value = res.data
  } finally {
    loading.value = false
  }
}

function openDialog(item?: any) {
  editing.value = item || null
  form.value = item ? { ...item } : {}
  dialog.value = true
}

async function save() {
  if (editing.value) {
    await api.put(`/endpoint/${editing.value.id}`, form.value)
  } else {
    await api.post('/endpoint', form.value)
  }
  dialog.value = false
  fetchItems()
}

onMounted(() => fetchItems())
</script>
```

## API Service

Location: `src/services/api.ts`

```typescript
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

// Add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

export default api
```

## Router Configuration

Location: `src/router/index.ts`

```typescript
{
  path: '/items',
  component: () => import('../views/ItemsView.vue'),
  meta: { requiresAuth: true },
}
```

## Deployment

```bash
cd /work/vcomputer
docker-compose build frontend
docker-compose up -d frontend
```