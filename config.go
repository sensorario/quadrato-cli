package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".quadrato", "config.json")
}

func loadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{}
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveConfig(cfg Config) {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0700)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(path, data, 0600)
}

func mustToken() string {
	cfg := loadConfig()
	if cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "%s Non autenticato. Esegui: %s\n",
			red.Render("✗"), bold.Render("quadrato login"))
		os.Exit(1)
	}
	return cfg.Token
}
