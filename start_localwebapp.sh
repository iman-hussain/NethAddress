#!/bin/bash

echo ""
echo "========================================"
echo "  AddressIQ Local Development Setup"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Check if .env exists
echo -e "${YELLOW}[1/5] Checking for .env file...${NC}"
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}      Copied .env.example to .env${NC}"
else
    echo -e "${GREEN}      .env file already exists.${NC}"
fi
echo ""

# Step 2: Build frontend with dynamic build info
echo -e "${YELLOW}[2/5] Building frontend with current commit info...${NC}"
if command -v pwsh &> /dev/null; then
    pwsh -ExecutionPolicy Bypass -File build-frontend.ps1
elif command -v powershell &> /dev/null; then
    powershell -ExecutionPolicy Bypass -File build-frontend.ps1
else
    echo -e "${YELLOW}      PowerShell not found, updating build info manually...${NC}"
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILDDATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    sed -i.bak "s/name=\"build-commit\" content=\"[^\"]*\"/name=\"build-commit\" content=\"$COMMIT\"/" frontend/index.html
    sed -i.bak "s/name=\"build-date\" content=\"[^\"]*\"/name=\"build-date\" content=\"$BUILDDATE\"/" frontend/index.html
    rm -f frontend/index.html.bak
fi
echo -e "${GREEN}      Frontend build info updated.${NC}"
echo ""

# Step 3: Run docker-compose
echo -e "${YELLOW}[3/5] Starting Docker services (PostgreSQL, Redis, Backend)...${NC}"
docker-compose -f docker-compose.local.yml up --build -d
echo -e "${GREEN}      Docker services started in background.${NC}"
echo ""

# Step 4: Start frontend server
echo -e "${YELLOW}[4/5] Starting frontend server on port 3000...${NC}"
cd frontend && python3 -m http.server 3000 &
echo -e "${GREEN}      Frontend server started in background.${NC}"
echo ""

# Step 5: Open browser and tail logs
echo -e "${YELLOW}[5/5] Opening http://localhost:3000 in your browser...${NC}"
xdg-open http://localhost:3000 &
echo -e "${GREEN}      Browser opened. Tailing backend logs below (Ctrl+C to stop).${NC}"
echo ""
docker-compose -f docker-compose.local.yml logs -f backend