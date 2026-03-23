package configs

// Default returns the default configuration for the application.
func Default() *Config {
	return &Config{
		API:     &APIConfig{},
		Dockerd: &DockerdConfig{},
		FileMD:  &FileMDConfig{},
	}
}
