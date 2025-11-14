package cli

import (
	"fmt"
	"os"
)

// writeGlobalTemplate создаёт минимальный шаблон global.yaml.
func writeGlobalTemplate(path string) error {
	content := `# Глобальная конфигурация для scrum-eye
azure:
  organization: "https://dev.azure.com/your-org"

auth:
  azurePat: "CHANGE_ME_AZURE_PAT"
  teamcityToken: "CHANGE_ME_TEAMCITY_TOKEN"

teamcity:
  baseUrl: "https://teamcity.example.com"

storage:
  path: "./data"

defaults:
  branch: "develop"
  maxBuilds: 20
  sprintMode: "current"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("не удалось создать global.yaml: %w", err)
	}
	return nil
}

// writeTeamTemplate создаёт минимальный шаблон конфигурации для одной скрам-команды.
func writeTeamTemplate(path, teamName string) error {
	content := fmt.Sprintf(`# Конфигурация скрам-команды "%s"
name: "%s"
description: "Опиши команду"

azure:
  project: "YourProjectName"
  team: "%s"
  board:
    iterationPath: "YourProject\\%s"
    areaPath: "YourProject\\%s"
  repos:
    - name: "your-repo-name"
      defaultBranch: "develop"

teamcity:
  buildConfigs:
    - id: "Your_TeamCity_BuildConfig_Id"

metrics:
  maxBuilds: 20
  defaultBranch: "develop"
  wipLimit: 10
  wipPerPerson: 3
  overloadStoryPoints: 20

diff:
  baselineDays: 1
`, teamName, teamName, teamName, teamName, teamName)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("не удалось создать конфиг команды %s: %w", teamName, err)
	}
	return nil
}
