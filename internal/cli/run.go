package cli

import (
	"context"
	"fmt"
	"scrum-eye/internal/collector"
	"scrum-eye/internal/config"
	"scrum-eye/internal/report"
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

	err = ensureConfigurationExists(paths)
	if err != nil {
		return err
	}

	cfg, err := config.Load(paths.GlobalPath, paths.TeamsDir, paths.TeamName)
	if err != nil {
		return err
	}

	boardsClient := azureboards.NewClient(cfg.Team.AzureDevOps)

	dataCollector := collector.NewCollector(boardsClient)

	project, err := dataCollector.Collect(ctx)

	report.PrintCurrentSprint(project)

	defer ctx.Done()
	return nil
}

func ensureConfigurationExists(paths ConfigPaths) error {
	created, err := ensureDirExists(paths.RootDir,
		fmt.Sprintf("Папка с конфигами (%s) не найдена. Создать её?", paths.RootDir))
	if err != nil {
		return err
	}

	if err := ensureGlobalConfig(paths.GlobalPath); err != nil {
		return err
	}

	created, err = ensureTeamConfig(paths.TeamsDir, paths.TeamFile, paths.TeamName)
	if err != nil {
		return err
	}

	if created {
		fmt.Println()
		fmt.Println("✅ Конфигурация готова к использованию:")
		fmt.Println("  Root:   ", paths.RootDir)
		fmt.Println("  Global: ", paths.GlobalPath)
		fmt.Println("  Team:   ", paths.TeamFile)
	}

	return nil
}
