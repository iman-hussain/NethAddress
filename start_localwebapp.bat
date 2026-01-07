@echo off
echo.
echo ========================================
echo   AddressIQ Local Development Setup
echo ========================================
echo.

REM Step 1: Check if .env exists
echo [1/3] Checking for .env file...
if not exist .env (
    copy .env.example .env
    echo       Copied .env.example to .env
) else (
    echo       .env file already exists.
)
echo.

REM Step 2: Check Docker and start services if available
echo [2/3] Checking Docker status...
docker info >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo       Docker is not running. Skipping Redis/PostgreSQL.
    echo       Backend will run without caching.
) else (
    echo       Docker is running. Starting services...
    docker-compose -f docker-compose.local.yml stop backend >nul 2>&1
    docker-compose -f docker-compose.local.yml up -d db cache
    echo       Waiting for services to be ready...
    timeout /t 5 /nobreak >nul
    echo       Docker services started.
)
echo.

REM Step 2.5: Start Backend Locally with localhost overrides
echo [2.5/3] Starting Backend locally...
REM Clear REDIS_URL if Docker not running to avoid connection attempts
if %ERRORLEVEL% NEQ 0 (
    start "AddressIQ Backend" cmd /k "cd backend && backend.exe"
) else (
    start "AddressIQ Backend" cmd /k "set DATABASE_URL=postgres://addressiq_user:addressiq_password@localhost:5432/addressiq_db?sslmode=disable && set REDIS_URL=redis://localhost:6379 && cd backend && backend.exe"
)
echo       Backend started in new window.
echo.

REM Step 3: Start frontend server
echo [3/3] Checking for existing process on port 3000...
for /f "tokens=5" %%a in ('netstat -aon ^| find ":3000" ^| find "LISTENING"') do (
    echo       Killing process %%a on port 3000...
    taskkill /f /pid %%a >nul 2>&1
)

echo       Starting frontend server on port 3000...
start cmd /k "cd frontend && python -m http.server 3000"
echo       Frontend server started in new window.
echo.

REM Step 4: Open browser
echo [4/4] Opening http://localhost:3000 in your browser...
start http://localhost:3000
echo       Browser opened. Check the "AddressIQ Backend" window for logs.
echo.