package handlers

import (
	"fmt"
	"log"
	"minecrap_hoster/internal/minecraft"
	"net/http"
	"os"
	"time"
)

// Represents the core HTTP request handler with associated Minecraft server instance
type Handler struct {
	server *minecraft.MinecraftServer
}

// Creates a new handler instance with server validation
func NewHandler(server *minecraft.MinecraftServer) *Handler {
	if server == nil {
		panic("Server must not be nil.")
	}
	log.Printf("Handler created with server instance")
	return &Handler{server: server}
}

type routeConfig struct {
	path    string
	handler http.HandlerFunc
	logMsg  string
}

// Registers all HTTP routes and their corresponding handlers
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	log.Printf("Registering routes...")

	// Static file handler
	mux.HandleFunc("/", h.logRequest(http.FileServer(http.Dir("static")).ServeHTTP, "Static file request"))

	// Define routes configuration
	routes := []routeConfig{
		{"/api/server/start", h.HandleStart, "Start endpoint"},
		{"/api/server/stop", h.HandleStop, "Stop endpoint"},
		{"/api/server/force-stop", h.HandleForceStop, "Force stop endpoint"},
		{"/api/server/status", h.HandleStatus, "Status endpoint"},
		{"/api/server/logs", h.HandleLogs, "Logs SSE endpoint"},
		{"/api/server/command", h.HandleCommand, "Command endpoint"},
		{"/api/server/restart", h.HandleRestart, "Restart endpoint"},
		{"/api/server/auto-restart", h.HandleToggleAutoRestart, "Auto-restart toggle endpoint"},
		{"/api/server/auto-restart/status", h.HandleGetAutoRestart, "Auto-restart status endpoint"},
		{"/api/hoster/shutdown", h.HandleShutdownHoster, "Hoster shutdown endpoint"},
	}

	// Register routes with logging middleware
	for _, route := range routes {
		mux.HandleFunc(route.path, h.logRequest(route.handler, route.logMsg))
	}

	log.Printf("All routes registered")
}

// Middleware function to log incoming requests
func (h *Handler) logRequest(next http.HandlerFunc, logMsg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s from %s", logMsg, r.URL.Path, r.RemoteAddr)
		next(w, r)
	}
}

// Processes and executes server commands
func (h *Handler) HandleCommand(w http.ResponseWriter, r *http.Request) {
	if err := validateCommandRequest(w, r); err != nil {
		return
	}

	command := r.PostForm.Get("command")
	if err := h.server.ExecuteCommand(command); err != nil {
		log.Printf("Failed to execute command: %v", err)
		http.Error(w, fmt.Sprintf("Failed to execute command: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func validateCommandRequest(w http.ResponseWriter, r *http.Request) error {
	if err := AssertMethodPost(w, r); err != nil {
		return err
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return err
	}

	if command := r.PostForm.Get("command"); command == "" {
		http.Error(w, "Command cannot be empty", http.StatusBadRequest)
		return fmt.Errorf("empty command")
	}

	return nil
}

// Initiates the Minecraft server start sequence
func (h *Handler) HandleStart(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodPost(w, r); err != nil {
		return
	}

	if err := h.server.Start(); err != nil {
		log.Printf("Failed to start server: %v", err)
		http.Error(w, fmt.Sprintf("Failed to start server: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithMessage(w, "Server starting...", http.StatusOK)
}

// Gracefully stops the Minecraft server
func (h *Handler) HandleStop(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodPost(w, r); err != nil {
		return
	}

	if err := h.server.Stop(); err != nil {
		log.Printf("Failed to stop server: %v", err)
		http.Error(w, fmt.Sprintf("Failed to stop server: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithMessage(w, "Server stopping", http.StatusOK)
}

// Forces an immediate server shutdown
func (h *Handler) HandleForceStop(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodPost(w, r); err != nil {
		return
	}

	if err := h.server.ForceStop(); err != nil {
		log.Printf("Failed to force-stop server: %v", err)
		http.Error(w, fmt.Sprintf("Failed to force-stop server: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithMessage(w, "Server force-stopped", http.StatusOK)
}

// Handles server restart requests
func (h *Handler) HandleRestart(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodPost(w, r); err != nil {
		return
	}

	if err := h.server.Restart(); err != nil {
		log.Printf("Failed to restart server: %v", err)
		http.Error(w, fmt.Sprintf("Failed to restart server: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithMessage(w, "Server restarting...", http.StatusOK)
}

// Toggles the auto-restart feature and returns the new state
func (h *Handler) HandleToggleAutoRestart(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodPost(w, r); err != nil {
		return
	}

	enabled := h.server.ToggleAutoRestart()
	respondWithJSON(w, map[string]bool{"enabled": enabled})
}

// Returns the current auto-restart state
func (h *Handler) HandleGetAutoRestart(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodGet(w, r); err != nil {
		return
	}

	enabled := h.server.GetAutoRestart()
	respondWithJSON(w, map[string]bool{"enabled": enabled})
}

// Initiates a graceful shutdown of both the Minecraft server and the hoster
func (h *Handler) HandleShutdownHoster(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodPost(w, r); err != nil {
		return
	}

	if err := h.initiateShutdown(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop server: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithMessage(w, "Initiating shutdown sequence...", http.StatusOK)
	go h.waitForShutdown()
}

func (h *Handler) initiateShutdown() error {
	if h.server.Status == minecraft.Running || h.server.Status == minecraft.Starting {
		return h.server.Stop()
	}
	return nil
}

func (h *Handler) waitForShutdown() {
	for h.server.Status != minecraft.Stopped {
		time.Sleep(500 * time.Millisecond)
	}
	log.Printf("Minecraft server stopped, shutting down hoster...")
	os.Exit(0)
}

// Returns the current server status as HTML
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if err := AssertMethodGet(w, r); err != nil {
		return
	}

	status := h.server.Status
	log.Printf("Current server status: %d", status)
	respondWithHTML(w, getStatusHTML(status))
}

// Converts server status to styled HTML representation
func getStatusHTML(status uint8) string {
	statusConfig := map[uint8]struct {
		class string
		text  string
	}{
		minecraft.Running:  {"green", "Running"},
		minecraft.Starting: {"blue", "Starting"},
		minecraft.Stopping: {"yellow", "Stopping"},
		minecraft.Stopped:  {"red", "Stopped"},
	}

	config, exists := statusConfig[status]
	if !exists {
		config = struct {
			class string
			text  string
		}{"gray", "Unknown"}
	}

	return fmt.Sprintf(`<span class="px-2 py-1 bg-%s-100 text-%s-800 rounded-full">%s</span>`,
		config.class, config.class, config.text)
}

// HTTP method validation helpers
func AssertMethodPost(w http.ResponseWriter, r *http.Request) error {
	return assertMethod(w, r, http.MethodPost)
}

func AssertMethodGet(w http.ResponseWriter, r *http.Request) error {
	return assertMethod(w, r, http.MethodGet)
}

func assertMethod(w http.ResponseWriter, r *http.Request, method string) error {
	if r.Method != method {
		http.Error(w, fmt.Sprintf("Only %s method allowed", method), http.StatusMethodNotAllowed)
		return fmt.Errorf("invalid method: %s", r.Method)
	}
	return nil
}

// Response helpers
func respondWithMessage(w http.ResponseWriter, message string, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, message)
}

func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"enabled":%v}`, data.(map[string]bool)["enabled"])
}

func respondWithHTML(w http.ResponseWriter, html string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
