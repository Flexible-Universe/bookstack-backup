package bookstack

import (
	"os"

	"github.com/Flexible-Universe/bookstack-backup/internal/bookstack"
	"gopkg.in/yaml.v2"
)

// LoadConfig reads and parses a YAML config file.
func LoadConfig(path string) (bookstack.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return bookstack.Config{}, err
	}
	defer file.Close()

	var cfg bookstack.Config
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return bookstack.Config{}, err
	}
	return cfg, nil
}
