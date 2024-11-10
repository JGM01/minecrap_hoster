# Minecraft Server Manager

A simple web-based Minecraft server manager for Fabric servers.

## Quick Start

1. Place the executable in a directory with your Minecraft server files
2. Make sure you have Java installed
3. Run the executable:

```bash
# Linux/Mac
./minecrap_hoster

# Windows
minecrap_hoster.exe
```

4. Open your web browser to `http://localhost:8080`

## Command Line Options

- `-port`: HTTP server port (default "8080")
- `-java`: Path to Java executable (default "java")
- `-server`: Path to server jar (default "fabric-server-mc.1.20.1-loader.0.16.5-launcher.1.0.1.jar")
- `-memory`: Memory allocation in MB (default 8192)
- `-max-logs`: Maximum number of log lines to keep (default 1000)
- `-g1gc`: Use G1 Garbage Collector (default true)

Example:
```bash
./minecrap_hoster -port 8081 -memory 16384
```

## Directory Structure

The program expects:
```
├── minecrap_hoster (or minecrap_hoster.exe)
├── fabric-server-mc.1.20.1-loader.0.16.5-launcher.1.0.1.jar
└── static/
    └── index.html
```

## Requirements

- Java 17 or higher
- Fabric server jar file
