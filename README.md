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

### Building from Source

Prerequisites:
- Go 1.21 or higher
- Make
- Java 17+ (for testing)

```bash
# Clone the repository
git clone https://github.com/jgm01/minecrap_hoster.git
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

### Runtime

- Java 17 or higher
- Fabric server jar file
- Modern web browser with SSE support
