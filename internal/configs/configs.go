package configs

// APIConfig represents the configuration for the API server.
type APIConfig struct{}

// DockerdConfig represents the configuration for the Docker Daemon.
type DockerdConfig struct{}

// FileMDConfig represents the configuration for the File Management Daemon.
type FileMDConfig struct{}

// Config represents the configuration for the application.
type Config struct {
	API     *APIConfig     `koanf:"api"`
	Dockerd *DockerdConfig `koanf:"dockerd"`
	FileMD  *FileMDConfig  `koanf:"filemd"`
}

// LoadConfig loads the configuration for the application.
func LoadConfig() (*Config, error) {
	return Default(), nil
}
