package report

import (
	"fmt"
	"scrum-eye/internal/domain"
	"strings"
	"time"
)

func PrintCurrentSprint(project *domain.Project) {
	if project == nil || project.CurrentSprint == nil {
		fmt.Println("❌ No sprint information available")
		return
	}

	sprint := project.CurrentSprint
	width := 60
	line := strings.Repeat("─", width)

	now := time.Now()

	startDateStr := "N/A"
	if sprint.StartDate != nil {
		startDateStr = sprint.StartDate.Format("2006-01-02")
	}

	endDateStr := "N/A"
	daysLeftStr := "N/A"
	if sprint.EndDate != nil {
		endDateStr = sprint.EndDate.Format("2006-01-02")
		daysLeft := int(sprint.EndDate.Sub(now).Hours() / 24)
		daysLeftStr = fmt.Sprintf("%d", daysLeft)
	}

	// Подсчёт по типам
	typeCounts := map[domain.WorkItemType]int{}
	for _, wi := range sprint.WorkItems {
		typeCounts[wi.Type]++
	}

	// Формируем строку сводки по типам
	summaryParts := make([]string, 0)
	total := len(sprint.WorkItems)
	if total > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("Total: %d", total))
	}
	for _, t := range []domain.WorkItemType{
		domain.WorkItemStory,
		domain.WorkItemBug,
		domain.WorkItemTask,
		domain.WorkItemEpic,
		domain.WorkItemFeature,
		domain.WorkItemUnknown,
	} {
		if count, ok := typeCounts[t]; ok && count > 0 {
			summaryParts = append(summaryParts, fmt.Sprintf("%s: %d", t, count))
		}
	}
	summaryLine := "No work items"
	if len(summaryParts) > 0 {
		summaryLine = strings.Join(summaryParts, ", ")
	}

	fmt.Printf("\n┌%s┐\n", line)
	fmt.Printf("│ %-*s│\n", width, " 🏃 Current Sprint")
	fmt.Printf("├%s┤\n", line)
	fmt.Printf("│ %-*s│\n", width, fmt.Sprintf("   Name: %s", sprint.Name))
	fmt.Printf("│ %-*s│\n", width, fmt.Sprintf("   Start Date: %s", startDateStr))
	fmt.Printf("│ %-*s│\n", width, fmt.Sprintf("   End Date: %s", endDateStr))
	fmt.Printf("│ %-*s│\n", width, fmt.Sprintf("   Days Left: %s", daysLeftStr))
	fmt.Printf("│ %-*s│\n", width, fmt.Sprintf("   Work Items: %s", summaryLine))

	// Если нет задач — закрываем блок
	if len(sprint.WorkItems) == 0 {
		fmt.Printf("└%s┘\n\n", line)
		return
	}

	// Таблица с задачами
	fmt.Printf("├%s┤\n", line)
	fmt.Printf("│ %-*s│\n", width, "   Work Items List:")
	fmt.Printf("│ %-*s│\n", width, "   ID    Type       Name")
	fmt.Printf("│ %-*s│\n", width, "   ----  ---------- ---------------------------------")

	for _, wi := range sprint.WorkItems {
		idStr := fmt.Sprintf("%d", wi.ID)
		typeStr := string(wi.Type)
		// Оставляем место под отступы/ID/тип и немного под границу
		nameWidth := width - len("   ") - 4 /*ID*/ - 2 /*spaces*/ - 10 /*Type*/ - 3
		name := truncate(wi.Name, nameWidth)
		lineStr := fmt.Sprintf("   %-4s %-10s %s", idStr, typeStr, name)
		fmt.Printf("│ %-*s│\n", width, lineStr)
	}

	fmt.Printf("└%s┘\n\n", line)
}

// truncate обрезает строку до max символов и добавляет многоточие при необходимости.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}
