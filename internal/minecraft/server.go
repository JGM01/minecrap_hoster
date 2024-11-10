package minecraft

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"
)

// Creates and validates a new MinecraftServer instance
func NewServer(config ServerConfig) *MinecraftServer {
	validateConfig(config)
	return &MinecraftServer{
		Status: Stopped,
		LogBuf: make([]string, 0, config.MaxLogLines),
		config: config,
	}
}

func validateConfig(config ServerConfig) {
	if config.JavaPath == "" {
		panic("Java PATH must be non-empty.")
	}
	if config.ExecutablePath == "" {
		panic("The server's executable path must be non-empty.")
	}
	if config.MemoryUtilizationMB <= 0 {
		panic("Memory utilization must be positive.")
	}
	if config.MaxLogLines <= 0 {
		panic("Maximum log lines must be positive.")
	}
}

// Constructs the Java command with appropriate arguments
func (s *MinecraftServer) buildJavaCommand() *exec.Cmd {
	args := buildJavaArgs(s.config)
	log.Printf("Building Java command: %s %v", s.config.JavaPath, args)
	return exec.Command(s.config.JavaPath, args...)
}

func buildJavaArgs(config ServerConfig) []string {
	args := make([]string, 0, 8)

	if config.ServerFlag {
		args = append(args, "-server")
	}

	args = append(args, fmt.Sprintf("-Xmx%dM", config.MemoryUtilizationMB))

	if config.UseG1GC {
		args = append(args, "-XX:+UseG1GC")
	}

	args = append(args, "-jar", config.ExecutablePath, "nogui")
	return args
}

// Initializes and starts the Minecraft server process
func (s *MinecraftServer) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.validateStartState(); err != nil {
		return err
	}

	if err := s.initializeProcess(); err != nil {
		return err
	}

	log.Printf("Server started successfully")
	return nil
}

func (s *MinecraftServer) validateStartState() error {
	log.Printf("Start requested. Current status: %d", s.Status)
	if s.Status != Stopped {
		return fmt.Errorf("cannot start server: current state is %d", s.Status)
	}
	return nil
}

func (s *MinecraftServer) initializeProcess() error {
	s.Command = s.buildJavaCommand()

	if err := s.setupPipes(); err != nil {
		return err
	}

	s.Status = Starting

	if err := s.Command.Start(); err != nil {
		s.handleStartError(err)
		return fmt.Errorf("failed to start process: %v", err)
	}

	s.startMonitoring()
	s.Status = Running
	return nil
}

func (s *MinecraftServer) setupPipes() error {
	var err error

	s.stdin, err = s.Command.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := s.Command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := s.Command.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	go s.handleLogs(stdout)
	go s.handleLogs(stderr)

	return nil
}

func (s *MinecraftServer) handleStartError(err error) {
	log.Printf("Failed to start process: %v", err)
	s.Status = Stopped
	s.stdin = nil
}

func (s *MinecraftServer) startMonitoring() {
	go s.monitorProcess()
}

// Monitors the server process and handles its termination
func (s *MinecraftServer) monitorProcess() {
	log.Printf("Process monitor started")
	err := s.Command.Wait()

	s.mutex.Lock()
	wasRunning := s.Status == Running
	s.updateServerState()
	autoRestart := s.autoRestart
	s.mutex.Unlock()

	s.handleProcessExit(err, wasRunning, autoRestart)
}

func (s *MinecraftServer) updateServerState() {
	s.Status = Stopped
	s.stdin = nil
	s.Command = nil
}

func (s *MinecraftServer) handleProcessExit(err error, wasRunning, autoRestart bool) {
	if err != nil {
		log.Printf("Process exited with error: %v", err)
	} else {
		log.Printf("Process exited normally")
	}

	if autoRestart && wasRunning {
		s.handleAutoRestart()
	}

	s.AddLog("Server process has stopped")
	log.Printf("Process monitor complete")
}

func (s *MinecraftServer) handleAutoRestart() {
	log.Printf("Auto-restart enabled, attempting restart...")
	go func() {
		time.Sleep(5 * time.Second)
		if err := s.Start(); err != nil {
			log.Printf("Auto-restart failed: %v", err)
		}
	}()
}

// Initiates graceful server shutdown
func (s *MinecraftServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.validateStopState(); err != nil {
		return err
	}

	if err := s.sendStopCommand(); err != nil {
		return err
	}

	s.Status = Stopping
	log.Printf("Stop command sent successfully")
	return nil
}

func (s *MinecraftServer) validateStopState() error {
	log.Printf("Stop requested. Current status: %d", s.Status)
	if s.Status != Running {
		return fmt.Errorf("cannot stop server: current state is %d", s.Status)
	}
	return nil
}

func (s *MinecraftServer) sendStopCommand() error {
	if s.stdin == nil {
		log.Printf("Error: stdin is nil")
		return fmt.Errorf("server stdin is not available")
	}

	log.Printf("Sending stop command...")
	_, err := fmt.Fprintln(s.stdin, "stop")
	if err != nil {
		log.Printf("Failed to write stop command: %v", err)
		return fmt.Errorf("failed to send stop command: %v", err)
	}
	return nil
}

// Forces immediate server termination
func (s *MinecraftServer) ForceStop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.validateForceStopState(); err != nil {
		return err
	}

	if err := s.Command.Process.Kill(); err != nil {
		log.Printf("Failed to kill process: %v", err)
		return fmt.Errorf("failed to kill process: %v", err)
	}

	s.Status = Stopping
	log.Printf("Force stop successful")
	return nil
}

func (s *MinecraftServer) validateForceStopState() error {
	log.Printf("Force stop requested. Current status: %d", s.Status)
	if s.Command == nil || s.Status == Stopped {
		return fmt.Errorf("server is not running")
	}
	return nil
}

// Performs a server restart operation
func (s *MinecraftServer) Restart() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.Status != Running {
		return fmt.Errorf("cannot restart server: current state is %d", s.Status)
	}

	if err := s.Stop(); err != nil {
		return fmt.Errorf("failed to stop server during restart: %v", err)
	}

	s.waitForStop()
	return s.Start()
}

func (s *MinecraftServer) waitForStop() {
	for s.Status != Stopped {
		s.mutex.Unlock()
		time.Sleep(100 * time.Millisecond)
		s.mutex.Lock()
	}
}

// Log management methods
func (s *MinecraftServer) handleLogs(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	scanner.Buffer(make([]byte, 1024*16), 1024*16)

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Server output: %s", line)
		s.addLogLine(line)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading logs: %v", err)
		s.addLogLine("Error reading logs: " + err.Error())
	}
}

func (s *MinecraftServer) addLogLine(line string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.LogBuf) >= s.config.MaxLogLines {
		s.LogBuf = s.LogBuf[1:]
	}
	s.LogBuf = append(s.LogBuf, line)
}

// Server control and status methods
func (s *MinecraftServer) ExecuteCommand(command string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.Status != Running {
		return fmt.Errorf("cannot execute command: server is not running")
	}

	if s.stdin == nil {
		return fmt.Errorf("server stdin is not available")
	}

	_, err := fmt.Fprintln(s.stdin, command)
	if err != nil {
		return fmt.Errorf("failed to send command: %v", err)
	}

	return nil
}

func (s *MinecraftServer) ToggleAutoRestart() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.autoRestart = !s.autoRestart
	return s.autoRestart
}

func (s *MinecraftServer) GetAutoRestart() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.autoRestart
}

// Public log access methods
func (s *MinecraftServer) AddLog(line string) {
	log.Printf("External log added: %s", line)
	s.addLogLine(line)
}

func (s *MinecraftServer) GetLogs() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	logs := make([]string, len(s.LogBuf))
	copy(logs, s.LogBuf)
	return logs
}

func (s *MinecraftServer) GetLogsSince(index int) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if index >= len(s.LogBuf) {
		return []string{}
	}

	logs := make([]string, len(s.LogBuf)-index)
	copy(logs, s.LogBuf[index:])
	return logs
}
