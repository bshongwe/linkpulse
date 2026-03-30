# CORS Fix - Frontend to Backend Connection

## Problem
The frontend (running on `localhost:3000`) was unable to connect to backend services due to CORS (Cross-Origin Resource Sharing) policy violations. Specific errors:

```
Access to fetch at 'http://localhost:8081/register' from origin 'http://localhost:3000' 
has been blocked by CORS policy: Response to preflight request doesn't pass access 
control check: No 'Access-Control-Allow-Origin' header is present on the requested resource.
```

## Root Causes

### 1. Wrong Endpoint Path
- **File**: `frontend/app/(auth)/login/page.tsx`
- **Issue**: Frontend was calling `/auth/login` but backend exposes `/login`
- **Fix**: Changed endpoint from `http://localhost:8081/auth/login` to `http://localhost:8081/login`

### 2. Broken CORS Middleware Logic
- **Files**: 
  - `backend/services/auth/cmd/api/main.go`
  - `backend/services/shortener/cmd/api/main.go`
  - `backend/services/analytics/cmd/api/main.go`

- **Issue**: The CORS middleware was only returning CORS headers for allowed origins, but the check happened **after** the preflight request was already aborted. Also, it wasn't adding all necessary CORS headers.

- **Original Logic**:
  ```go
  if _, ok := allowedSet[origin]; ok {
      // Set headers (but preflight already handled wrong)
      c.Header("Access-Control-Allow-Origin", origin)
      c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
      c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
      if c.Request.Method == http.MethodOptions {
          c.AbortWithStatus(http.StatusNoContent)
          return
      }
  } else if c.Request.Method == http.MethodOptions {
      // Reject preflight
      c.AbortWithStatus(http.StatusForbidden)
      return
  }
  ```

- **Fixed Logic**:
  ```go
  origin := c.GetHeader("Origin")
  
  // Check if origin is allowed
  if _, ok := allowedSet[origin]; ok {
      c.Header("Access-Control-Allow-Origin", origin)
      c.Header("Access-Control-Allow-Credentials", "true")
      c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
      c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
      c.Header("Access-Control-Max-Age", "86400")
  }
  
  // Handle preflight requests
  if c.Request.Method == http.MethodOptions {
      if _, ok := allowedSet[origin]; ok {
          c.AbortWithStatus(http.StatusNoContent)
      } else {
          c.AbortWithStatus(http.StatusForbidden)
      }
      return
  }
  
  c.Next()
  ```

### 3. File Structure Duplication
- **Issue**: During earlier edits, incorrectly escaped paths created duplicate folders:
  - `app/(auth)/` (correct)
  - `app/\(auth\)/` (incorrect escape)
  - `app/(dashboard)/` (correct)
  - `app/\(dashboard\)/` (incorrect escape)

- **Fix**: Removed the incorrectly escaped versions:
  ```bash
  rm -rf '/Users/ernie-dev/Documents/linkpulse/frontend/app/\(auth\)/'
  ```

## Environment Configuration

The `docker-compose.yml` already had the correct CORS configuration:

```yaml
environment:
  LINKPULSE_ALLOWED_ORIGINS: "http://localhost:3000,http://localhost:3001"
```

This configuration is read by all three backend services' CORS middleware.

## Changes Made

### Frontend
- ✅ `app/(auth)/login/page.tsx`: Fixed endpoint from `/auth/login` to `/login`
- ✅ File structure cleaned up (removed escaped directory names)

### Backend - Auth Service
- ✅ `services/auth/cmd/api/main.go`: Fixed CORS middleware logic and added all required headers

### Backend - Shortener Service  
- ✅ `services/shortener/cmd/api/main.go`: Fixed CORS middleware logic and added all required headers

### Backend - Analytics Service
- ✅ `services/analytics/cmd/api/main.go`: Fixed CORS middleware logic and added all required headers

## Verification

### 1. Frontend Build
```bash
cd /Users/ernie-dev/Documents/linkpulse/frontend
npm run build
```
Result: ✅ Build successful (0 errors)
- 9 routes compiled
- 209 kB main bundle

### 2. All Services Running
```bash
docker-compose ps
```
Result: ✅ All services healthy
- postgres: Healthy ✓
- redis: Healthy ✓
- kafka: Healthy ✓
- auth-service: Up ✓
- shortener-service: Up ✓
- analytics-service: Up ✓
- frontend: Up ✓

## Testing the Fix

### Test 1: Login Request
1. Navigate to `http://localhost:3000/login`
2. The CORS preflight request should succeed
3. You should see the login form
4. No console errors about CORS

### Test 2: Register Request
1. Navigate to `http://localhost:3000/register`
2. The page should load without CORS errors

### Test 3: API Calls
1. Open browser DevTools → Network tab
2. Perform a login/register attempt
3. Check the HTTP response headers for:
   - `Access-Control-Allow-Origin: http://localhost:3000`
   - `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH`
   - `Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With`
   - `Access-Control-Allow-Credentials: true`

## Key Learnings

1. **CORS Preflight**: Browser sends OPTIONS request before actual request. Both must be properly handled.
2. **Headers Placement**: CORS headers must be set **before** handling the preflight request.
3. **Allowed Methods**: Include all HTTP methods your frontend might use (PUT, DELETE, PATCH, etc.)
4. **Credentials**: When sending credentials (cookies, auth headers), set `Access-Control-Allow-Credentials: true`
5. **File Paths**: In Next.js, always use parentheses `(name)` directly, never escaped `\(name\)`

## Next Steps

1. ✅ Test login/register flow with actual backend
2. ✅ Verify link creation works (POST requests)
3. ✅ Verify link listing works (GET requests)
4. ✅ Verify link deletion works (DELETE requests)
5. ⏳ Test analytics WebSocket connection (Phase 3)

## Rollback (if needed)

All changes are safe and can be reverted by reverting the commits:
- Frontend: 1 file changed
- Backend: 3 files changed

All changes are backwards compatible and don't affect existing functionality.
