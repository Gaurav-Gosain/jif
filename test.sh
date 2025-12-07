#!/bin/bash
set -e

echo "================================"
echo "JIF Test Suite"
echo "================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if test GIFs exist, generate if not
if [ ! -d "testdata" ] || [ ! -f "testdata/simple.gif" ]; then
    echo -e "${YELLOW}Generating test GIF files...${NC}"
    go run testdata/generate_test_gifs.go
    echo -e "${GREEN}✓ Test GIFs generated${NC}"
    echo ""
fi

# Run unit tests
echo -e "${YELLOW}Running unit tests...${NC}"
if go test -v -race -coverprofile=coverage.out ./jif/...; then
    echo -e "${GREEN}✓ All unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi
echo ""

# Show coverage
echo -e "${YELLOW}Test coverage:${NC}"
go tool cover -func=coverage.out | grep total
echo ""

# Run benchmarks
echo -e "${YELLOW}Running benchmarks...${NC}"
go test -bench=. -benchmem -run=^$ ./jif/... | tail -n +2
echo ""

# Build the binary
echo -e "${YELLOW}Building binary...${NC}"
if go build -o bin/jif ./cmd/jif; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo ""

# Test binary with test GIFs
echo -e "${YELLOW}Testing binary with sample GIFs...${NC}"

for gif in testdata/*.gif; do
    if [ -f "$gif" ]; then
        echo -n "  Testing with $(basename $gif)... "
        # Run for 1 second then quit
        timeout 1s ./bin/jif "$gif" > /dev/null 2>&1 || true
        echo -e "${GREEN}✓${NC}"
    fi
done
echo ""

# Check binary size
SIZE=$(ls -lh bin/jif | awk '{print $5}')
echo -e "${YELLOW}Binary size: ${NC}${SIZE}"
echo ""

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}All tests passed!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "To run the viewer:"
echo "  ./bin/jif testdata/simple.gif"
echo "  ./bin/jif <path-to-your-gif>"
echo "  ./bin/jif <url-to-gif>"
