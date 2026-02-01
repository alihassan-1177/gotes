 #!/usr/bin/bash

# --- Configuration ---
BINARY_NAME="gotes"
INSTALL_PATH="$HOME/go/bin"
BUILD_FLAGS="-ldflags=-s" # Makes the binary smaller

# --- Colors for output ---
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔨 Starting build for ${BINARY_NAME}...${NC}"

# 1. Clean up old binary
if [ -f "$BINARY_NAME" ]; then
    echo "Cleaning up old binary..."
    rm "$BINARY_NAME"
fi

# 2. Tidy up Go modules
echo "Syncing dependencies..."
go mod tidy

# 3. Build the binary
echo -e "Compiling..."
if go build $BUILD_FLAGS -o "bin/$BINARY_NAME" main.go; then
    echo -e "${GREEN}✅ Build successful!${NC}"
else
    echo -e "${RED}❌ Build failed! Check errors above.${NC}"
    exit 1
fi

