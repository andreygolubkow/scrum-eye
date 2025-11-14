package common

import (
	"fmt"
	"os"
	"time"
)

// Config is the root configuration for dev-digest, loaded from YAML.
// All tokens can be provided either inline or via environment variables
// referenced as ${ENV_VAR}. Empty or missing sections will disable the module.
//
// Example available in: sample.config.yaml

type Config struct {
	Console  ConsoleConfig   `yaml:"console"`
	TeamCity *TeamCityConfig `yaml:"teamcity"`
	Azure    *AzureConfig    `yaml:"azure"`
}

type ConsoleConfig struct {
	Color   *bool         `yaml:"color"`   // default: auto (enabled when TTY)
	Timeout time.Duration `yaml:"timeout"` // default per-module timeout (e.g., 20s)
}

type TeamCityConfig struct {
	BaseURL string        `yaml:"base_url"`
	Token   string        `yaml:"token"`
	Branch  string        `yaml:"branch"`
	Builds  []string      `yaml:"builds"` // Build configuration IDs
	Timeout time.Duration `yaml:"timeout"`
}

type AzureConfig struct {
	Organization string             `yaml:"organization"` // e.g., https://dev.azure.com/{org}
	PAT          string             `yaml:"pat"`
	Boards       *AzureBoardsConfig `yaml:"boards"`
	Repos        *AzureReposConfig  `yaml:"repos"`
	Timeout      time.Duration      `yaml:"timeout"`
}

type AzureBoardsConfig struct {
	Project string `yaml:"project"`
	Team    string `yaml:"team"`
}

type AzureReposConfig struct{}

// ResolveEnv expands ${VAR} references in string fields
func (c *Config) ResolveEnv() {
	if c.TeamCity != nil {
		c.TeamCity.Token = expandEnv(c.TeamCity.Token)
	}
	if c.Azure != nil {
		c.Azure.PAT = expandEnv(c.Azure.PAT)
	}
}

func expandEnv(s string) string {
	if s == "" {
		return s
	}
	// Simple ${VAR} replacement
	out := ""
	for i := 0; i < len(s); {
		if s[i] == '$' && i+1 < len(s) && s[i+1] == '{' {
			j := i + 2
			for j < len(s) && s[j] != '}' {
				j++
			}
			if j < len(s) {
				key := s[i+2 : j]
				out += os.Getenv(key)
				i = j + 1
				continue
			}
		}
		out += string(s[i])
		i++
	}
	return out
}

// DefaultPerModuleTimeout returns a safe default if config isn't set.
func (c *Config) DefaultPerModuleTimeout() time.Duration {
	if c.Console.Timeout > 0 {
		return c.Console.Timeout
	}
	return 20 * time.Second
}

// Validate performs minimal validation
func (c *Config) Validate() error {
	if c.TeamCity != nil {
		if c.TeamCity.BaseURL == "" {
			return fmt.Errorf("teamcity.base_url is required when teamcity section is present")
		}
	}
	if c.Azure != nil {
		if c.Azure.Organization == "" {
			return fmt.Errorf("azure.organization is required when azure section is present")
		}
	}
	return nil
}
