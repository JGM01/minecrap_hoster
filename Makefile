# Binary names
BINARY_NAME=minecrap_hoster
WINDOWS_BINARY=$(BINARY_NAME).exe

# Build directories
BUILD_DIR=build
STATIC_DIR=static

# Main source file
MAIN_GO=cmd/server/main.go

# Version information
VERSION=1.0.0
BUILD_TIME=$(shell date +%FT%T%z)

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

.PHONY: all clean windows linux mac build-all

all: clean build-all

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Create build directory and copy static files
prepare:
	mkdir -p $(BUILD_DIR)
	cp -r $(STATIC_DIR) $(BUILD_DIR)/

# Build for Windows
windows: prepare
	mkdir -p $(BUILD_DIR)/windows_amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/windows_amd64/$(WINDOWS_BINARY) $(MAIN_GO)
	cp README.md $(BUILD_DIR)/windows_amd64/
	cd $(BUILD_DIR) && zip -r $(BINARY_NAME)_windows_amd64.zip windows_amd64

# Build for Linux
linux: prepare
	mkdir -p $(BUILD_DIR)/linux_amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/linux_amd64/$(BINARY_NAME) $(MAIN_GO)
	cp README.md $(BUILD_DIR)/linux_amd64/
	cd $(BUILD_DIR) && tar czf $(BINARY_NAME)_linux_amd64.tar.gz linux_amd64

# Build for macOS
mac: prepare
	mkdir -p $(BUILD_DIR)/darwin_amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/darwin_amd64/$(BINARY_NAME) $(MAIN_GO)
	cp README.md $(BUILD_DIR)/darwin_amd64/
	cd $(BUILD_DIR) && tar czf $(BINARY_NAME)_darwin_amd64.tar.gz darwin_amd64

# Build for all platforms
build-all: windows linux mac
