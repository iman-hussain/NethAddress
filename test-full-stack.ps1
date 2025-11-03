#!/usr/bin/env pwsh
# Comprehensive full-stack test

Write-Host "=== AddressIQ Full Stack Test ===" -ForegroundColor Cyan
Write-Host ""

# Check if .env exists and has API URLs
Write-Host "1. Checking backend .env configuration..." -ForegroundColor Yellow
$envPath = "backend/.env"
if (Test-Path $envPath) {
    $envContent = Get-Content $envPath -Raw
    $hasBAG = $envContent -match "BAG_API_URL=https"
    $hasWeather = $envContent -match "KNMI_WEATHER_API_URL=https"
    $hasSolar = $envContent -match "KNMI_SOLAR_API_URL=https"
    $hasAir = $envContent -match "LUCHTMEETNET_API_URL=https"
    
    Write-Host "  BAG API URL: $(if($hasBAG){'✓'}else{'✗'})" -ForegroundColor $(if($hasBAG){'Green'}else{'Red'})
    Write-Host "  Weather API URL: $(if($hasWeather){'✓'}else{'✗'})" -ForegroundColor $(if($hasWeather){'Green'}else{'Red'})
    Write-Host "  Solar API URL: $(if($hasSolar){'✓'}else{'✗'})" -ForegroundColor $(if($hasSolar){'Green'}else{'Red'})
    Write-Host "  Air Quality API URL: $(if($hasAir){'✓'}else{'✗'})" -ForegroundColor $(if($hasAir){'Green'}else{'Red'})
    
    if (!$hasBAG -or !$hasWeather -or !$hasSolar -or !$hasAir) {
        Write-Host "  ⚠ Missing API URLs in .env - copying from production template" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "  Required .env entries:" -ForegroundColor Yellow
        Write-Host "  BAG_API_URL=https://api.bag.kadaster.nl/lvbag/individuelebevragingen/v2/adressenuitgebreid" -ForegroundColor Gray
        Write-Host "  KNMI_WEATHER_API_URL=https://api.open-meteo.com/v1" -ForegroundColor Gray
        Write-Host "  KNMI_SOLAR_API_URL=https://api.open-meteo.com/v1" -ForegroundColor Gray
        Write-Host "  LUCHTMEETNET_API_URL=https://api.luchtmeetnet.nl/open_api/stations" -ForegroundColor Gray
    }
} else {
    Write-Host "  ✗ .env file not found!" -ForegroundColor Red
}
Write-Host ""

# Check Go dependencies
Write-Host "2. Checking Go dependencies..." -ForegroundColor Yellow
Push-Location backend
$goModCheck = go mod verify 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ Go modules verified" -ForegroundColor Green
} else {
    Write-Host "  ✗ Go module issues: $goModCheck" -ForegroundColor Red
}
Pop-Location
Write-Host ""

# Run backend tests
Write-Host "3. Running backend tests..." -ForegroundColor Yellow
Push-Location backend
$testOutput = go test ./... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ All backend tests passed" -ForegroundColor Green
} else {
    Write-Host "  ✗ Backend tests failed:" -ForegroundColor Red
    Write-Host $testOutput
}
Pop-Location
Write-Host ""

# Build backend
Write-Host "4. Building backend..." -ForegroundColor Yellow
Push-Location backend
go build -o ../test-backend.exe . 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0 -and (Test-Path ../test-backend.exe)) {
    Write-Host "  ✓ Backend built successfully" -ForegroundColor Green
    
    # Start backend in background
    Write-Host ""
    Write-Host "5. Starting backend server..." -ForegroundColor Yellow
    $env:PORT = "8082"
    $backendProcess = Start-Process -FilePath "..\test-backend.exe" -PassThru -WindowStyle Hidden -RedirectStandardOutput "../backend-test.log" -RedirectStandardError "../backend-test-error.log"
    Start-Sleep -Seconds 3
    
    if (!$backendProcess.HasExited) {
        Write-Host "  ✓ Backend server started (PID: $($backendProcess.Id))" -ForegroundColor Green
        
        # Test health endpoint
        Write-Host ""
        Write-Host "6. Testing backend endpoints..." -ForegroundColor Yellow
        try {
            $health = Invoke-RestMethod -Uri "http://localhost:8082/healthz" -TimeoutSec 5
            Write-Host "  ✓ Health check passed" -ForegroundColor Green
            
            # Test search endpoint
            $search = Invoke-RestMethod -Uri "http://localhost:8082/search?address=1012LG+1" -TimeoutSec 10
            Write-Host "  ✓ Search endpoint responding" -ForegroundColor Green
            
            # Parse response to count API results
            if ($search -match 'data-response=''([^'']+)''') {
                $jsonData = $matches[1] -replace '&quot;', '"'
                $data = $jsonData | ConvertFrom-Json
                $totalAPIs = $data.apiResults.Count
                $successAPIs = ($data.apiResults | Where-Object { $_.status -eq 'success' }).Count
                $errorAPIs = ($data.apiResults | Where-Object { $_.status -eq 'error' }).Count
                $notConfiguredAPIs = ($data.apiResults | Where-Object { $_.status -eq 'not_configured' }).Count
                
                Write-Host "  API Results: $totalAPIs total" -ForegroundColor Cyan
                Write-Host "    ✓ Success: $successAPIs" -ForegroundColor Green
                Write-Host "    ✗ Error: $errorAPIs" -ForegroundColor Red
                Write-Host "    ⚠ Not Configured: $notConfiguredAPIs" -ForegroundColor Yellow
                
                # Show which APIs succeeded
                $successNames = ($data.apiResults | Where-Object { $_.status -eq 'success' }).name
                if ($successNames) {
                    Write-Host "  Working APIs:" -ForegroundColor Green
                    $successNames | ForEach-Object { Write-Host "    • $_" -ForegroundColor Green }
                }
            }
            
        } catch {
            Write-Host "  ✗ Backend test failed: $_" -ForegroundColor Red
        }
        
        # Clean up
        Write-Host ""
        Write-Host "7. Cleaning up..." -ForegroundColor Yellow
        Stop-Process -Id $backendProcess.Id -Force
        Start-Sleep -Seconds 1
        Write-Host "  ✓ Backend stopped" -ForegroundColor Green
        
    } else {
        Write-Host "  ✗ Backend failed to start" -ForegroundColor Red
        if (Test-Path "../backend-test-error.log") {
            Write-Host "  Error log:" -ForegroundColor Red
            Get-Content "../backend-test-error.log"
        }
    }
    
    # Clean up test binary
    if (Test-Path ../test-backend.exe) {
        Remove-Item ../test-backend.exe -Force
    }
} else {
    Write-Host "  ✗ Backend build failed" -ForegroundColor Red
}
Pop-Location

# Check frontend build info
Write-Host ""
Write-Host "8. Checking frontend build info..." -ForegroundColor Yellow
$indexPath = "frontend/index.html"
if (Test-Path $indexPath) {
    $indexContent = Get-Content $indexPath -Raw
    if ($indexContent -match 'meta name="build-commit" content="([^"]+)"') {
        Write-Host "  Build commit: $($matches[1])" -ForegroundColor Cyan
    }
    if ($indexContent -match 'meta name="build-date" content="([^"]+)"') {
        Write-Host "  Build date: $($matches[1])" -ForegroundColor Cyan
    }
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
