#!/bin/bash
# Demo script to showcase jif features

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

clear

echo -e "${CYAN}"
cat << "EOF"
     _ ___ _____ 
    | |_ _|  ___|
 _  | || || |_   
| |_| || ||  _|  
 \___/|___|_|    
                 
EOF
echo -e "${NC}"

echo -e "${GREEN}JIF - Terminal GIF Viewer Demo${NC}"
echo ""

# Check if jif exists
if [ ! -f "./jif" ]; then
    echo -e "${YELLOW}Building jif...${NC}"
    go build -o bin/jif ./cmd/jif
    echo ""
fi

# Generate test data if needed
if [ ! -d "testdata" ] || [ ! -f "testdata/simple.gif" ]; then
    echo -e "${YELLOW}Generating test GIF files...${NC}"
    go run testdata/generate_test_gifs.go
    echo ""
fi

echo -e "${CYAN}=== Demo 1: Simple 2-Frame Animation ===${NC}"
echo "Press 'q' to continue to next demo..."
echo ""
sleep 2
./bin/jif testdata/simple.gif

clear
echo -e "${CYAN}=== Demo 2: Multi-Frame Animation ===${NC}"
echo "Try these controls:"
echo "  - Space: Pause/Resume"
echo "  - ←/→: Navigate frames manually"
echo "  - ?: Show help"
echo ""
echo "Press 'q' to continue to next demo..."
echo ""
sleep 3
./bin/jif testdata/multi.gif

clear
echo -e "${CYAN}=== Demo 3: Fast Animation ===${NC}"
echo "This GIF has very short frame delays (50ms)"
echo ""
echo "Press 'q' to continue..."
echo ""
sleep 2
./bin/jif testdata/fast.gif

clear
echo -e "${CYAN}=== Demo 4: GIF Disposal Methods ===${NC}"
echo "This GIF tests different disposal methods"
echo "ensuring accurate frame composition"
echo ""
echo "Press 'q' to finish demo..."
echo ""
sleep 2
./bin/jif testdata/disposal.gif

clear
echo -e "${GREEN}"
cat << "EOF"
┌──────────────────────────────────────┐
│     Demo Complete!                   │
│                                      │
│  You can now use jif with your own  │
│  GIF files:                          │
│                                      │
│    ./bin/jif path/to/your.gif            │
│    ./bin/jif https://example.com/a.gif   │
│                                      │
│  Press ? while viewing for help      │
└──────────────────────────────────────┘
EOF
echo -e "${NC}"
echo ""
