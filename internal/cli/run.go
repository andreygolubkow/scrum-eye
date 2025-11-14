package cli

import (
	"context"
	"fmt"
	collector "scrum-eye/internal/collector"
	"scrum-eye/internal/config"
	"scrum-eye/internal/sources/azureboards"
)

func Run(args []string) error {
	ctx := context.Background()

	teamName, customPath, err := parseArgs(args)
	if err != nil {
		printUsage()
		return err
	}
	if teamName == "" {
		printUsage()
		return fmt.Errorf("team name is required")
	}

	paths, err := resolveConfigPaths(teamName, customPath)
	if err != nil {
		return err
	}

	if err := ensureDirExists(paths.RootDir,
		fmt.Sprintf("Папка с конфигами (%s) не найдена. Создать её?", paths.RootDir)); err != nil {
		return err
	}

	if err := ensureGlobalConfig(paths.GlobalPath); err != nil {
		return err
	}

	if err := ensureTeamConfig(paths.TeamsDir, paths.TeamFile, paths.TeamName); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("✅ Конфигурация готова к использованию:")
	fmt.Println("  Root:   ", paths.RootDir)
	fmt.Println("  Global: ", paths.GlobalPath)
	fmt.Println("  Team:   ", paths.TeamFile)

	cfg, err := config.Load(paths.GlobalPath, paths.TeamsDir, paths.TeamName)
	if err != nil {
		return err
	}

	boardsClient := azureboards.NewClient(cfg.Team.AzureDevOps)

	collector := collector.NewCollector(boardsClient)

	snapshot, err := collector.Collect(ctx)

	fmt.Println(snapshot)
	fmt.Println(err)

	defer ctx.Done()
	return nil
}
