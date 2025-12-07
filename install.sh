#!/bin/bash
# JIF Installation Script
# Installs the latest release of jif for your platform

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REPO="Gaurav-Gosain/jif"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="jif"

echo -e "${CYAN}"
cat <<"EOF"
     _ ___ _____ 
    | |_ _|  ___|
 _  | || || |_   
| |_| || ||  _|  
 \___/|___|_|    
                 
EOF
echo -e "${NC}"

echo -e "${GREEN}JIF Installation Script${NC}"
echo ""

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "${OS}" in
Linux*) OS_TYPE=Linux ;;
Darwin*) OS_TYPE=Darwin ;;
CYGWIN* | MINGW* | MSYS*) OS_TYPE=Windows ;;
*)
	echo -e "${RED}Unsupported OS: ${OS}${NC}"
	exit 1
	;;
esac

case "${ARCH}" in
x86_64 | amd64) ARCH_TYPE=x86_64 ;;
i386 | i686) ARCH_TYPE=i386 ;;
arm64 | aarch64) ARCH_TYPE=arm64 ;;
armv7l) ARCH_TYPE=armv7 ;;
armv6l) ARCH_TYPE=armv6 ;;
*)
	echo -e "${RED}Unsupported architecture: ${ARCH}${NC}"
	exit 1
	;;
esac

echo -e "${YELLOW}Detected platform:${NC} ${OS_TYPE}_${ARCH_TYPE}"
echo ""

# Get latest release info
echo -e "${YELLOW}Fetching latest release...${NC}"
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest")

if [ -z "$LATEST_RELEASE" ]; then
	echo -e "${RED}Failed to fetch release information${NC}"
	exit 1
fi

VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
echo -e "${GREEN}Latest version: ${VERSION}${NC}"
echo ""

# Construct download URL
ARCHIVE_NAME="jif_${VERSION#v}_${OS_TYPE}_${ARCH_TYPE}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

echo -e "${YELLOW}Downloading ${ARCHIVE_NAME}...${NC}"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

# Download and extract
if ! curl -L -o "${TMP_DIR}/${ARCHIVE_NAME}" "${DOWNLOAD_URL}"; then
	echo -e "${RED}Failed to download binary${NC}"
	echo -e "${YELLOW}URL: ${DOWNLOAD_URL}${NC}"
	exit 1
fi

echo -e "${YELLOW}Extracting archive...${NC}"
tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "${TMP_DIR}"

# Create install directory if it doesn't exist
mkdir -p "${INSTALL_DIR}"

# Install binary
echo -e "${YELLOW}Installing to ${INSTALL_DIR}...${NC}"
mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
	echo ""
	echo -e "${YELLOW}Warning: ${INSTALL_DIR} is not in your PATH${NC}"
	echo -e "${YELLOW}Add the following line to your ~/.bashrc or ~/.zshrc:${NC}"
	echo -e "${CYAN}export PATH=\"${INSTALL_DIR}:\$PATH\"${NC}"
	echo ""
fi

echo -e "${GREEN}JIF ${VERSION} installed successfully!${NC}"
echo ""
echo -e "${CYAN}Usage:${NC}"
echo -e "  ${BINARY_NAME} animation.gif"
echo -e "  ${BINARY_NAME} https://example.com/animation.gif"
echo ""
echo -e "${CYAN}For help:${NC}"
echo -e "  ${BINARY_NAME} (press ? while viewing)"
echo ""
