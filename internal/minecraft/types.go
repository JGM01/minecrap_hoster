package minecraft

import (
	"io"
	"os/exec"
	"sync"
)

const (
	Stopped  uint8 = 0
	Starting uint8 = 1
	Running  uint8 = 2
	Stopping uint8 = 3
)

// ServerConfig holds all server configuration parameters
type ServerConfig struct {
	JavaPath            string
	ExecutablePath      string
	MemoryUtilizationMB int
	MaxLogLines         int
	UseG1GC             bool // Whether to use G1 Garbage Collector
	ServerFlag          bool // Whether to use -server flag
}

type MinecraftServer struct {
	Command *exec.Cmd
	Status  uint8
	LogBuf  []string
	stdin   io.WriteCloser
	mutex   sync.RWMutex
	config  ServerConfig

	autoRestart bool
}

// InitConfig returns default server configuration
func InitConfig() ServerConfig {
	return ServerConfig{
		JavaPath:            "java",
		ExecutablePath:      "fabric-server-mc.1.20.1-loader.0.16.5-launcher.1.0.1.jar",
		MemoryUtilizationMB: 8192, // 8GB
		MaxLogLines:         1000,
		UseG1GC:             true,
		ServerFlag:          true,
	}
}
