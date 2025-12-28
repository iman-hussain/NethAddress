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
echo -e "${YELLOW}[2/3] Starting Docker services (PostgreSQL, Redis)...${NC}"
docker-compose -f docker-compose.local.yml stop backend
docker-compose -f docker-compose.local.yml up -d db cache
echo -e "${GREEN}      Docker services started in background.${NC}"
echo ""

# Step 2.5: Start Backend Locally
echo -e "${YELLOW}[2.5/3] Starting Backend locally...${NC}"
export POSTGRES_DB=addressiq_db
export POSTGRES_USER=addressiq_user
export POSTGRES_PASSWORD=addressiq_password
export DATABASE_URL=postgres://addressiq_user:addressiq_password@localhost:5432/addressiq_db?sslmode=disable
export REDIS_URL=redis://localhost:6379
# Build backend to ensure binary exists
cd backend && go build -o backend main.go && cd ..
./backend/backend &
echo -e "${GREEN}      Backend started in background (using localhost for DB/Redis).${NC}"
echo ""

# Step 3: Start frontend server
echo -e "${YELLOW}[3/3] Checking for existing process on port 3000...${NC}"
# Kill any process listening on port 3000
if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null ; then
    echo -e "${YELLOW}      Killing existing process on port 3000...${NC}"
    lsof -Pi :3000 -sTCP:LISTEN -t | xargs kill -9
fi

echo -e "${YELLOW}      Starting frontend server on port 3000...${NC}"
cd frontend && python3 -m http.server 3000 &
echo -e "${GREEN}      Frontend server started in background.${NC}"
echo ""

# Step 4: Open browser and tail logs
echo -e "${YELLOW}[4/4] Opening http://localhost:3000 in your browser...${NC}"
xdg-open http://localhost:3000 &
echo -e "${GREEN}      Browser opened. Tailing backend logs below (Ctrl+C to stop).${NC}"
echo ""
docker-compose -f docker-compose.local.yml logs -f backend