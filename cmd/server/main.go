package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"minecrap_hoster/internal/handlers"
	"minecrap_hoster/internal/minecraft"
)

// Command-line flags
var (
	port          = flag.String("port", "8080", "HTTP server port")
	java_path     = flag.String("java", "java", "Path to Java executable")
	jar_path      = flag.String("jar", "fabric-server-mc.1.20.1-loader.0.16.5-launcher.1.0.1.jar", "Path to server jar")
	memory_mb     = flag.Int("memory", 8192, "Memory allocation in MB")
	max_log_lines = flag.Int("max-logs", 1000, "Maximum number of log lines to keep")
	use_g1gc      = flag.Bool("g1gc", true, "Use G1 Garbage Collector")
	jvm_server    = flag.Bool("jvm-server", true, "Use -server JVM flag")
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Set up logging with timestamps
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	// Initialize and validate configuration
	config, err := buildConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create server instance
	server := minecraft.NewServer(config)

	// Create and configure HTTP handler
	handler := handlers.NewHandler(server)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Start HTTP server
	server_addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting server on http://localhost%s", server_addr)

	if err := startServer(server_addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// buildConfig creates and validates the server configuration.
func buildConfig() (minecraft.ServerConfig, error) {
	config := minecraft.ServerConfig{
		JavaPath:            *java_path,
		ExecutablePath:      *jar_path, // Changed from server_path to jar_path
		MemoryUtilizationMB: *memory_mb,
		MaxLogLines:         *max_log_lines,
		UseG1GC:             *use_g1gc,
		ServerFlag:          *jvm_server, // Changed from server_flag to jvm_server
	}

	// Ensure paths exist and are accessible
	if err := validatePaths(&config); err != nil {
		return config, err
	}

	// Log configuration
	log.Printf("Configuration:")
	log.Printf("  Java Path: %s", config.JavaPath)
	log.Printf("  Server Jar: %s", config.ExecutablePath)
	log.Printf("  Memory: %d MB", config.MemoryUtilizationMB)
	log.Printf("  Max Log Lines: %d", config.MaxLogLines)
	log.Printf("  Use G1GC: %v", config.UseG1GC)
	log.Printf("  JVM Server Flag: %v", config.ServerFlag)

	return config, nil
}

// validatePaths ensures required files exist and are accessible.
func validatePaths(config *minecraft.ServerConfig) error {
	// Check Java executable
	java_path, err := exec.LookPath(config.JavaPath)
	if err != nil {
		return fmt.Errorf("java executable not found: %v", err)
	}
	config.JavaPath = java_path

	// Check server.jar
	server_path := filepath.Clean(config.ExecutablePath)
	if _, err := os.Stat(server_path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("server.jar not found at %s", server_path)
		}
		return fmt.Errorf("error accessing server.jar: %v", err)
	}
	config.ExecutablePath = server_path

	return nil
}

// startServer starts the HTTP server with the given configuration.
func startServer(addr string, handler http.Handler) error {
	// Change from ":8080" to "0.0.0.0:8080" to listen on all interfaces
	server := &http.Server{
		Addr:           "0.0.0.0" + addr, // Changed this line
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Create static directory if it doesn't exist
	if err := os.MkdirAll("static", 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %v", err)
	}

	log.Printf("Server accessible at:")
	log.Printf("  Local: http://localhost%s", addr)
	log.Printf("  Network: http://<server-ip>%s", addr)

	return server.ListenAndServe()
}
