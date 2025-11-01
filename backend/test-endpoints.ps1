# Test script for AddressIQ API endpoints
$baseUrl = "http://localhost:8081"
$postcode = "3541ED"
$houseNumber = "53"

Write-Host "🧪 Testing AddressIQ API Endpoints" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Root endpoint
Write-Host "📍 Test 1: Root endpoint (GET /)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/" -Method Get
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Service: $($response.service)" -ForegroundColor Gray
    Write-Host "Version: $($response.version)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 2: Health check
Write-Host "📍 Test 2: Health check (GET /healthz)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/healthz" -Method Get
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Status: $($response.status)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 3: Legacy search endpoint
Write-Host "📍 Test 3: Legacy search (GET /search)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/search?postcode=$postcode&houseNumber=$houseNumber" -Method Get
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Address: $($response.bagData.address)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 4: Property data endpoint
Write-Host "📍 Test 4: Property data (GET /api/property)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property?postcode=$postcode&houseNumber=$houseNumber" -Method Get
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Address: $($response.property.address)" -ForegroundColor Gray
    Write-Host "Data Sources: $($response.property.dataSources -join ', ')" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 5: Property scores endpoint
Write-Host "📍 Test 5: Property scores (GET /api/property/scores)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property/scores?postcode=$postcode&houseNumber=$houseNumber" -Method Get
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "ESG Score: $($response.scores.esgScore)" -ForegroundColor Gray
    Write-Host "Profit Score: $($response.scores.profitScore)" -ForegroundColor Gray
    Write-Host "Opportunity Score: $($response.scores.opportunityScore)" -ForegroundColor Gray
    Write-Host "Overall Score: $($response.scores.overallScore)" -ForegroundColor Gray
    Write-Host "Risk Level: $($response.scores.riskLevel)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 6: Recommendations endpoint
Write-Host "📍 Test 6: Recommendations (GET /api/property/recommendations)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property/recommendations?postcode=$postcode&houseNumber=$houseNumber" -Method Get
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Recommendations count: $($response.recommendations.Count)" -ForegroundColor Gray
    foreach ($rec in $response.recommendations) {
        Write-Host "  • $rec" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 7: Full analysis endpoint
Write-Host "📍 Test 7: Full analysis (GET /api/property/analysis)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property/analysis?postcode=$postcode&houseNumber=$houseNumber" -Method Get
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Address: $($response.property.address)" -ForegroundColor Gray
    Write-Host "Overall Score: $($response.scores.overallScore)" -ForegroundColor Gray
    Write-Host "Data Sources: $($response.property.dataSources.Count)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed: $_" -ForegroundColor Red
}
Write-Host ""

Write-Host "=================================" -ForegroundColor Cyan
Write-Host "✅ All tests completed!" -ForegroundColor Green
