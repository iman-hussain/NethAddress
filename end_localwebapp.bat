@echo off

echo.
echo ========================================
echo   NethAddress Local Development - STOP
echo ========================================
echo.

REM Colors (using Windows color codes)
echo [1/2] Stopping Docker services...

docker-compose -f docker-compose.local.yml down

echo.
echo [2/2] Stopping frontend server...

REM Kill Python HTTP server processes
taskkill /f /im python.exe >nul 2>&1

echo.
echo ========================================
echo   NethAddress stopped successfully!
echo ========================================
echo.
echo All services have been stopped.
echo Run start_localwebapp.bat to start again.
echo.

pause