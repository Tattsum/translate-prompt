package config

import (
	"os"
	"strconv"
	"strings"
)

// Server holds runtime configuration loaded from environment variables.
type Server struct {
	ListenHost         string
	Port               int
	AllowedOrigins     []string
	InvestigateEnabled bool
}

// LoadServerFromEnv reads server configuration. Defaults preserve local development behavior.
func LoadServerFromEnv(defaultPort int) Server {
	listenHost := os.Getenv("LISTEN_HOST")
	if listenHost == "" {
		listenHost = "127.0.0.1"
	}

	port := defaultPort
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			port = p
		}
	}

	allowedOrigins := parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	investigateEnabled := true
	if v := os.Getenv("INVESTIGATE_ENABLED"); v != "" {
		investigateEnabled = strings.EqualFold(v, "true") || v == "1"
	}

	return Server{
		ListenHost:         listenHost,
		Port:               port,
		AllowedOrigins:     allowedOrigins,
		InvestigateEnabled: investigateEnabled,
	}
}

func parseAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if o := strings.TrimSpace(part); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}
