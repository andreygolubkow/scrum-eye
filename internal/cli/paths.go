package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

type ConfigPaths struct {
	RootDir    string
	GlobalPath string
	TeamsDir   string
	TeamFile   string
	TeamName   string
}

func resolveConfigPaths(teamName, customPath string) (ConfigPaths, error) {
	var root string

	if customPath != "" {
		abs, err := filepath.Abs(customPath)
		if err != nil {
			return ConfigPaths{}, fmt.Errorf("не удалось получить абсолютный путь для %s: %w", customPath, err)
		}
		root = abs
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return ConfigPaths{}, fmt.Errorf("не удалось получить домашнюю директорию: %w", err)
		}
		root = filepath.Join(home, ".scrum-eye")
	}

	globalPath := filepath.Join(root, "global.yaml")
	teamsDir := filepath.Join(root, "teams")
	teamFile := filepath.Join(teamsDir, teamName+".yaml")

	return ConfigPaths{
		RootDir:    root,
		GlobalPath: globalPath,
		TeamsDir:   teamsDir,
		TeamFile:   teamFile,
		TeamName:   teamName,
	}, nil
}
