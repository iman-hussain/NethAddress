#!/bin/bash

echo ""
echo "========================================"
echo "  NethAddress Local Development - STOP"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Stop Docker services
echo -e "${YELLOW}[1/2] Stopping Docker services...${NC}"
docker-compose -f docker-compose.local.yml down
echo -e "${GREEN}      Docker services stopped.${NC}"
echo ""

# Step 2: Stop frontend server
echo -e "${YELLOW}[2/2] Stopping frontend server...${NC}"
# Kill Python HTTP server processes
pkill -f "python3 -m http.server 3000" 2>/dev/null || true
echo -e "${GREEN}      Frontend server stopped.${NC}"
echo ""

echo "========================================"
echo "  NethAddress stopped successfully!"
echo "========================================"
echo ""
echo "All services have been stopped."
echo "Run ./start_localwebapp.sh to start again."
echo ""