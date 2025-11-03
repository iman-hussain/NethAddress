# AddressIQ Frontend Build Script
# Updates build commit hash and date in frontend/index.html

param(
    [string]$HtmlFile = "frontend/index.html"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  AddressIQ Frontend Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get git commit hash (short version)
try {
    $commit = git rev-parse --short HEAD
    if ($LASTEXITCODE -ne 0) { throw "Git command failed" }
} catch {
    Write-Warning "Could not get git commit hash, using 'unknown'"
    $commit = "unknown"
}

# Get current date/time in ISO format
$buildDate = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"

Write-Host "Build Info:" -ForegroundColor Green
Write-Host "  Commit: $commit" -ForegroundColor Green
Write-Host "  Date: $buildDate" -ForegroundColor Green
Write-Host ""

# Read the HTML file
$content = Get-Content $HtmlFile -Raw

# Replace build-commit meta tag
$content = $content -replace '(?<=name="build-commit" content=")[^"]*', $commit

# Replace build-date meta tag
$content = $content -replace '(?<=name="build-date" content=")[^"]*', $buildDate

# Write back to file
$content | Set-Content $HtmlFile -Encoding UTF8

Write-Host "✅ Frontend build info updated successfully!" -ForegroundColor Green
Write-Host ""