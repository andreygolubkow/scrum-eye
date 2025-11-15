package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ensureDirExists проверяет, что директория существует,
// и при необходимости предлагает её создать.
func ensureDirExists(dirPath string, prompt string) (created bool, err error) {
	info, err := os.Stat(dirPath)
	if err == nil {
		if info.IsDir() {
			return false, nil
		}
		return false, fmt.Errorf("%s существует, но это не директория", dirPath)
	}

	if !os.IsNotExist(err) {
		return false, err
	}

	// директории нет
	if !askYesNo(prompt, false) {
		return false, fmt.Errorf("директория %s не существует и не была создана", dirPath)
	}

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return false, fmt.Errorf("не удалось создать директорию %s: %w", dirPath, err)
	}

	fmt.Println("Создана директория:", dirPath)
	return true, nil
}

// ensureGlobalConfig проверяет наличие global.yaml,
// и при отсутствии предлагает создать шаблон.
func ensureGlobalConfig(globalPath string) error {
	_, err := os.Stat(globalPath)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	msg := fmt.Sprintf("Файл global.yaml (%s) не найден. Создать шаблонный файл?", globalPath)
	if !askYesNo(msg, false) {
		return fmt.Errorf("global.yaml не существует и не был создан")
	}

	if err := writeGlobalTemplate(globalPath); err != nil {
		return err
	}

	fmt.Println("Создан шаблон global.yaml:", globalPath)
	return nil
}

// ensureTeamConfig проверяет наличие config-файла команды,
// и при отсутствии предлагает создать шаблон.
func ensureTeamConfig(teamsDir, teamFile, teamName string) (created bool, err error) {
	dirCreated, err := ensureDirExists(
		teamsDir,
		fmt.Sprintf("Папка с командами (%s) не найдена. Создать её?", teamsDir))

	if err != nil {
		return false, err
	}

	_, err = os.Stat(teamFile)
	if err == nil {
		return dirCreated, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}

	msg := fmt.Sprintf("Файл конфигурации команды (%s) не найден. Создать шаблон для команды '%s'?",
		teamFile, teamName)
	if !askYesNo(msg, false) {
		return false, fmt.Errorf("конфиг команды %s не существует и не был создан", teamName)
	}

	if err := writeTeamTemplate(teamFile, teamName); err != nil {
		return false, err
	}

	fmt.Println("Создан шаблон конфигурации команды:", teamFile)
	return true, nil
}

// askYesNo задаёт вопрос пользователю и возвращает true/false.
// defaultYes=false значит, что Enter без ввода = "нет".
func askYesNo(question string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		defStr := "y/N"
		if defaultYes {
			defStr = "Y/n"
		}

		fmt.Printf("%s [%s]: ", question, defStr)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))

		if text == "" {
			return defaultYes
		}
		if text == "y" || text == "yes" || text == "д" || text == "да" {
			return true
		}
		if text == "n" || text == "no" || text == "н" || text == "нет" {
			return false
		}

		fmt.Println("Пожалуйста, ответь 'y' или 'n'.")
	}
}
