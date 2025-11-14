package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func parseArgs(args []string) (teamName string, customPath string, err error) {
	for _, a := range args {
		if strings.HasPrefix(a, "--path=") {
			customPath = strings.TrimPrefix(a, "--path=")
			continue
		}

		// первый не-флаг — это имя команды
		if !strings.HasPrefix(a, "-") && teamName == "" {
			teamName = a
			continue
		}

		if a == "--path" || a == "-path" {
			return "", "", errors.New("формат --path без значения не поддерживается, используй --path=<путь>")
		}

		// всё остальное считаем лишними аргументами
		if !strings.HasPrefix(a, "--") {
			return "", "", fmt.Errorf("лишний аргумент: %s", a)
		}
	}

	return teamName, customPath, nil
}

func printUsage() {
	fmt.Println("Использование:")
	fmt.Println("  scrum-eye.exe <team-name> [--path=<путь к папке с конфигами>]")
	fmt.Println()
	fmt.Println("По умолчанию конфиги ищутся в:")
	fmt.Println("  $HOME/.scrum-eye/global.yaml")
	fmt.Println("  $HOME/.scrum-eye/teams/<team-name>.yaml")
	fmt.Println()
	fmt.Println("Примеры:")
	fmt.Println("  scrum-eye.exe my-team")
	fmt.Println("  scrum-eye.exe my-team --path=C:\\configs\\scrum-eye")

	// маленький бонус: если хочется подсказать HOME:
	home, err := os.UserHomeDir()
	if err == nil {
		fmt.Printf("\nТекущая домашняя директория: %s\n", home)
	}
}
