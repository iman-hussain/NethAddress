@echo off
setlocal enabledelayedexpansion

REM Change to the directory where this script lives
cd /d "%~dp0"

echo.
echo ========================================
echo   NethAddress Local Development Setup
echo ========================================
echo   Working directory: %CD%
echo ========================================
echo.

REM Step 1: Check if .env exists
echo [1/5] Checking for .env file...
if not exist .env (
    if exist .env.example (
        copy .env.example .env >nul
        echo       Copied .env.example to .env
    ) else (
        echo       WARNING: No .env.example found. Continuing without .env
    )
) else (
    echo       .env file already exists.
)
echo.

REM Step 2: Compile SCSS to CSS
echo [2/5] Compiling SCSS...
where sass >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo       Sass not found. Checking for npm...
    where npm >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo       ERROR: npm not found. Please install Node.js from https://nodejs.org/
        goto :error
    )
    echo       Installing sass globally via npm...
    call npm install -g sass
    if %ERRORLEVEL% NEQ 0 (
        echo       ERROR: Failed to install sass.
        goto :error
    )
)
echo       Running: sass frontend/static/scss/main.scss frontend/static/css/styles.css
call sass frontend/static/scss/main.scss frontend/static/css/styles.css --style=expanded
if %ERRORLEVEL% NEQ 0 (
    echo       ERROR: SCSS compilation failed. Check syntax errors above.
    goto :error
)
echo       SCSS compiled successfully.
echo.
REM Step 3: Check Docker and start services if available
echo [3/5] Checking Docker status...
set DOCKER_AVAILABLE=0
docker info >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo       Docker is running. Starting services...
    docker-compose -f docker-compose.local.yml stop backend >nul 2>&1
    docker-compose -f docker-compose.local.yml up -d db cache
    echo       Waiting for services to be ready...
    timeout /t 5 /nobreak >nul
    echo       Docker services started.
    set DOCKER_AVAILABLE=1
) else (
    echo       Docker is not running. Skipping Redis/PostgreSQL.
    echo       Backend will run without caching.
)
echo.

REM Step 4: Start Backend Locally
echo [4/5] Starting Backend locally...
if not exist backend\backend.exe (
    echo       ERROR: backend\backend.exe not found.
    echo       Please build the backend first: cd backend ^&^& go build -o backend.exe main.go
    goto :error
)
if "%DOCKER_AVAILABLE%"=="1" (
    start "NethAddress Backend" cmd /k "cd /d %CD%\backend && set DATABASE_URL=postgres://nethaddress_user:nethaddress_password@localhost:5432/nethaddress_db?sslmode=disable && set REDIS_URL=redis://localhost:6379 && backend.exe"
) else (
    start "NethAddress Backend" cmd /k "cd /d %CD%\backend && backend.exe"
)
echo       Backend started in new window.
echo.

REM Step 5: Start frontend server
echo [5/5] Starting frontend server...
for /f "tokens=5" %%a in ('netstat -aon ^| find ":3000" ^| find "LISTENING" 2^>nul') do (
    echo       Killing existing process %%a on port 3000...
    taskkill /f /pid %%a >nul 2>&1
)

echo       Starting frontend server on port 3000...
start "NethAddress Frontend" cmd /k "cd /d %CD%\frontend && python -m http.server 3000"
echo       Frontend server started in new window.
echo.

REM Open browser
echo Opening http://localhost:3000 in your browser...
timeout /t 2 /nobreak >nul
start http://localhost:3000
echo.
echo ========================================
echo   NethAddress is now running!
echo ========================================
echo   Frontend: http://localhost:3000
echo   Backend:  http://localhost:8080
echo.
echo   Check the backend window for API logs.
echo   Press any key to close this window...
echo ========================================
pause >nul
goto :eof

:error
echo.
echo ========================================
echo   ERROR: Setup failed. See message above.
echo ========================================
echo   Press any key to close this window...
pause >nul
exit /b 1