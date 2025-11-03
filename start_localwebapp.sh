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
echo -e "${YELLOW}[1/3] Checking for .env file...${NC}"
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}      Copied .env.example to .env${NC}"
else
    echo -e "${GREEN}      .env file already exists.${NC}"
fi
echo ""

# Step 2: Run docker-compose
echo -e "${YELLOW}[2/3] Starting Docker services (PostgreSQL, Redis, Backend)...${NC}"
docker-compose -f docker-compose.local.yml up --build -d
echo -e "${GREEN}      Docker services started in background.${NC}"
echo ""

# Step 3: Start frontend server
echo -e "${YELLOW}[3/3] Starting frontend server on port 3000...${NC}"
cd frontend && python3 -m http.server 3000 &
echo -e "${GREEN}      Frontend server started in background.${NC}"
echo ""

# Step 4: Open browser and tail logs
echo -e "${YELLOW}[4/4] Opening http://localhost:3000 in your browser...${NC}"
xdg-open http://localhost:3000 &
echo -e "${GREEN}      Browser opened. Tailing backend logs below (Ctrl+C to stop).${NC}"
echo ""
docker-compose -f docker-compose.local.yml logs -f backend