#!/bin/bash
set -e

echo ""
echo "========================================"
echo "  NethAddress Local Development Setup"
echo "========================================"
echo ""

# Colours
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Colour

# Step 1: Check if .env exists
echo -e "${YELLOW}[1/5] Checking for .env file...${NC}"
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}      Copied .env.example to .env${NC}"
else
    echo -e "${GREEN}      .env file already exists.${NC}"
fi
echo ""

# Step 2: Compile SCSS
echo -e "${YELLOW}[2/5] Compiling SCSS...${NC}"
if ! command -v sass &> /dev/null; then
    echo -e "${YELLOW}      Sass not found. Checking for npm...${NC}"
    if ! command -v npm &> /dev/null; then
        echo -e "${RED}      ERROR: npm not found. Please install Node.js from https://nodejs.org/${NC}"
        exit 1
    fi
    echo -e "${YELLOW}      Installing sass globally via npm...${NC}"
    npm install -g sass
fi
echo "      Running: sass frontend/static/scss/main.scss frontend/static/css/styles.css"
sass frontend/static/scss/main.scss frontend/static/css/styles.css --style=expanded
echo -e "${GREEN}      SCSS compiled successfully.${NC}"
echo ""

# Step 3: Run docker-compose
echo -e "${YELLOW}[3/5] Starting Docker services (PostgreSQL, Redis)...${NC}"
if ! docker info &> /dev/null; then
    echo -e "${YELLOW}      Docker is not running. Skipping Redis/PostgreSQL.${NC}"
    DOCKER_AVAILABLE=0
else
    docker-compose -f docker-compose.local.yml stop backend 2>/dev/null || true
    docker-compose -f docker-compose.local.yml up -d db cache
    echo -e "${GREEN}      Docker services started in background.${NC}"
    DOCKER_AVAILABLE=1
fi
echo ""

# Step 4: Start Backend Locally
echo -e "${YELLOW}[4/5] Starting Backend locally...${NC}"
export POSTGRES_DB=nethaddress_db
export POSTGRES_USER=nethaddress_user
export POSTGRES_PASSWORD=nethaddress_password
export DATABASE_URL=postgres://nethaddress_user:nethaddress_password@localhost:5432/nethaddress_db?sslmode=disable
export REDIS_URL=redis://localhost:6379

# Build backend if binary doesn't exist
if [ ! -f backend/backend ]; then
    echo -e "${YELLOW}      Building backend binary...${NC}"
    cd backend && go build -o backend main.go && cd ..
fi
./backend/backend &
BACKEND_PID=$!
echo -e "${GREEN}      Backend started (PID: $BACKEND_PID).${NC}"
echo ""

# Step 5: Start frontend server
echo -e "${YELLOW}[5/5] Starting frontend server...${NC}"
# Kill any process listening on port 3000
if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${YELLOW}      Killing existing process on port 3000...${NC}"
    lsof -Pi :3000 -sTCP:LISTEN -t | xargs kill -9 2>/dev/null || true
fi

cd frontend && python3 -m http.server 3000 &
FRONTEND_PID=$!
cd ..
echo -e "${GREEN}      Frontend server started (PID: $FRONTEND_PID).${NC}"
echo ""

# Open browser
echo -e "${YELLOW}Opening http://localhost:3000 in your browser...${NC}"
sleep 2
if command -v xdg-open &> /dev/null; then
    xdg-open http://localhost:3000 &
elif command -v open &> /dev/null; then
    open http://localhost:3000 &
fi

echo ""
echo "========================================"
echo -e "${GREEN}  NethAddress is now running!${NC}"
echo "========================================"
echo "  Frontend: http://localhost:3000"
echo "  Backend:  http://localhost:8080"
echo ""
echo "  Press Ctrl+C to stop all services."
echo "========================================"
echo ""

# Wait for user interrupt
trap "echo ''; echo 'Shutting down...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" INT TERM
wait