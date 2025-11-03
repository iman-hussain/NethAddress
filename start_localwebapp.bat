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

REM Step 2: Run docker-compose
echo [2/3] Starting Docker services (PostgreSQL, Redis, Backend)...
docker-compose -f docker-compose.local.yml up --build -d
echo       Docker services started in background.
echo.

REM Step 3: Start frontend server
echo [3/3] Starting frontend server on port 3000...
start cmd /k "cd frontend && python -m http.server 3000"
echo       Frontend server started in new window.
echo.

REM Step 4: Open browser
echo [4/4] Opening http://localhost:3000 in your browser...
start http://localhost:3000
echo       Browser opened. Tailing backend logs below (Ctrl+C to stop).
echo.
docker-compose -f docker-compose.local.yml logs -f backend