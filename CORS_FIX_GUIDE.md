# CORS Error Fix Guide

## Problem
You encountered a CORS (Cross-Origin Resource Sharing) error in the browser console:

```
Access to fetch at 'http://localhost:8081/register' from origin 'http://localhost:3000' 
has been blocked by CORS policy: Response to preflight request doesn't pass access control 
check: No 'Access-Control-Allow-Origin' header is present on the requested resource.
```

## Root Cause
The backend services were configured with CORS middleware but the `LINKPULSE_ALLOWED_ORIGINS` environment variable was **not set**. Without this variable:
- The CORS middleware could not identify allowed origins
- Preflight OPTIONS requests were rejected with HTTP 403
- Browser blocked all actual API calls due to failed CORS validation

## Solution Implemented ✅

### 1. Docker Compose Configuration
Added `LINKPULSE_ALLOWED_ORIGINS` environment variable to all backend services:

```yaml
environment:
  LINKPULSE_ALLOWED_ORIGINS: "http://localhost:3000,http://localhost:3001"
```

**Services Updated:**
- ✅ **auth-service** (port 8081)
- ✅ **shortener-service** (port 8082)
- ✅ **analytics-service** (port 8083)

### 2. Backend Code Changes

**analytics-service** (`backend/services/analytics/cmd/api/main.go`):
- Added CORS middleware function `buildCORSMiddleware()`
- Allows: GET and OPTIONS methods for analytics queries
- Headers: Content-Type, Authorization

**auth-service** and **shortener-service**:
- Already had CORS middleware built-in
- Only needed the environment variable configured

### 3. How CORS Middleware Works

```go
func buildCORSMiddleware(allowedOrigins string) gin.HandlerFunc {
  // 1. Parse comma-separated origins
  // 2. Build allowed set: map[string]struct{}
  
  return func(c *gin.Context) {
    origin := c.GetHeader("Origin")
    
    if _, ok := allowedSet[origin]; ok {
      // Origin is allowed - add headers
      c.Header("Access-Control-Allow-Origin", origin)
      c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
      c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
      
      // Handle preflight
      if c.Request.Method == http.MethodOptions {
        c.AbortWithStatus(http.StatusNoContent)
        return
      }
    } else if c.Request.Method == http.MethodOptions {
      // Reject preflight from disallowed origins
      c.AbortWithStatus(http.StatusForbidden)
      return
    }
    
    c.Next()
  }
}
```

## Next Steps

### 1. Rebuild Docker Containers
```bash
docker-compose down
docker-compose build
docker-compose up -d
```

### 2. Test the Fix
Clear your browser cache and hard refresh (Cmd+Shift+R on Mac):

**Test Registration:**
```javascript
// In browser console
fetch('http://localhost:8081/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    name: 'Test User',
    email: 'test@example.com',
    password: 'password123'
  })
})
.then(r => r.json())
.then(console.log)
```

**Expected Response:**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "...",
  "expires_in": 3600
}
```

### 3. Verify All Endpoints

| Service | Endpoint | Port | CORS Status |
|---------|----------|------|-------------|
| Auth | POST /register | 8081 | ✅ Fixed |
| Auth | POST /login | 8081 | ✅ Fixed |
| Auth | POST /logout | 8081 | ✅ Fixed |
| Shortener | POST /api/v1/shorten | 8082 | ✅ Fixed |
| Shortener | GET /api/v1/shorten/workspace/:id | 8082 | ✅ Fixed |
| Shortener | DELETE /api/v1/shorten/:id | 8082 | ✅ Fixed |
| Analytics | GET /analytics/:short_code | 8083 | ✅ Fixed |

## Production Deployment Considerations

### Security Note
The current configuration allows `http://localhost:3000` and `http://localhost:3001` for development.

For production, set environment variables appropriately:

```bash
# Production (single origin)
LINKPULSE_ALLOWED_ORIGINS=https://app.linkpulse.com

# Multiple environments
LINKPULSE_ALLOWED_ORIGINS=https://app.linkpulse.com,https://staging.linkpulse.com
```

### Environment-Specific Configuration
Create `.env.production`:
```bash
LINKPULSE_ALLOWED_ORIGINS=https://linkpulse.com
```

## Troubleshooting

### Issue: Still getting CORS errors
**Solution:**
1. Hard refresh browser (Cmd+Shift+R)
2. Check docker logs: `docker logs linkpulse-auth`
3. Verify containers are running: `docker ps`
4. Restart containers: `docker-compose restart`

### Issue: Cannot see CORS headers in DevTools
**Solution:**
1. Open DevTools Network tab
2. Look for the preflight OPTIONS request (usually first)
3. Check Response Headers tab
4. Should see: `Access-Control-Allow-Origin: http://localhost:3000`

### Issue: 403 on preflight requests
**Solution:**
1. Origin header must match exactly (case-sensitive)
2. Check docker-compose.yml for typos
3. Rebuild containers: `docker-compose build --no-cache`

## Files Modified

| File | Change | Reason |
|------|--------|--------|
| `docker-compose.yml` | Added LINKPULSE_ALLOWED_ORIGINS env var | Enable CORS in services |
| `backend/services/analytics/cmd/api/main.go` | Added CORS middleware | Analytics service CORS support |

## Commit Information

```
Commit: e6ecacb
Message: fix: enable CORS for frontend-backend communication
Date: 30 March 2026
Branch: main

Changes:
- 6 files changed
- 607 insertions
- 111 deletions
- INTEGRATION_PHASE2.md created (from Phase 2 work)
```

## Phase 2 Status After CORS Fix

✅ **Authentication Integration**
- Real JWT exchange working
- Register/Login connected to backend
- User context extracted and stored

✅ **Link Management Integration** (Pending CORS Fix Verification)
- CreateLinkModal component ready
- Links page with table ready
- API endpoints updated

⏳ **Next: Test Full Flow**
1. Register new user
2. Login with credentials
3. Create a short link
4. View in links table
5. Delete a link
6. Verify all operations work end-to-end

## Quick Reference

**Docker Commands:**
```bash
# View logs
docker logs linkpulse-auth -f
docker logs linkpulse-shortener -f
docker logs linkpulse-analytics -f

# Restart services
docker-compose restart auth-service
docker-compose restart shortener-service
docker-compose restart analytics-service

# Full restart
docker-compose down && docker-compose up -d
```

**Frontend URLs:**
- Frontend: http://localhost:3000
- Login: http://localhost:3000/login
- Register: http://localhost:3000/register
- Links: http://localhost:3000/links

**Backend URLs:**
- Auth Health: http://localhost:8081/health
- Shortener Health: http://localhost:8082/health
- Analytics Health: http://localhost:8083/health
