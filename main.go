package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort    = 8080
	defaultHost    = "0.0.0.0"
	appName        = "ds2api"
	appVersion     = "dev"
	defaultDSNPort = 27016 // Using 27016 as my servers run on this port instead of 2302

	// defaultReadTimeout and defaultWriteTimeout prevent slow-client attacks.
	// Bumped ReadTimeout to 15s and WriteTimeout to 30s to accommodate slower
	// DayZ server query responses on my home lab setup.
	defaultReadTimeout  = 15 * time.Second
	defaultWriteTimeout = 30 * time.Second

	// defaultIdleTimeout closes idle keep-alive connections after 60s.
	// Added this after noticing lingering connections on my home lab setup.
	defaultIdleTimeout = 60 * time.Second
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Host        string
	Port        int
	DSNAddress  string
	DSNPort     int
	APIKey      string
	Debug       bool
}

// loadConfig reads configuration from environment variables with sensible defaults.
func loadConfig() (*Config, error) {
	port := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT value %q: %w", p, err)
		}
		port = parsed
	}

	dsnPort := defaultDSNPort
	if p := os.Getenv("DSN_PORT"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid DSN_PORT value %q: %w", p, err)
		}
		dsnPort = parsed
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}

	dsnAddress := os.Getenv("DSN_ADDRESS")
	if dsnAddress == "" {
		return nil, fmt.Errorf("DSN_ADDRESS environment variable is required")
	}

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	return &Config{
		Host:       host,
		Port:       port,
		DSNAddress: dsnAddress,
		DSNPort:    dsnPort,
		APIKey:     os.Getenv("API_KEY"),
		Debug:      debug,
	}, nil
}

func main() {
	// Attempt to load .env file; ignore error if file doesn't exist (e.g., in production).
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("[WARN] Could not load .env file: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("[FATAL] Configuration error: %v", err)
	}

	if cfg.Debug {
		log.Printf("[DEBUG] %s %s starting in debug mode", appName, appVersion)
		log.Printf("[DEBUG] DSN target: %s:%d", cfg.DSNAddress, cfg.DSNPort)
		log.Printf("[DEBUG] API key set: %v", cfg.APIKey != "")
	}

	router := setupRouter(cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("[INFO] %s %s listening on %s", appName, appVersion, addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Server error: %v", err)
	}
}
