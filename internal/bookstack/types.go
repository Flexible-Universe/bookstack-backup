package bookstack

// InstanceConfig holds configuration for a single BookStack instance.
type InstanceConfig struct {
	Name        string       `yaml:"name"`
	BaseURL     string       `yaml:"base_url"`
	TokenID     string       `yaml:"token_id"`
	TokenSecret string       `yaml:"token_secret"`
	BackupPath  string       `yaml:"backup_path"`
	Schedule    string       `yaml:"schedule"`
	Target      TargetConfig `yaml:"target"`
}

// TargetConfig defines whether to crawl a book or a shelve.
type TargetConfig struct {
	Type string `yaml:"type"` // "book" or "shelve"
	ID   string `yaml:"id"`
}

// Config holds the overall YAML configuration.
type Config struct {
	Instances []InstanceConfig `yaml:"instances"`
}
