package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(globalPath, teamsDir, teamName string) (*AppConfig, error) {
	var g GlobalConfig
	if err := loadYAML(globalPath, &g); err != nil {
		return nil, fmt.Errorf("load global: %w", err)
	}

	teamPath := filepath.Join(teamsDir, teamName+".yaml")

	var t TeamConfig
	if err := loadYAML(teamPath, &t); err != nil {
		return nil, fmt.Errorf("load team %s: %w", teamName, err)
	}

	t = *merge(g, t)

	return &AppConfig{
		Global: g,
		Team:   t,
	}, nil
}

func loadYAML(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

func merge(global GlobalConfig, team TeamConfig) *TeamConfig {
	if team.AzureDevOps.Organisation == "" {
		team.AzureDevOps.Organisation = global.AzureDevOps.Organization
	}
	if team.AzureDevOps.Token == "" {
		team.AzureDevOps.Token = global.AzureDevOps.Token
	}
	return &team
}
