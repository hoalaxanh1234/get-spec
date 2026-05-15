name: devops-helper
description: Comprehensive DevOps assistance for V-Computer (Docker, Prisma, CI/CD)
license: MIT
compatibility: opencode
metadata:
  audience: solo-developers
  workflow: step-by-step-validation

---

## Overview

DevOps skill for V-Computer project covering Docker containerization, database management, environment setup, and deployment automation. Follows your mixed local/remote approach with validation at each step.

---

## Core Capabilities

### Environment Setup & Management

**Development (Docker Compose)**:
```bash
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# View logs from specific service
docker-compose -f docker-compose.dev.yml logs [frontend|backend]

# Stop and clean containers
docker-compose down --volumes

# Rebuild frontend/backend
docker-compose build frontend backend
```

**Production Deployment**:
```bash
./scripts/start-prod.sh  # Your production startup script
docker-compose -f docker-compose.yml up -d
```

### Database Management (Prisma)

**Development Migrations**:
1. **Create migration for schema changes**:
   ```bash
   npx prisma migrate dev --name [migration-name]
   ```
2. **Generate Prisma Client after changes**:
   ```bash
   npx prisma generate
   ```

**Database Reset (Testing)**:
```bash
# Drop and recreate database
npx prisma migrate reset

# With seed data if available
npx prisma migrate reset --force-reset-seed
```

**Production Migrations**:
```bash
npx prisma migrate deploy
npx prisma generate
```

### Environment Configuration

**.env File Setup**:
- Copy template: `cp .env.example .env`
- Required variables (from `.env.example`):
  - `DATABASE_URL`: PostgreSQL connection string
  - `JWT_SECRET`: Secret key for JWT tokens
  - `VITE_API_URL`: Frontend API endpoint (`/api` or full URL)

**Service Configuration**:
```yaml
# Backend .env
NODE_ENV=development|production
PORT=3000

# Frontend .env (if using env files)
VITE_API_URL=http://localhost:3000/api
```

### Build & Test Commands

**Backend (NestJS)**:
```bash
cd backend

# TypeScript check
npm run lint
npm run typecheck  # if configured

# Tests (if available)
npm test
npm run test:cov   # With coverage

# Build
npm run build
```

**Frontend (Vue 3 + Vite)**:
```bash
cd frontend

# Type check & lint
npm run lint
npm run typecheck  # if configured

# Development server
npm run dev        # Hot reload at http://localhost:5173

# Production build
npm run build      # Optimized bundle
npm run preview    # Preview production build
```

### Health Checks & Verification

**Service Status**:
```bash
# Check backend health endpoint
curl http://localhost:3000/api/health 2>/dev/null || echo "Backend not running"

# Verify database connection (via Prisma introspection)
npx prisma db pull --schema-backend/backend/prisma/schema.prisma 2>&1 | head -5
```

**Docker Container Status**:
```bash
docker-compose ps           # List all containers and status
docker images               # View container images

# Check logs for errors
docker-compose logs backend frontend nginx
```

### Deployment Scripts

**Production Startup**:
```bash
cd /work/vcomputer

# Run your production script (already exists)
./scripts/start-prod.sh

# Or manual docker compose start
docker-compose -f docker-compose.yml up -d
```

**Nginx Configuration**:
- Config at: `/work/vcomputer/nginx.conf`
- Nginx serves frontend builds and proxies to backend API
- SSL certificates in `/ssl/` directory (if using HTTPS)

---

## Step-by-Step Workflow

### Adding New Feature - DevOps Checklist

**1. Database Changes**:
```bash
# Check if new entities needed
ls /work/vcomputer/backend/prisma/*.prisma  # If exists, edit schema

# Create migration
cd backend
npx prisma migrate dev --name [feature-name]
npm run generate
```

**2. Backend Development Mode**:
```bash
cd backend
NODE_ENV=development npm run start:dev
# Expected output on port 3000
```

**3. Frontend Development Mode**:
```bash
cd frontend  
npm run dev
# Expected output at http://localhost:5173
```

**4. Test Integration**:
- Open browser at `http://localhost:5173`
- Check console for API errors (should proxy to `/api`)
- Verify backend logs show requests

### Troubleshooting Common Issues

**Database Connection Failed**:
```bash
# Check PostgreSQL is running
docker-compose -f docker-compose.dev.yml ps | grep db

# View database container logs
docker-compose logs db

# Test connection manually (if psql available)
psql postgres://user:pass@host:5432/dbname
```

**Port Already in Use**:
```bash
# Find process on port 3000/5173
lsof -i :3000  # Backend
lsof -i :5173  # Frontend

# Or kill all and restart containers
docker-compose down && docker-compose up -d
```

**Frontend Can't Reach API**:
- Check `frontend/.env` has correct `VITE_API_URL`
- Verify backend running on expected port (default: 3000)
- Check nginx.conf proxy settings if using production build

---

## Quick Reference Commands

### Daily Development Flow
```bash
# Morning setup
./scripts/start.sh                    # Start all services
docker-compose logs -f        # Follow combined logs

# Evening cleanup
docker-compose down           # Stop containers (keeps data)

# Database operations
npx prisma migrate dev --name [name]  # New migration
npx prisma generate                  # Regenerate client
```

### Production Deployment Flow
```bash
./scripts/start-prod.sh                # Full production start
docker-compose ps              # Verify all containers running
curl http://localhost:3000/api/health  # Quick health check
```

---

## Project-Specific Paths

- **Docker Compose Dev**: `/work/vcomputer/docker-compose.dev.yml`
- **Docker Compose Prod**: `/work/vcomputer/docker-compose.yml`
- **Prisma Schema**: `/work/vcomputer/backend/prisma/schema.prisma`
- **Production Script**: `/work/vcomputer/scripts/start-prod.sh`
- **Nginx Config**: `/work/vcomputer/nginx.conf`
- **SSL Certificates**: `/work/vcomputer/ssl/`

---

## When to Ask User

Before executing any command, confirm:
1. What operation needs to be performed? (setup/migrate/deploy/debug)
2. Current environment state? (dev/prod/local/remote)
3. Expected outcome and success criteria?
4. Should I verify before making changes? Yes/No

---

## Validation Checklist Before Changes

- [ ] Backup current `.env` files if modifying
- [ ] Test locally first in dev containers
- [ ] Check existing services status (`docker-compose ps`)
- [ ] Review migration name conflicts (unique naming)
- [ ] Verify port availability before starting services
- [ ] Confirm database URL format matches your PostgreSQL setup
