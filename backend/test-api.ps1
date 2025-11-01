# AddressIQ API Test Script
Write-Host "=================================="
Write-Host "AddressIQ API Endpoint Tests"
Write-Host "=================================="
Write-Host ""

$baseUrl = "http://localhost:8080"
$testPostcode = "3541ED"
$testHouseNumber = "53"

# Test 1: Health Check
Write-Host "1. Testing Health Check..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/healthz" -Method Get
    Write-Host "✅ Health Check: $($response.status)" -ForegroundColor Green
} catch {
    Write-Host "❌ Health Check Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 2: Root Endpoint (API Info)
Write-Host "2. Testing Root Endpoint..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/" -Method Get
    Write-Host "✅ Root Endpoint: $($response.service) v$($response.version)" -ForegroundColor Green
    Write-Host "   Available endpoints: $($response.endpoints.Count)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Root Endpoint Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 3: Legacy Search
Write-Host "3. Testing Legacy Search..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/search?address=$testPostcode+$testHouseNumber" -Method Get
    Write-Host "✅ Legacy Search: $($response.address)" -ForegroundColor Green
    Write-Host "   Coordinates: [$($response.coordinates[0]), $($response.coordinates[1])]" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Legacy Search Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 4: Property Data
Write-Host "4. Testing Property Data..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property?postcode=$testPostcode&houseNumber=$testHouseNumber" -Method Get
    Write-Host "✅ Property Data: $($response.property.address)" -ForegroundColor Green
    Write-Host "   Data sources: $($response.property.dataSources -join ', ')" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Property Data Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 5: Property Scores
Write-Host "5. Testing Property Scores..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property/scores?postcode=$testPostcode&houseNumber=$testHouseNumber" -Method Get
    Write-Host "✅ Property Scores:" -ForegroundColor Green
    Write-Host "   ESG Score: $($response.scores.esgScore)" -ForegroundColor Cyan
    Write-Host "   Profit Score: $($response.scores.profitScore)" -ForegroundColor Cyan
    Write-Host "   Opportunity Score: $($response.scores.opportunityScore)" -ForegroundColor Cyan
    Write-Host "   Overall Score: $($response.scores.overallScore)" -ForegroundColor Cyan
    Write-Host "   Risk Level: $($response.scores.riskLevel)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Property Scores Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 6: Recommendations
Write-Host "6. Testing Recommendations..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property/recommendations?postcode=$testPostcode&houseNumber=$testHouseNumber" -Method Get
    Write-Host "✅ Recommendations: $($response.recommendations.Count) found" -ForegroundColor Green
    foreach ($rec in $response.recommendations) {
        Write-Host "   - $rec" -ForegroundColor Cyan
    }
} catch {
    Write-Host "❌ Recommendations Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 7: Full Analysis
Write-Host "7. Testing Full Analysis..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/property/analysis?postcode=$testPostcode&houseNumber=$testHouseNumber" -Method Get
    Write-Host "✅ Full Analysis:" -ForegroundColor Green
    Write-Host "   Address: $($response.property.address)" -ForegroundColor Cyan
    Write-Host "   Overall Score: $($response.scores.overallScore)" -ForegroundColor Cyan
    Write-Host "   Recommendations: $($response.scores.recommendations.Count)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Full Analysis Failed: $_" -ForegroundColor Red
}
Write-Host ""

Write-Host "=================================="
Write-Host "All Tests Complete!"
Write-Host "=================================="
