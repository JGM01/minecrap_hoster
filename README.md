# Minecraft Server Manager

A modern web-based Minecraft server manager with real-time monitoring and control capabilities. Built specifically for Fabric servers, this tool provides an intuitive interface for managing your Minecraft server through a web browser.

## Features

- 🎮 Real-time server control (start/stop/restart)
- 📊 Live console output with SSE (Server-Sent Events)
- 🔄 Auto-restart capability on crashes
- ⌨️ Direct console command input
- 🚫 Force stop protection for data safety
- 📱 Responsive web interface

## Quick Start

### For Users

1. Download the latest release for your platform from the [Releases](https://github.com/yourusername/minecrap_hoster/releases) page
2. Place the executable in your Minecraft server directory
3. Ensure you have Java 17 or higher installed
4. Run the executable:

```bash
# Linux/macOS
./minecrap_hoster

# Windows
minecrap_hoster.exe
```

5. Open your web browser to `http://localhost:8080`

### Directory Structure

The program expects the following structure:
```
├── minecrap_hoster (or minecrap_hoster.exe)
├── fabric-server-mc.1.20.1-loader.0.16.5-launcher.1.0.1.jar
└── static/
    └── index.html
```

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | HTTP server port | 8080 |
| `-java` | Path to Java executable | "java" |
| `-jar` | Path to server jar | "fabric-server-mc.1.20.1-loader.0.16.5-launcher.1.0.1.jar" |
| `-memory` | Memory allocation in MB | 8192 |
| `-max-logs` | Maximum number of log lines to keep | 1000 |
| `-g1gc` | Use G1 Garbage Collector | true |
| `-jvm-server` | Use server JVM flag | true |

Example with custom settings:
```bash
./minecrap_hoster -port 8081 -memory 16384 -max-logs 2000
```

## For Developers

### Technology Stack

- Backend: Go 1.21+
- Frontend: HTML, JavaScript (with HTMX), Tailwind CSS
- Real-time: Server-Sent Events (SSE)
- Process Management: Native Go exec package

### Building from Source

Prerequisites:
- Go 1.21 or higher
- Make
- Java 17+ (for testing)

```bash
# Clone the repository
git clone https://github.com/yourusername/minecrap_hoster.git
cd minecrap_hoster

# Build for all platforms
make build-all

# Or build for specific platform
make windows  # Creates build/windows_amd64/minecrap_hoster.exe
make linux    # Creates build/linux_amd64/minecrap_hoster
make mac      # Creates build/darwin_amd64/minecrap_hoster
```

### Project Structure

```
├── cmd/
│   └── server/
│       └── main.go       # Application entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   └── minecraft/        # Minecraft server management
├── static/              # Static web files
│   └── index.html       # Web interface
├── Makefile            # Build configuration
└── go.mod             # Go module definition
```

### Key Components

- `minecraft.MinecraftServer`: Core server management
- `handlers.Handler`: HTTP request handling
- Server-Sent Events (SSE) for real-time updates
- HTMX for dynamic UI updates

### Runtime

- Java 17 or higher
- Fabric server jar file
- Minimum 512MB RAM (8GB recommended)
- Modern web browser with SSE support

## Security

The server manager listens on all interfaces by default. For production use, consider:
- Using a firewall
- Configuring reverse proxy with SSL/TLS
- Implementing authentication
